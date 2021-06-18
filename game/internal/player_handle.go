package internal

import (
	common "caidaxiao/base"
	"caidaxiao/msg"
	"sync"
	"time"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

//ClientInfo 在线用户数据结构 (心跳用)
type ClientInfo struct {
	agent  gate.Agent //连接
	expire int64      //连接过期时间戳 (心跳)
}

var (
	// 下面兩個參數是已登錄子遊戲玩家為了後面方便mapping用的

	userIDFromAgent_ sync.Map // key:agent value:userid(int32)
	AgentFromuserID_ sync.Map // key:userID(int32) value:&ClientInfo
	allUser_         sync.Map // key:userID(int32) value:*msg.PlayerInfo
	emptyRoundID     = ""
)

//PlayerExitRoom 玩家退出房间
func (p *Player) PlayerExitRoom() {
	rid, _ := hall.UserRoom.Load(p.Id)
	v, _ := hall.RoomRecord.Load(rid)
	if v != nil {
		room := v.(*Room)
		if p.IsAction == true || p.IsBanker == true {
			var exist bool
			for _, v := range room.UserLeave {
				if v == p.Id {
					exist = true
				}
			}
			if exist == false {
				log.Debug("添加离线玩家UserLeave:%v", p.Id)
				room.UserLeave = append(room.UserLeave, p.Id)
			}

			leave := &msg.LeaveRoom_S2C{}
			leave.PlayerInfo = new(msg.PlayerInfo)
			leave.PlayerInfo.Id = common.Int32ToStr(p.Id)
			leave.PlayerInfo.NickName = p.NickName
			leave.PlayerInfo.HeadImg = p.HeadImg
			leave.PlayerInfo.Account = p.Account
			p.SendMsg(leave, "LeaveRoom_S2C")
		} else {
			room.ExitFromRoom(p)
		}
	} else {
		log.Debug("Player Exit Room, But Not Found Player Room~")
	}
}

func (p *Player) PlayerAction(m *msg.PlayerAction_C2S) {
	rid, _ := hall.UserRoom.Load(p.Id)
	v, _ := hall.RoomRecord.Load(rid)

	if v != nil {
		room := v.(*Room)
		// 不是下注阶段，不能进行下注
		if room.GameStat != msg.GameStep_DownBet {
			return
		}
		// 判断玩家金额是否足够下注的金额
		if p.Account < float64(m.DownBet) {
			log.Debug("玩家金额不足,不能进行下注~")
			return
		}
		if m.DownBet != 1 && m.DownBet != 5 && m.DownBet != 10 && m.DownBet != 50 &&
			m.DownBet != 100 && m.DownBet != 500 && m.DownBet != 1000 {
			log.Debug("玩家下注筹码错误!")
			return
		}

		// 当下玩家下注限红设定
		totalBet := p.DownBetMoney.BigDownBet + p.DownBetMoney.SmallDownBet + p.DownBetMoney.LeopardDownBet
		log.Debug("玩家:%v 下注金額:%v 下注類型:%v 最大限注:%v", p.Id, m.DownBet, m.DownPot, p.MaxBet)
		if p.MinBet > 0 || p.MaxBet > 0 {
			if m.DownBet < p.MinBet || totalBet+m.DownBet > p.MaxBet {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETLIMITBET
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			}
		}
		switch m.DownPot {
		case msg.PotType_LeopardPot: // 设定单个区域限红为1000
			if (room.PotMoneyCount.LeopardDownBet+m.DownBet)*WinLeopard > 1000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			}
		case msg.PotType_BigPot: // 设定全区的最大限红为10000
			if (room.PotMoneyCount.BigDownBet+m.DownBet)-room.PotMoneyCount.SmallDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			} else if (p.DownBetMoney.BigDownBet+m.DownBet)-p.DownBetMoney.SmallDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			}
		case msg.PotType_SmallPot:
			if (room.PotMoneyCount.SmallDownBet+m.DownBet)-room.PotMoneyCount.BigDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			} else if (p.DownBetMoney.SmallDownBet+m.DownBet)-p.DownBetMoney.BigDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data, "ErrorMsg_S2C")
				return
			}
		}

		room.userBetMutex.Lock()
		defer room.userBetMutex.Unlock()

		p.IsAction = m.IsAction
		if p.IsAction == true {
			// 记录玩家在该房间总下注 和 房间注池的总金额
			switch m.DownPot {
			case msg.PotType_LeopardPot:
				p.DownBetMoney.LeopardDownBet += m.DownBet
				room.PotMoneyCount.LeopardDownBet += m.DownBet
				room.PlayerTotalMoney.LeopardDownBet += m.DownBet
			case msg.PotType_BigPot:
				p.DownBetMoney.BigDownBet += m.DownBet
				room.PotMoneyCount.BigDownBet += m.DownBet
				room.PlayerTotalMoney.BigDownBet += m.DownBet
			case msg.PotType_SmallPot:
				p.DownBetMoney.SmallDownBet += m.DownBet
				room.PotMoneyCount.SmallDownBet += m.DownBet
				room.PlayerTotalMoney.SmallDownBet += m.DownBet
			}

			p.Account -= float64(m.DownBet)
			p.TotalDownBet += m.DownBet
			if p.IsRobot == false {
				lockMoney(p, float64(m.DownBet), room.RoundID)
			}
			// 返回玩家行动数据
			action := &msg.PlayerAction_S2C{}
			action.Id = common.Int32ToStr(p.Id)
			action.DownBet = m.DownBet
			action.DownPot = m.DownPot
			action.IsAction = p.IsAction
			action.Account = p.Account
			room.BroadCastMsg(action, "PlayerAction_S2C")

			// 广播房间更新注池金额
			potChange := &msg.PotChangeMoney_S2C{}
			potChange.PlayerData = p.RespPlayerData()
			potChange.PotMoneyCount = new(msg.DownBetMoney)
			potChange.PotMoneyCount.BigDownBet = room.PotMoneyCount.BigDownBet
			potChange.PotMoneyCount.SmallDownBet = room.PotMoneyCount.SmallDownBet
			potChange.PotMoneyCount.SingleDownBet = room.PotMoneyCount.SingleDownBet
			potChange.PotMoneyCount.DoubleDownBet = room.PotMoneyCount.DoubleDownBet
			potChange.PotMoneyCount.PairDownBet = room.PotMoneyCount.PairDownBet
			potChange.PotMoneyCount.StraightDownBet = room.PotMoneyCount.StraightDownBet
			potChange.PotMoneyCount.LeopardDownBet = room.PotMoneyCount.LeopardDownBet
			room.BroadCastMsg(potChange, "PotChangeMoney_S2C")
		}
	}
}

// func (p *Player) BankerAction(m *msg.BankerData_C2S) {
// 	if m.Status == 2 {
// 		if p.Account > float64(m.TakeMoney) {
// 			rid, _ := hall.UserRoom.Load(p.Id)
// 			r, _ := hall.RoomRecord.Load(rid)
// 			if r != nil {
// 				room := r.(*Room)
// 				room.bankerList[p.Id] = m.TakeMoney
// 			}
// 		}
// 	}
// 	if m.Status == 3 {
// 		if p.IsBanker == true {
// 			p.IsDownBanker = true
// 		}
// 	}
// }

// 載入玩家列表(初始化)
func LoadUserList() {
	common.Debug_log("gameModule LoadUserList")
	cmd := SearchCMD{
		DBName: dbName,
		CName:  playerInfo,
	}
	users := make([]*msg.PlayerInfo, 0)
	ok := FindAllItems(cmd, &users)
	if !ok {
		common.Debug_log("[ERROR]查无此表:USER")
		return
	}
	if len(users) == 0 {
		common.Debug_log("[ERROR]表中无资料:USER")
		return
	}
	for _, user := range users {
		// allUser[user.UserID] = user
		allUser_.Store(common.Str2int32(user.Id), user)
	}
	// serverData.SumUser = float64(len(allUser))
	allUserlength := 0
	allUser_.Range(func(_, _ interface{}) bool {
		allUserlength++
		return true
	})
	ServerSurPool.SumUser = float64(allUserlength)
}

// 客戶端非正常退出
func unusualLogout(a gate.Agent, reason string) {
	userID, ok := userIDFromAgent_.Load(a)
	if ok {
		unbindAgentWithUser(userID.(int32))
		common.Debug_log("用户ID:%d非正常退出游戏,原因:%s", userID, reason)
		sendLogout(userID.(int32))
	}
}

// 清除線上玩家map中的資料
func unbindAgentWithUser(userID int32) {
	client, ok := AgentFromuserID_.Load(userID)
	if ok {
		client.(*ClientInfo).agent.Destroy()
		AgentFromuserID_.Delete(userID)
		userIDFromAgent_.Delete(client.(*ClientInfo).agent)
	}
}

// 客戶端登出
func sendLogout(userID int32) {
	common.GetInstance().Login.Go("UserLogout", userID)
}

// 更新玩家資訊(玩家結算更新餘額)要存到DB
func UpdateUserData(userID int32) {
	user, ok := allUser_.Load(userID)
	if !ok {
		return
	}
	update := bson.M{
		"$set": bson.M{
			"nickname": user.(*msg.PlayerInfo).NickName,
			"headimg":  user.(*msg.PlayerInfo).HeadImg,
			"account":  user.(*msg.PlayerInfo).Account,
		}}
	cmd := SearchCMD{
		DBName: dbName,
		CName:  playerInfo,
		ItemID: bson.ObjectId(user.(*msg.PlayerInfo).Id),
		Update: update,
	}
	UpdateItemByID(cmd)
}

// 關閉服務時儲存所有用戶訊息
func SaveAllUserInfo() {
	common.Debug_log("gameModule SaveALLUserInfo")

	pairs := make([]interface{}, 0)

	allUser_.Range(func(_, user interface{}) bool {
		selector := bson.M{"_id": user.(*msg.PlayerInfo).Id}
		update := bson.M{
			"$set": bson.M{
				"nickname": user.(*msg.PlayerInfo).NickName,
				"headimg":  user.(*msg.PlayerInfo).HeadImg,
				"account":  user.(*msg.PlayerInfo).Account,
			}}
		pairs = append(pairs, selector, update)
		return true
	})

	if len(pairs) == 0 {
		common.Debug_log("[Error] gameModule SaveAllUserInfo 并无玩家资料资料")
		return
	}
	cmd := SearchCMD{
		DBName: dbName,
		CName:  playerInfo,
	}
	BulkUpdateAll(cmd, pairs)
}

// LogoutAllUsers 在服务器关闭时登出所有用户登出全部房间用户
func LogoutAllUsers() {
	allUser_.Range(func(_, v interface{}) bool {
		userID := common.Str2int32(v.(*msg.PlayerInfo).Id)
		sendLogout(userID)
		return true
	})
}

// 锁定金额
func lockMoney(user *Player, moneyLock float64, round_id string) {
	user.LockMoney += moneyLock
	AddTurnoverRecord("UserLockMoney", common.AmountFlowReq{
		UserID:    user.Id,
		Money:     moneyLock,
		RoundID:   round_id,
		Order:     bson.NewObjectId().Hex(),
		Reason:    "锁定用户投注的钱",
		TimeStamp: time.Now().Unix(),
	})
}

// 解锁全部金额
func unlockMoney(user *Player) float64 {
	user.LockMoney = 0
	return user.Account
}

//TurnoverRecord 流水记录
type TurnoverRecord struct {
	ID           bson.ObjectId `bson:"_id"`          //与中心服务器通信中的order字段
	UserID       int32         `bson:"userID"`       //用户ID
	MoneyChanged float64       `bson:"moneyChanged"` //资金变化
	Balance      float64       `bson:"balance"`      //用户余额
	BetMoney     float64       `bson:"betMoney"`     //用户余额
	LockBalance  float64       `bson:"lockBalance"`  //锁定金额
	Tax          float64       `bson:"tax"`          //扣税金额
	Reason       string        `bson:"reason"`       //流水产生原因
	TimeStamp    int64         `bson:"timestamp"`    //流水产生时间
	PackID       string        `bson:"packID"`       //流水产生时间
	Valid        bool          `bson:"valid"`        //是否有效
	Date         string        `bson:"date"`
}

// AddTurnoverRecord 增加一条流水记录
func AddTurnoverRecord(event string, data common.AmountFlowReq) {
	cmd := SearchCMD{
		DBName: dbName,
		CName:  "TURNOVER", // DateFromTimeStamp(data.TimeStamp),
	}
	record := &TurnoverRecord{
		ID:           bson.ObjectIdHex(data.Order),
		Date:         common.DateFromTimeStamp(data.TimeStamp),
		UserID:       data.UserID,
		MoneyChanged: data.Money,
		BetMoney:     data.BetMoney,
		Reason:       data.Reason,
		TimeStamp:    data.TimeStamp,
		PackID:       data.RoundID,
	}
	ok := AddOneItemRecord(cmd, record)
	if ok {
		common.GetInstance().Login.Go(event, data)
	}
}
