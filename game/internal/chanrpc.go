package internal

import (
	"caidaxiao/msg"
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"time"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	log.Debug("<-------------新链接请求连接--------------->")

	p := &Player{}
	p.Init()
	p.ConnAgent = a
	p.ConnAgent.SetUserData(p)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	p, ok := a.UserData().(*Player)
	if ok && p.ConnAgent == a {
		log.Debug("<-------------%v 主动断开链接--------------->", p.Id)

		p.IsOnline = false
		if p.IsAction == true || p.IsBanker == true {
			rid := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
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
				leaveHall := &msg.Logout_S2C{}
				a.WriteMsg(leaveHall)
			}
		} else {
			rid := hall.UserRoom[p.Id]
			v, _ := hall.RoomRecord.Load(rid)
			if v != nil {
				room := v.(*Room)
				if p.IsBanker == true {
					room.IsConBanker = false
					nowTime := time.Now().Unix()
					p.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), room.RoomId)
					reason := "庄家申请下庄"
					c4c.BankerStatus(p, 0, nowTime, p.RoundId, reason)
				}
				hall.UserRecord.Delete(p.Id)
				p.PlayerExitRoom()
				//c4c.UserLogoutCenter(p.Id, p.Password, p.Token)  //todo
				leaveHall := &msg.Logout_S2C{}
				a.WriteMsg(leaveHall)
				a.Close()
			}
		}
	}
}
