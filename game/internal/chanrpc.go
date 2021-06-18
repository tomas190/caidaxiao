package internal

import (
	common "caidaxiao/base"
	"caidaxiao/msg"
	"math"
	"strconv"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
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
	common.Debug_log("新的客户端连接:%+v", a)
	// p := &Player{}
	// p.Init()
	// p.ConnAgent = a
	// p.ConnAgent.SetUserData(p)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	p, ok := a.UserData().(*Player)

	if ok {
		log.Debug("<-------------%v 主动断开链接--------------->", p.Id)
		if p.IsAction == true { //有下注不能登出中心服等待結算後登出
			var exist bool
			rid, _ := hall.UserRoom.Load(p.Id)
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
				for _, v := range room.UserLeave {
					if v == p.Id {
						exist = true
					}
				}
				if exist == false {
					log.Debug("添加离线玩家UserLeave:%v", p.Id)
					room.UserLeave = append(room.UserLeave, p.Id)
				}
				p.IsOnline = false
				leaveHall := &msg.Logout_S2C{}
				// a.WriteMsg(leaveHall)
				p.SendMsg(leaveHall, "Logout_S2C")
			}
		} else {
			hall.UserRecord.Delete(p.Id)
			p.PlayerExitRoom()
			leaveHall := &msg.Logout_S2C{}
			p.SendMsg(leaveHall, "Logout_S2C")
			unusualLogout(a, "连接断开")
			a.Close()
		}
	}

}

// 玩家進入子遊戲服務
func playerEnterGame(args []interface{}) {
	common.Debug_log("gameModule playerEnterGame")
	cInfo := args[0].(common.UserInfo)
	client, ok := AgentFromuserID_.Load(cInfo.UserID)
	common.Debug_log("用户登陆返回,userID=%v\n", cInfo.UserID)
	if !ok {
		common.Debug_log("用户不在线,userID=%v\n", cInfo.UserID)
		return
	}
	//从中心服务器获取到的用户信息
	u := &Player{}
	u.Id = cInfo.UserID
	u.HeadImg = cInfo.UserHead
	u.NickName = cInfo.UserName
	u.PackageId = cInfo.PackageID
	u.Account = cInfo.Balance - cInfo.LockBalance

	log.Debug("玩家正常登陆:%v", u.Id)
	login := &msg.Login_S2C{}
	login.PlayerInfo = new(msg.PlayerInfo)
	login.PlayerInfo.Id = common.Int32ToStr(u.Id)
	login.PlayerInfo.NickName = u.NickName
	login.PlayerInfo.HeadImg = u.HeadImg
	login.PlayerInfo.Account = u.Account

	for _, v := range hall.roomList {
		if v != nil {
			if v.RoomId == "1" {
				login.PlayerNumR1 = v.PlayerLength()
				login.Room01 = v.IsOpenRoom
			} else if v.RoomId == "2" {
				login.PlayerNumR2 = v.PlayerLength()
				login.Room02 = v.IsOpenRoom
			}
		}
	}
	log.Debug("Room01:%v,Room02:%v", login.Room01, login.Room02)

	u.Init()

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
	unbindAgentWithUser(cInfo.UserID)
}

// 登陸中心服後的處理
func respondStart(args []interface{}) {
	common.Debug_log("gameModule respondStart")
	go func() {
		StartHttpServer() // 监听接口
	}()
	common.Debug_log("彩源猜大小游戏服务器启动成功 version:" + versionCode)
	arrPackages := args[0].([]common.LoginResponse)
	mapTaxPercent = make(map[int]float64)
	for _, v := range arrPackages {
		mapTaxPercent[v.PackageID] = float64(v.TaxPercent) * math.Pow10(-2)
	}
	common.Debug_log("respondStart=%+v", mapTaxPercent)
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
