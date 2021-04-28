package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"time"
)

func init() {
	handlerReg(&msg.Ping{}, handlePing)

	handlerReg(&msg.Login_C2S{}, handleLogin)
	handlerReg(&msg.Logout_C2S{}, handleLogout)
	handlerReg(&msg.JoinRoom_C2S{}, handleJoinRoom)
	handlerReg(&msg.LeaveRoom_C2S{}, handleLeaveRoom)

	handlerReg(&msg.PlayerAction_C2S{}, handlePlayerAction)

	handlerReg(&msg.BankerData_C2S{}, handleBankerData)
	handlerReg(&msg.EmojiChat_C2S{}, handleEmojiChat)
}

// 注册消息处理函数
func handlerReg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	a := args[1].(gate.Agent)

	pingTime := time.Now().UnixNano() / 1e6
	pong := &msg.Pong{
		ServerTime: pingTime,
	}
	a.WriteMsg(pong)
}

func handleLogin(args []interface{}) {
	m := args[0].(*msg.Login_C2S)
	a := args[1].(gate.Agent)

	log.Debug("handleLogin 用户登入游戏~ :%v", m.Id)
	v, ok := hall.UserRecord.Load(m.Id)
	if ok { // 说明用户已存在
		p := v.(*Player)
		if p.ConnAgent == a { // 用户和链接都相同
			log.Debug("同一用户相同连接重复登录~")
			//ErrorResp(a, msg.ErrorMsg_UserRepeatLogin, "重复登录")
			return
		} else { // 用户相同，链接不相同
			err := hall.ReplacePlayerAgent(p.Id, a)
			if err != nil {
				log.Error("用户链接替换错误", err)
			}

			rId := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rId)
			if v != nil {
				// 玩家如果已在游戏中，则返回房间数据
				room := v.(*Room)
				for i, userId := range room.UserLeave {
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
					// 把玩家从掉线列表中移除
					if userId == p.Id {
						room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
						break
					}
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
				}
			}

			login := &msg.Login_S2C{}
			user, _ := hall.UserRecord.Load(p.Id)
			if user != nil {
				u := user.(*Player)
				login.PlayerInfo = new(msg.PlayerInfo)
				login.PlayerInfo.Id = u.Id
				login.PlayerInfo.NickName = u.NickName
				login.PlayerInfo.HeadImg = u.HeadImg
				login.PlayerInfo.Account = u.Account

				roomId := hall.UserRoom[p.Id]
				rm, _ := hall.RoomRecord.Load(roomId)
				if rm != nil {
					login.Backroom = true
				}
				for _, v := range hall.roomList {
					if v != nil {
						if v.RoomId == "1" {
							login.PlayerNumR1 = v.PlayerLength()
							login.Room01 = v.IsOpenRoom
						}
						if v.RoomId == "2" {
							login.PlayerNumR2 = v.PlayerLength()
							login.Room02 = v.IsOpenRoom
						}
					}
				}
				a.WriteMsg(login)
				//p.ConnAgent.Destroy()
				p.ConnAgent = a
				p.ConnAgent.SetUserData(u) //p
				p.IsOnline = true
				log.Debug("用户重连或顶替，发送登陆信息~,房间数据:%v", login.Backroom)
				if login.Backroom == true {
					room := rm.(*Room)
					roomData := room.RespRoomData()
					enter := &msg.EnterRoom_S2C{}
					enter.RoomData = roomData
					p.SendMsg(enter)
				}
			}

			// 处理重连
			for _, r := range hall.roomList {
				for _, v := range r.PlayerList {
					if v != nil && v.Id == p.Id {
						roomData := r.RespRoomData()
						enter := &msg.EnterRoom_S2C{}
						enter.RoomData = roomData
						p.SendMsg(enter)
					}
				}
			}
		}
	} else if !hall.agentExist(a) { // 玩家首次登入
		//u := &Player{}
		//u.Id = m.Id
		//u.Password = m.PassWord
		//u.NickName = m.Id
		//u.Account = 800
		//u.HeadImg = "3.png"
		//login := &msg.Login_S2C{}
		//login.PlayerInfo = new(msg.PlayerInfo)
		//login.PlayerInfo.Id = u.Id
		//login.PlayerInfo.NickName = u.NickName
		//login.PlayerInfo.HeadImg = u.HeadImg
		//login.PlayerInfo.Account = u.Account
		//for _, v := range hall.roomList {
		//	if v != nil {
		//		if v.RoomId == "1" {
		//			login.PlayerNumR1 = v.PlayerLength()
		//		}
		//		if v.RoomId == "2" {
		//			login.PlayerNumR2 = v.PlayerLength()
		//		}
		//	}
		//}
		//a.WriteMsg(login)
		//
		//u.Init()
		//// 重新绑定信息
		//u.ConnAgent = a
		//a.SetUserData(u)
		//
		//u.Password = m.GetPassWord()
		//u.Token = m.GetToken()
		//
		//hall.UserRecord.Store(u.Id, u)
		//hall.UserRoom[u.Id] = "1"
		//
		//rId := hall.UserRoom[u.Id]
		//v, _ := hall.RoomRecord.Load(rId)
		//if v != nil {
		//	// 玩家如果已在游戏中，则返回房间数据
		//	room := v.(*Room)
		//	for i, userId := range room.UserLeave {
		//		log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
		//		// 把玩家从掉线列表中移除
		//		if userId == u.Id {
		//			room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
		//			log.Debug("AllocateUser 清除玩家记录~:%v", userId)
		//			break
		//		}
		//		log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
		//	}
		//}
		c4c.UserLoginCenter(m.GetId(), m.GetPassWord(), m.GetToken(), func(u *Player) { //todo

			log.Debug("玩家首次登陆:%v", u.Id)
			login := &msg.Login_S2C{}
			login.PlayerInfo = new(msg.PlayerInfo)
			login.PlayerInfo.Id = u.Id
			login.PlayerInfo.NickName = u.NickName
			login.PlayerInfo.HeadImg = u.HeadImg
			login.PlayerInfo.Account = u.Account
			for _, v := range hall.roomList {
				if v != nil {
					if v.RoomId == "1" {
						login.PlayerNumR1 = v.PlayerLength()
						login.Room01 = v.IsOpenRoom
					}
					if v.RoomId == "2" {
						login.PlayerNumR2 = v.PlayerLength()
						login.Room02 = v.IsOpenRoom
					}
				}
			}
			log.Debug("Room01:%v,Room02:%v", login.Room01, login.Room02)
			a.WriteMsg(login)

			u.Init()
			// 重新绑定信息
			u.ConnAgent = a
			a.SetUserData(u)

			u.Password = m.GetPassWord()
			u.Token = m.GetToken()

			hall.UserRecord.Store(u.Id, u)

			rId := hall.UserRoom[u.Id]
			v, _ := hall.RoomRecord.Load(rId)
			if v != nil {
				// 玩家如果已在游戏中，则返回房间数据
				room := v.(*Room)
				for i, userId := range room.UserLeave {
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
					// 把玩家从掉线列表中移除
					if userId == u.Id {
						room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
						break
					}
					log.Debug("AllocateUser 长度~:%v", len(room.UserLeave))
				}
			}
		})
	}
}

func handleLogout(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveHall 玩家退出大厅~ : %v", p.Id)
	if ok {
		if p.IsAction == true {
			var exist bool
			rid := hall.UserRoom[p.Id]
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
				a.WriteMsg(leaveHall)
			}
		} else {
			c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			p.IsOnline = false
			hall.UserRecord.Delete(p.Id)
			leaveHall := &msg.Logout_S2C{}
			a.WriteMsg(leaveHall)
			p.ConnAgent.Close()
		}
	}
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*msg.JoinRoom_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleJoinRoom 玩家加入房间~ : %v", p.Id)

	if ok {
		hall.PlayerJoinRoom(m.RoomId, p)
	}
}

func handleLeaveRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveRoom 玩家退出房间~ : %v", p.Id)

	if ok {
		p.PlayerExitRoom()
	}
}

func handlePlayerAction(args []interface{}) {
	m := args[0].(*msg.PlayerAction_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handlePlayerAction 玩家开始行动~ : %v", p.Id)

	if ok {
		p.PlayerAction(m)
	}
}

func handleBankerData(args []interface{}) {
	m := args[0].(*msg.BankerData_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleBankerData 庄家行动状态~ : %v", p.Id)

	if ok {
		p.BankerAction(m)
	}
}

func handleEmojiChat(args []interface{}) {
	m := args[0].(*msg.EmojiChat_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleEmojiChat 玩家发送表情~ : %v", p.Id)
	if ok {
		roomId := hall.UserRoom[p.Id]
		r, _ := hall.RoomRecord.Load(roomId)
		if r != nil {
			room := r.(*Room)
			data := &msg.EmojiChat_S2C{}
			data.ActNum = m.ActNum
			data.ActId = p.Id
			data.GoalId = m.GoalId
			room.BroadCastMsg(data)
		}
	}
}
