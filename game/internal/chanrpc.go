package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

var (
	agentWarning int32    = 5000
	agentIP      sync.Map // key:IP(string) value:num(int)
)

func init() {

	// 開始遊戲服務
	skeleton.RegisterChanRPC("StartServer", respondStart)

	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)

	// 玩家登入登出子遊戲
	skeleton.RegisterChanRPC("UserLogin", playerEnterGame)
	skeleton.RegisterChanRPC("UserLogout", playerExitGame)

	// 用戶輸贏錢
	skeleton.RegisterChanRPC("WinMoney", respondWinMoney)
	skeleton.RegisterChanRPC("LoseMoney", respondLoseMoney)
}

func rpcNewAgent(args []interface{}) {
	// log.Debug("<-------------新链接请求连接--------------->")
	a := args[0].(gate.Agent)
	// sameIPattack(a)
	ServerSurPool.AgentNum++
	if ServerSurPool.AgentNum > agentWarning {
		TgMsg := fmt.Sprintf("彩源猜大小服务agent: %v", ServerSurPool.AgentNum)
		common.SendTextToTelegramChat(common.TgbotChatID, TgMsg, common.TgbotToken)
		agentWarning += 100
	}
	common.Debug_log("新的客户端连接:[%v]%v 目前agent數量為:%v", a.RemoteAddr().Network(), a.RemoteAddr().String(), ServerSurPool.AgentNum)
	// p := &Player{}
	// p.Init()
	// p.ConnAgent = a
	// p.ConnAgent.SetUserData(p)
}

// // 預防相同IP攻擊(TODO)
// func sameIPattack(newAgent gate.Agent) {
// 	newIP := newAgent.RemoteAddr().String()
// 	connNum, ok := agentIP.Load(newIP)
// 	if ok {
// 		if connNum.(int) >= 5 {
// 			CloseAgent(newAgent) // 關閉對方連線
// 			//加入黑名單機制
// 		}
// 		connNum = connNum.(int) + 1
// 	} else {
// 		agentIP.Store(newIP, 0)
// 	}

// }

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	unusualLogout(a, "连接断开")
}

// 玩家進入子遊戲服務
func playerEnterGame(args []interface{}) {
	// common.Debug_log("gameModule playerEnterGame")
	cInfo := args[0].(common.UserInfo)
	client, ok := AgentFromuserID_.Load(cInfo.UserID)
	common.Debug_log("用户登陆返回,userID=%v\n", cInfo.UserID)
	if !ok {
		common.Debug_log("用户不在线,userID=%v\n", cInfo.UserID)
		return
	}
	//从中心服务器获取到的用户信息

	log.Debug("玩家正常登陆:%v", cInfo.UserID)
	login := &msg.Login_S2C{}
	login.PlayerInfo = new(msg.PlayerInfo)
	login.PlayerInfo.Id = common.Int32ToStr(cInfo.UserID)
	login.PlayerInfo.NickName = cInfo.UserName
	login.PlayerInfo.HeadImg = cInfo.UserHead
	login.PlayerInfo.Account = cInfo.Balance - cInfo.LockBalance

	u := &Player{}
	for _, v := range hall.roomList {
		if v != nil {
			if v.RoomId == "1" {
				login.PlayerNumR1 = v.PlayerLength()
				login.Room01 = v.IsOpenRoom
			} else if v.RoomId == "2" {
				login.PlayerNumR2 = v.PlayerLength()
				login.Room02 = v.IsOpenRoom
			}
			for _, v := range v.PlayerList {
				if v.Id == cInfo.UserID {
					u = v
				}
			}
		}
	}
	// log.Debug("Room01:%v,Room02:%v", login.Room01, login.Room02)

	if u.Id != cInfo.UserID {
		u.Id = cInfo.UserID
		u.HeadImg = cInfo.UserHead
		u.NickName = cInfo.UserName
		u.PackageId = cInfo.PackageID
		u.Account = cInfo.Balance - cInfo.LockBalance
		u.Init()
	}

	//游戏数据库中缓存的用户数据
	// lInfo, ok := allUser[cInfo.UserID]
	lInfo, ok := allUser_.Load(cInfo.UserID)

	if ok {
		//已有用戶更新資訊
		u.updateInfo(cInfo, lInfo)
	} else {
		//用户信息保存到数据库中(首次遊玩)
		u.InsertPlayerInfo()
		ServerSurPool.SumUser++
	}

	// u.Password = m.GetPassWord()
	// u.Token = m.GetToken()

	limitData := LoadUserLimitBet(u)
	minBet, _ := strconv.Atoi(limitData.MinBet)
	maxBet, _ := strconv.Atoi(limitData.MaxBet)
	u.MinBet = int32(minBet)
	u.MaxBet = int32(maxBet)

	hall.UserRecord.Store(u.Id, u)

	rId, _ := hall.UserRoom.Load(u.Id) // 玩家的房間
	v, _ := hall.RoomRecord.Load(rId)  // 房間是否存在
	if v != nil {
		// 玩家如果已在游戏中，则返回房间数据(TODO:斷線重連前端先回大廳?前端如果在房間傳進入房間)
		room := v.(*Room)
		for i, userId := range room.UserLeave {
			// log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
			// 把玩家从掉线列表中移除
			if userId == u.Id {
				room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
				log.Debug("AllocateUser 清除玩家记录~:%v", userId)
				break
			}
			// log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
		}
	}

	// a.WriteMsg(login)
	u.SendMsg(login, "Login_S2C")
	client.(*ClientInfo).agent.SetUserData(u)

}

// 玩家離開子遊戲服務
func playerExitGame(args []interface{}) {
	cInfo := args[0].(common.UserInfo)
	common.Debug_log("用户登出中心服成功,userID=%d\n", cInfo.UserID)
	rid, ok := hall.UserRoom.Load(cInfo.UserID)
	if ok {
		room, _ := hall.RoomRecord.Load(rid)
		for k, v := range room.(*Room).PlayerList {
			if v.Id == cInfo.UserID && v.IsRobot == false {
				room.(*Room).PlayerList = append(room.(*Room).PlayerList[:k], room.(*Room).PlayerList[k+1:]...) //这里两个同样的用户名退出，会报错
				log.Debug("%v 玩家登出从房间列表删除成功 ~", v.Id)
			}
		}
	}

	unbindAgentWithUser(cInfo.UserID)
}

// 登陸中心服後的處理
func respondStart(args []interface{}) {
	// common.Debug_log("gameModule respondStart")
	go func() {
		StartHttpServer() // 监听接口
	}()

	arrPackages := args[0].([]common.LoginResponse)
	mapTaxPercent = make(map[int]float64)
	for _, v := range arrPackages {
		mapTaxPercent[v.PackageID] = float64(v.TaxPercent) * math.Pow10(-2)
	}
	serverStart := fmt.Sprintf("彩源猜大小游戏服务器启动成功\n启动时间:" + common.TimeNowStr() + "\n版本号:" + versionCode)
	canterport := common.IntToStr(conf.Server.CenterServerPort)
	canterurl := fmt.Sprintf("ws://" + conf.Server.CenterServer + ":" + canterport)
	switch canterurl {
	case "ws://161.117.178.174:12345":
		serverStart = fmt.Sprintf(serverStart + "\n环境:DEV")
		common.Debug_log(serverStart)
	case "ws://172.16.100.2:9502", "ws://172.16.1.41:9502":
		serverStart = fmt.Sprintf(serverStart + "\n环境:PRE")
		common.Debug_log(serverStart)
		common.SendTextToTelegramChat(common.TgbotChatID, serverStart, common.TgbotToken)
	default:
		serverStart = fmt.Sprintf(serverStart + "\n环境:OL")
		common.Debug_log(serverStart)
		common.SendTextToTelegramChat(common.TgbotChatID, serverStart, common.TgbotToken)
	}
	// common.Debug_log(canterurl)

}

func respondWinMoney(args []interface{}) {
	// common.Debug_log("gameModule respondWinMoney")
	data := args[0].(common.AmountFlowRes)
	record := UpdateTurnoverRecord(data)
	if record == nil {
		return
	}

	client, ok := AgentFromuserID_.Load(data.UserID)
	if !ok {
		common.Debug_log("加钱,用户%d不存在\n", data.UserID)
		return
	}
	a := client.(*ClientInfo).agent
	p, ok := a.UserData().(*Player)
	p.updateBalance(data)

}

func respondLoseMoney(args []interface{}) {
	// common.Debug_log("gameModule respondLoseMoney")
	data := args[0].(common.AmountFlowRes)
	record := UpdateTurnoverRecord(data)
	if record == nil {
		return
	}
	client, ok := AgentFromuserID_.Load(data.UserID)
	if !ok {
		common.Debug_log("扣钱,用户%d不存在\n", data.UserID)
		return
	}
	a := client.(*ClientInfo).agent
	p, ok := a.UserData().(*Player)
	p.updateBalance(data)

}

// UpdateTurnoverRecord 更新流水记录
func UpdateTurnoverRecord(data common.AmountFlowRes) *TurnoverRecord {
	cmd := SearchCMD{
		DBName: dbName,
		CName:  "TURNOVER", //DateFromTimeStamp(data.TimeStamp),
		ItemID: bson.ObjectIdHex(data.Order),
		Update: bson.M{"$set": bson.M{
			"tax":         data.Tax,
			"valid":       true,
			"balance":     data.Balance,
			"lockBalance": data.LockBalance,
		}},
	}
	record := &TurnoverRecord{}
	ok := FindAndUpdateItemByID(cmd, record)
	if ok {
		return record
	}
	return nil
}

// 更新用戶餘額
func (user *Player) updateBalance(data common.AmountFlowRes) {
	// common.Debug_log("gameModule *BaseUser updateBalance")
	common.Debug_log("玩家:%v 餘額更新為:%v 鎖定金額更新為:%v", user.Id, data.Balance-data.LockBalance, data.LockBalance)
	user.Account = data.Balance - data.LockBalance
	user.LockMoney = data.LockBalance
}
