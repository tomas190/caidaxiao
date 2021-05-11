package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/log"
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
			leave.PlayerInfo.Id = p.Id
			leave.PlayerInfo.NickName = p.NickName
			leave.PlayerInfo.HeadImg = p.HeadImg
			leave.PlayerInfo.Account = p.Account
			p.SendMsg(leave)
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
		if p.MaxBet > 0 {
			if totalBet+m.DownBet > p.MaxBet {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETLIMITBET
				p.SendMsg(data)
				return
			}
		}

		// 设定单个区域限红为1000
		if m.DownPot == msg.PotType_LeopardPot {
			if (room.PotMoneyCount.LeopardDownBet+m.DownBet)*WinLeopard > 1000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data)
				return
			}
		}
		// 设定全区的最大限红为10000
		if m.DownPot == msg.PotType_BigPot {
			if (room.PotMoneyCount.BigDownBet+m.DownBet)-room.PotMoneyCount.SmallDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data)
				return
			}
			if (p.DownBetMoney.BigDownBet+m.DownBet)-p.DownBetMoney.SmallDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data)
				return
			}
		}
		if m.DownPot == msg.PotType_SmallPot {
			if (room.PotMoneyCount.SmallDownBet+m.DownBet)-room.PotMoneyCount.BigDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data)
				return
			}
			if (p.DownBetMoney.SmallDownBet+m.DownBet)-p.DownBetMoney.BigDownBet > 10000 {
				data := &msg.ErrorMsg_S2C{}
				data.MsgData = RECODE_DOWNBETMONEYFULL
				p.SendMsg(data)
				return
			}
		}

		room.userRoomMutex.Lock()
		defer room.userRoomMutex.Unlock()

		p.IsAction = m.IsAction
		if p.IsAction == true {
			// 记录玩家在该房间总下注 和 房间注池的总金额
			if m.DownPot == msg.PotType_BigPot {
				p.DownBetMoney.BigDownBet += m.DownBet
				room.PotMoneyCount.BigDownBet += m.DownBet
				room.PlayerTotalMoney.BigDownBet += m.DownBet
			}
			if m.DownPot == msg.PotType_SmallPot {
				p.DownBetMoney.SmallDownBet += m.DownBet
				room.PotMoneyCount.SmallDownBet += m.DownBet
				room.PlayerTotalMoney.SmallDownBet += m.DownBet
			}
			if m.DownPot == msg.PotType_LeopardPot {
				p.DownBetMoney.LeopardDownBet += m.DownBet
				room.PotMoneyCount.LeopardDownBet += m.DownBet
				room.PlayerTotalMoney.LeopardDownBet += m.DownBet
			}

			p.Account -= float64(m.DownBet)
			p.TotalDownBet += m.DownBet

			// 返回玩家行动数据
			action := &msg.PlayerAction_S2C{}
			action.Id = p.Id
			action.DownBet = m.DownBet
			action.DownPot = m.DownPot
			action.IsAction = p.IsAction
			action.Account = p.Account
			room.BroadCastMsg(action)

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
			room.BroadCastMsg(potChange)
		}
	}
}

func (p *Player) BankerAction(m *msg.BankerData_C2S) {
	if m.Status == 2 {
		if p.Account > float64(m.TakeMoney) {
			rid, _ := hall.UserRoom.Load(p.Id)
			r, _ := hall.RoomRecord.Load(rid)
			if r != nil {
				room := r.(*Room)
				room.bankerList[p.Id] = m.TakeMoney
			}
		}
	}
	if m.Status == 3 {
		if p.IsBanker == true {
			p.IsDownBanker = true
		}
	}
}
