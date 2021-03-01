package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/log"
)

//PlayerExitRoom 玩家退出房间
func (p *Player) PlayerExitRoom() {
	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
	if v != nil {
		room := v.(*Room)
		if p.IsAction == true {
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

			// 玩家列表更新
			room.UpdatePlayerList()
			uptPlayerList := &msg.UptPlayerList_S2C{}
			uptPlayerList.PlayerList = room.RespUptPlayerList()
			room.BroadCastExcept(uptPlayerList, p)

		} else {
			room.ExitFromRoom(p)
		}
	} else {
		log.Debug("Player Exit Room, But Not Found Player Room~")
	}
}

func (p *Player) PlayerAction(m *msg.PlayerAction_C2S) {
	rId := hall.UserRoom[p.Id]
	v, _ := hall.RoomRecord.Load(rId)
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
		p.IsAction = m.IsAction

		// 判断玩家是否行动做相应处理
		if p.IsAction == true {
			var downBetMoney float64
			// 判断注池限红
			if m.DownPot == msg.PotType_BigPot {
				downBetMoney = float64(m.DownBet * WinBig)
			}
			if m.DownPot == msg.PotType_SmallPot {
				downBetMoney = float64(m.DownBet * WinSmall)
			}
			if m.DownPot == msg.PotType_SinglePot {
				downBetMoney = float64(m.DownBet * WinSingle)
			}
			if m.DownPot == msg.PotType_DoublePot {
				downBetMoney = float64(m.DownBet * WinDouble)
			}
			if m.DownPot == msg.PotType_PairPot {
				downBetMoney = float64(m.DownBet * WinPair)
			}
			if m.DownPot == msg.PotType_StraightPot {
				downBetMoney = float64(m.DownBet * WinStraight)
			}
			if m.DownPot == msg.PotType_LeopardPot {
				downBetMoney = float64(m.DownBet * WinLeopard)
			}
			// 各注池下注金额加上对应的倍数
			totalMoney := room.PotMoneyCount.BigDownBet*WinBig +
				room.PotMoneyCount.SmallDownBet*WinSmall +
				room.PotMoneyCount.SingleDownBet*WinSingle +
				room.PotMoneyCount.DoubleDownBet*WinDouble +
				room.PotMoneyCount.PairDownBet*WinPair +
				room.PotMoneyCount.StraightDownBet*WinStraight +
				room.PotMoneyCount.LeopardDownBet*WinLeopard
			if float64(totalMoney)+downBetMoney > room.BankerMoney {
				log.Debug("玩家下注已限红~")
				return
			}

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
			if m.DownPot == msg.PotType_SinglePot {
				p.DownBetMoney.SingleDownBet += m.DownBet
				room.PotMoneyCount.SingleDownBet += m.DownBet
				room.PlayerTotalMoney.SingleDownBet += m.DownBet
			}
			if m.DownPot == msg.PotType_DoublePot {
				p.DownBetMoney.DoubleDownBet += m.DownBet
				room.PotMoneyCount.DoubleDownBet += m.DownBet
				room.PlayerTotalMoney.DoubleDownBet += m.DownBet
			}
			if m.DownPot == msg.PotType_PairPot {
				p.DownBetMoney.PairDownBet += m.DownBet
				room.PotMoneyCount.PairDownBet += m.DownBet
				room.PlayerTotalMoney.PairDownBet += m.DownBet
			}
			if m.DownPot == msg.PotType_StraightPot {
				p.DownBetMoney.StraightDownBet += m.DownBet
				room.PotMoneyCount.StraightDownBet += m.DownBet
				room.PlayerTotalMoney.StraightDownBet += m.DownBet
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

func (p *Player) BankerAction(m *msg.BankerData_C2S)  {
	if m.Status == 2 {
		roomId := hall.UserRoom[p.Id]
		r, _ := hall.RoomRecord.Load(roomId)
		if r != nil {
			room := r.(*Room)
			room.bankerList[p.Id] = m.TakeMoney
		}
	}
}
