package internal

import (
	common "caidaxiao/base"
	"caidaxiao/msg"
	"reflect"
	"time"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

const (
	KICKOUT_OTHER_LOGIN int32 = 1 // 登入後踢前
	KICKOUT_DISABLE     int32 = 2 // 接口踢人
	KICKOUT_CLOSE_ROOM  int32 = 3 // 關房踢人
)

func init() {
	Gamehandler(&msg.Ping{}, HeartBeatHandler)

	Gamehandler(&msg.Login_C2S{}, handleLogin)
	Gamehandler(&msg.Logout_C2S{}, handleLogout)
	Gamehandler(&msg.JoinRoom_C2S{}, handleJoinRoom)
	Gamehandler(&msg.LeaveRoom_C2S{}, handleLeaveRoom)

	Gamehandler(&msg.PlayerAction_C2S{}, handlePlayerAction)

	// handlerReg(&msg.BankerData_C2S{}, handleBankerData)
	Gamehandler(&msg.EmojiChat_C2S{}, handleEmojiChat)

	Gamehandler(&msg.ShowTableInfo_C2S{}, ShowTableInfo)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func Gamehandler(m interface{}, h interface{}) {
	// skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
	skeleton.RegisterChanRPC(reflect.TypeOf(m), func(args []interface{}) {
		// 调用 handler
		skeleton.Go(func() {
			h.(func([]interface{}))(args)
		}, func() {})
	})
}

// func handlePing(args []interface{}) {
// 	a := args[1].(gate.Agent)

// 	pingTime := time.Now().UnixNano() / 1e6
// 	pong := &msg.Pong{
// 		ServerTime: pingTime,
// 	}
// 	a.WriteMsg(pong)
// }

// 客戶端傳送登入結構體
func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login_C2S) // C2S_Login結構
	a := args[1].(gate.Agent)     // 傳送結構體的玩家
	common.Debug_log("gameModule protobuf userLogin 用户登陆 UserID:%v(%T) UserPW:%v(%T) UserToken:%v(%T)", m.GetId(), m.GetId(), m.GetPassWord(), m.GetPassWord(), m.GetToken(), m.GetToken())

	//检查用户是否已登陆
	userID, _ := common.Str2int32(m.GetId())

	// if userID == "77777777" { //壓測robot測試
	// 	// common.Debug_log("機器人登陆")
	// 	RobotLogin(a)
	// 	return
	// }

	client, ok := AgentFromuserID_.Load(userID)
	if ok {

		if client.(*ClientInfo).lastLogin+800 > time.Now().UnixNano()/1e6 {
			return
		}
		common.Debug_log("-------------- 踢人成功 --------------")

		// 如果已经登陆过，需要通知之前登陆的用户被踢出游戏
		kickedBuf := &msg.KickedOutPush{
			ServerTime: time.Now().Unix(),
			Code:       0,
			Reason:     KICKOUT_OTHER_LOGIN,
		}
		client.(*ClientInfo).agent.WriteMsg(kickedBuf) //通知用戶 client.agent 前端踢掉
		userIDFromAgent_.Delete(client.(*ClientInfo).agent)
		CloseAgent(client.(*ClientInfo).agent) // 切斷先前Agent
	}

	bindAgentWithUser(a, userID)
	common.GetInstance().Login.Go("UserLogin", userID, m.GetPassWord(), m.GetToken())
}

func handleLogout(args []interface{}) {
	a := args[1].(gate.Agent)

	user := findUserByAgent(a)
	if user == nil { //不是虛有玩家
		common.Debug_log("玩家不存在")
		return
	}

	p, ok := a.UserData().(*Player)

	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		log.Debug("handleLeaveHall 玩家退出大厅~ : %v", p.Id)
		if p.IsAction == true { //有下注不能登出中心服等待結算後登出
			var exist bool
			rid, _ := hall.UserRoom.Load(p.Id)
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
				for _, uid := range room.UserLeave {
					if uid == p.Id {
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
			log.Debug("正常登出:%v", p.Id)
			// c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			p.IsOnline = false
			hall.UserRecord.Delete(p.Id)
			leaveHall := &msg.Logout_S2C{}
			p.SendMsg(leaveHall, "Logout_S2C")
			sendLogout(p.Id) // 登出
		}
	} else {
		log.Debug("UserData() err agent 找不到對應*Player")
	}
}

// 透過gateModule Agent來尋找用戶資料BaseUser
func findUserByAgent(a gate.Agent) *msg.PlayerInfo {
	// common.Debug_log("gameModule gate.Agent findUserByAgent *BaseUser")
	//查询用户ID

	userID, ok := userIDFromAgent_.Load(a)

	if !ok {
		log.Debug("userIDFromAgent_ err agent 找不到對應userID")
		CloseAgent(a) //沒有此玩家還戳登出讓他斷線
		return nil
	}
	//查询用户信息
	user, ok := allUser_.Load(userID.(int32))
	if !ok {
		log.Debug("userID找不到對應玩家PlayerInfo")
		return nil
	}
	return user.(*msg.PlayerInfo)
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*msg.JoinRoom_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)

	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		log.Debug("handleJoinRoom 玩家加入房间~ : %v", p.Id)
		hall.PlayerJoinRoom(m.RoomId, p)
	} else {
		CloseAgent(a) //沒有此玩家還加入房間讓他斷線
	}
}

func handleLeaveRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		if p.IsAction == false {
			log.Debug("handleLeaveRoom 玩家退出房间~ : %v", p.Id)
			p.PlayerExitRoom()
		}
	} else {
		CloseAgent(a) //沒有此玩家還離開房間讓他斷線
	}
}

func handlePlayerAction(args []interface{}) {
	m := args[0].(*msg.PlayerAction_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		// log.Debug("handlePlayerAction 玩家开始行动~ : %v", p.Id)
		p.PlayerAction(m)
	} else {
		CloseAgent(a) //沒有此玩家還下注讓他斷線
	}
}

// func handleBankerData(args []interface{}) {
// 	m := args[0].(*msg.BankerData_C2S)
// 	a := args[1].(gate.Agent)

// 	p, ok := a.UserData().(*Player)
// 	log.Debug("handleBankerData 庄家行动状态~ : %v", p.Id)

// 	if ok {
// 		p.BankerAction(m)
// 	}
// }

func handleEmojiChat(args []interface{}) {
	m := args[0].(*msg.EmojiChat_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)

	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		log.Debug("handleEmojiChat 玩家发送表情~ : %v", p.Id)
		rid, _ := hall.UserRoom.Load(p.Id)
		r, _ := hall.RoomRecord.Load(rid)
		if r != nil {
			room := r.(*Room)
			data := &msg.EmojiChat_S2C{}
			data.ActNum = m.ActNum
			data.ActId = common.Int32ToStr(p.Id)
			data.GoalId = m.GoalId
			room.BroadCastMsg(data, "EmojiChat_S2C")
		}
	} else {
		CloseAgent(a) //沒有此玩家還發送表情讓他斷線
	}
}

func ShowTableInfo(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)

	_, okIDFA := userIDFromAgent_.Load(a)
	if ok && okIDFA {
		log.Debug("ShowTableInfo 玩家发送房间信息~ : %v", p.Id)
		roomId, _ := hall.UserRoom.Load(p.Id)
		r, _ := hall.RoomRecord.Load(roomId)
		if r != nil {
			room := r.(*Room)
			data := &msg.ShowTableInfo_S2C{}
			data.RoomData = room.RespRoomData()
			p.SendMsg(data, "ShowTableInfo_S2C")
		}
	} else {
		CloseAgent(a) //沒有此玩家還想查看玩家列表讓他斷線
	}
}

// 將玩家新增到線上玩家map
func bindAgentWithUser(a gate.Agent, userID int32) {

	AgentFromuserID_.Store(userID, &ClientInfo{
		agent:     a,
		expire:    time.Now().Unix() + 10,
		lastLogin: time.Now().UnixNano() / 1e6, // 時間戳(毫秒)
	})
	userIDFromAgent_.Store(a, userID)

}

func CloseAgent(a gate.Agent) {
	skeleton.AfterFunc(1*time.Second, func() {
		DestroyAgent(a)
	})
}

//避免先断前端收不到踢人
func DestroyAgent(a gate.Agent) {
	ServerSurPool.AgentNum--
	a.Close()
	a.Destroy()
}
