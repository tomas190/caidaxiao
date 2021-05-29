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
	handlerReg(&msg.Ping{}, HeartBeatHandler)

	handlerReg(&msg.Login_C2S{}, handleLogin)
	handlerReg(&msg.Logout_C2S{}, handleLogout)
	handlerReg(&msg.JoinRoom_C2S{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom_C2S{}, handleLeaveRoom)

	handlerReg(&msg.PlayerAction_C2S{}, handlePlayerAction)

	// handlerReg(&msg.BankerData_C2S{}, handleBankerData)
	handlerReg(&msg.EmojiChat_C2S{}, handleEmojiChat)

	handlerReg(&msg.ShowTableInfo_C2S{}, ShowTableInfo)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
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
	// common.Debug_log("gameModule protobuf userLogin 用户登陆 UserID:%s UserPW:%s", m.GetId(), m.GetPassWord())

	//检查用户是否已登陆
	userID := common.Str2int32(m.GetId())

	// if userID == "77777777" { //壓測robot測試
	// 	// common.Debug_log("機器人登陆")
	// 	RobotLogin(a)
	// 	return
	// }

	client, ok := AgentFromuserID_.Load(userID)
	if ok {
		common.Debug_log("-------------- 踢人成功 --------------")

		// 如果已经登陆过，需要通知之前登陆的用户被踢出游戏
		kickedBuf := &msg.KickedOutPush{
			ServerTime: time.Now().Unix(),
			Code:       0,
			Reason:     KICKOUT_OTHER_LOGIN,
		}
		client.(*ClientInfo).agent.WriteMsg(kickedBuf) //通知用戶 client.agent 前端踢掉
		userIDFromAgent_.Delete(client.(*ClientInfo).agent)
	}

	bindAgentWithUser(a, userID)
	common.GetInstance().Login.Go("UserLogin", userID, m.GetPassWord(), m.GetToken())
}

// func handleLogin(args []interface{}) {
// 	m := args[0].(*msg.Login_C2S)
// 	a := args[1].(gate.Agent)

// 	log.Debug("handleLogin 用户登入游戏~ :%v", m.Id)
// 	v, ok := hall.UserRecord.Load(m.Id)
// 	if ok { // 说明用户已存在
// 		p := v.(*Player)
// 		if p.ConnAgent == a { // 用户和链接都相同
// 			log.Debug("同一用户相同连接重复登录~")
// 			//ErrorResp(a, msg.ErrorMsg_UserRepeatLogin, "重复登录")
// 			return
// 		} else { // 用户相同，链接不相同
// 			err := hall.ReplacePlayerAgent(p.Id, a)
// 			if err != nil {
// 				log.Error("用户链接替换错误", err)
// 			}

// 			rId, _ := hall.UserRoom.Load(p.Id)
// 			v, _ := hall.RoomRecord.Load(rId)
// 			if v != nil {
// 				// 玩家如果已在游戏中，则返回房间数据
// 				room := v.(*Room)
// 				for i, userId := range room.UserLeave {
// 					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
// 					// 把玩家从掉线列表中移除
// 					if userId == p.Id {
// 						room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
// 						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
// 						break
// 					}
// 					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
// 				}
// 			}

// 			login := &msg.Login_S2C{}
// 			user, _ := hall.UserRecord.Load(p.Id)
// 			if user != nil {
// 				u := user.(*Player)
// 				login.PlayerInfo = new(msg.PlayerInfo)
// 				login.PlayerInfo.Id = u.Id
// 				login.PlayerInfo.NickName = u.NickName
// 				login.PlayerInfo.HeadImg = u.HeadImg
// 				login.PlayerInfo.Account = u.Account

// 				rid, _ := hall.UserRoom.Load(p.Id)
// 				rm, _ := hall.RoomRecord.Load(rid)
// 				if rm != nil {
// 					login.Backroom = true
// 				}
// 				for _, v := range hall.roomList {
// 					if v != nil {
// 						if v.RoomId == "1" {
// 							login.PlayerNumR1 = v.PlayerLength()
// 							login.Room01 = v.IsOpenRoom
// 						}
// 						if v.RoomId == "2" {
// 							login.PlayerNumR2 = v.PlayerLength()
// 							login.Room02 = v.IsOpenRoom
// 						}
// 					}
// 				}
// 				// a.WriteMsg(login)
// 				//p.ConnAgent.Destroy()
// 				p.ConnAgent = a
// 				p.SendMsg(login, "Login_S2C")

// 				p.ConnAgent.SetUserData(u) //p
// 				p.IsOnline = true
// 				log.Debug("用户重连或顶替，发送登陆信息~,房间数据:%v", login.Backroom)
// 				if login.Backroom == true {
// 					room := rm.(*Room)
// 					roomData := room.RespRoomData()
// 					enter := &msg.EnterRoom_S2C{}
// 					enter.RoomData = roomData
// 					p.SendMsg(enter, "EnterRoom_S2C")
// 				}
// 			}

// 			// 处理重连
// 			for _, r := range hall.roomList {
// 				for _, v := range r.PlayerList {
// 					if v != nil && v.Id == p.Id {
// 						roomData := r.RespRoomData()
// 						enter := &msg.EnterRoom_S2C{}
// 						enter.RoomData = roomData
// 						p.SendMsg(enter, "EnterRoom_S2C")
// 					}
// 				}
// 			}
// 		}
// 	} else if !hall.agentExist(a) { // 玩家正常登入
// 		c4c.UserLoginCenter(m.GetId(), m.GetPassWord(), m.GetToken(), func(u *Player) { //todo

// 			log.Debug("玩家正常登陆:%v", u.Id)
// 			login := &msg.Login_S2C{}
// 			login.PlayerInfo = new(msg.PlayerInfo)
// 			login.PlayerInfo.Id = u.Id
// 			login.PlayerInfo.NickName = u.NickName
// 			login.PlayerInfo.HeadImg = u.HeadImg
// 			login.PlayerInfo.Account = u.Account

// 			// 判斷玩家是否第一次遊玩
// 			if u.FirstPlayerInfo() {
// 				u.InsertPlayerInfo()
// 				ServerSurPool.SumUser++
// 			}

// 			for _, v := range hall.roomList {
// 				if v != nil {
// 					if v.RoomId == "1" {
// 						login.PlayerNumR1 = v.PlayerLength()
// 						login.Room01 = v.IsOpenRoom
// 					}
// 					if v.RoomId == "2" {
// 						login.PlayerNumR2 = v.PlayerLength()
// 						login.Room02 = v.IsOpenRoom
// 					}
// 				}
// 			}
// 			log.Debug("Room01:%v,Room02:%v", login.Room01, login.Room02)

// 			u.Init()
// 			// 重新绑定信息
// 			u.ConnAgent = a

// 			// a.WriteMsg(login)
// 			u.SendMsg(login, "Login_S2C")

// 			a.SetUserData(u)

// 			// u.Password = m.GetPassWord()
// 			// u.Token = m.GetToken()

// 			limitData := LoadUserLimitBet(u)
// 			minBet, _ := strconv.Atoi(limitData.MinBet)
// 			maxBet, _ := strconv.Atoi(limitData.MaxBet)
// 			u.MinBet = int32(minBet)
// 			u.MaxBet = int32(maxBet)

// 			hall.UserRecord.Store(u.Id, u)

// 			rId, _ := hall.UserRoom.Load(u.Id)
// 			v, _ := hall.RoomRecord.Load(rId)
// 			if v != nil {
// 				// 玩家如果已在游戏中，则返回房间数据
// 				room := v.(*Room)
// 				for i, userId := range room.UserLeave {
// 					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
// 					// 把玩家从掉线列表中移除
// 					if userId == u.Id {
// 						room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
// 						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
// 						break
// 					}
// 					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
// 				}
// 			}
// 		})
// 	}
// }

func handleLogout(args []interface{}) {
	a := args[1].(gate.Agent)

	user := findUserByAgent(a)
	if user == nil { //不是虛有玩家
		common.Debug_log("玩家不存在")
		return
	}

	p, ok := a.UserData().(*Player)

	if ok {
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
			// c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			p.IsOnline = false
			hall.UserRecord.Delete(p.Id)
			leaveHall := &msg.Logout_S2C{}
			p.SendMsg(leaveHall, "Logout_S2C")
			sendLogout(p.Id) // 登出
			// a.WriteMsg(leaveHall)
			// p.ConnAgent.Close()
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
		return nil
	}
	//查询用户信息
	user, ok := allUser_.Load(common.Int32ToStr(userID.(int32)))
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

	if ok {
		log.Debug("handleJoinRoom 玩家加入房间~ : %v", p.Id)
		hall.PlayerJoinRoom(m.RoomId, p)
	}
}

func handleLeaveRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)

	if ok {
		log.Debug("handleLeaveRoom 玩家退出房间~ : %v", p.Id)
		p.PlayerExitRoom()
	}
}

func handlePlayerAction(args []interface{}) {
	m := args[0].(*msg.PlayerAction_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	if ok {
		log.Debug("handlePlayerAction 玩家开始行动~ : %v", p.Id)
		p.PlayerAction(m)
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

	if ok {
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
	}
}

func ShowTableInfo(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)

	if ok {
		log.Debug("ShowTableInfo 玩家发送房间信息~ : %v", p.Id)
		roomId, _ := hall.UserRoom.Load(p.Id)
		r, _ := hall.RoomRecord.Load(roomId)
		if r != nil {
			room := r.(*Room)
			data := &msg.ShowTableInfo_S2C{}
			data.RoomData = room.RespRoomData()
			p.SendMsg(data, "ShowTableInfo_S2C")
		}
	}
}

// 將玩家新增到線上玩家map
func bindAgentWithUser(a gate.Agent, userID int32) {

	AgentFromuserID_.Store(userID, &ClientInfo{
		agent:  a,
		expire: time.Now().Unix() + 10,
	})
	userIDFromAgent_.Store(a, userID)

}
