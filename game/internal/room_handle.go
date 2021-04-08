package internal

import (
	"caidaxiao/conf"
	"caidaxiao/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"sort"
	"time"
)

//JoinGameRoom 加入游戏房间
func (r *Room) JoinGameRoom(p *Player) {
	// 插入玩家信息
	if p.IsRobot == false {
		p.FindPlayerInfo()
	}

	hall.UserRoom[p.Id] = r.RoomId

	// 将用户添加到用户列表
	r.PlayerList = append(r.PlayerList, p)

	// 玩家列表更新
	uptPlayerList := &msg.UptPlayerList_S2C{}
	uptPlayerList.PlayerList = r.RespUptPlayerList()
	r.BroadCastMsg(uptPlayerList)

	// 判断房间人数是否小于两人，否则不能开始运行
	if r.PlayerLength() < 2 {
		// 房间游戏不能开始,房间设为等待状态
		r.RoomStat = RoomStatusNone

		// 返回前端房间信息
		data := &msg.JoinRoom_S2C{}
		roomData := r.RespRoomData()
		data.RoomData = roomData
		p.SendMsg(data)

		log.Debug("房间当前人数不足，无法开始游戏 ~")
		return
	}

	// 只要不小于两人,就属于游戏状态
	p.Status = msg.PlayerStatus_PlayGame

	//返回前端房间信息
	data := &msg.JoinRoom_S2C{}
	roomData := r.RespRoomData()
	data.RoomData = roomData
	if r.GameStat == msg.GameStep_Banker {
		data.RoomData.GameTime = BankerTime - r.counter
		//log.Debug("加入房间 BankerTime: %v", msg.GameTime)
	} else if r.GameStat == msg.GameStep_Banker2 {
		data.RoomData.GameTime = Banker2Time - r.counter
	} else if r.GameStat == msg.GameStep_DownBet {
		data.RoomData.GameTime = DownBetTime - r.counter
		//log.Debug("加入房间 DownBetTime: %v", msg.GameTime)
	} else if r.GameStat == msg.GameStep_Settle {
		data.RoomData.GameTime = SettleTime - r.counter
		//log.Debug("加入房间 SettleTime: %v", msg.GameTime)
	}
	p.SendMsg(data)

	if r.RoomStat != RoomStatusRun {
		// None和Over状态都直接开始运行游戏
		r.StartGameRun()
	}
}

//GameStart 游戏开始运行
func (r *Room) StartGameRun() {
	// 当前房间人数存在两人及两人以上才开始游戏
	if r.PlayerLength() < 6 {
		// 房间游戏不能开始,房间设为等待状态
		r.RoomStat = RoomStatusNone

		log.Debug("房间人数不够，不能重新开始游戏~")
		return
	}

	r.RoomStat = RoomStatusRun

	// 玩家列表更新
	r.UpdatePlayerList()
	uptPlayerList := &msg.UptPlayerList_S2C{}
	uptPlayerList.PlayerList = r.RespUptPlayerList()
	r.BroadCastMsg(uptPlayerList)

	// 获取桌面显示的6个玩家
	num := len(r.PlayerList) - 6
	r.TablePlayer = append(r.TablePlayer, r.PlayerList[:len(r.PlayerList)-num]...)

	//// 游戏阶段行动
	//if r.IsConBanker == false {
	//	// 庄家抢庄定时
	//	r.BankerTimerTask()
	//} else {
	//	// 庄家连庄定时
	//	r.Banker2TimerTask()
	//}

	// 下注阶段定时
	r.DownBetTimerTask()
	// 结算阶段定时
	r.SettlerTimerTask()
}

//GrabDealTimerTask 庄家抢庄定时器任务
func (r *Room) BankerTimerTask() {
	r.GameStat = msg.GameStep_Banker

	// 抢庄时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Banker
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	go func() {
		for range r.clock.C {
			r.counter++
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = BankerTime
			send.GameStep = msg.GameStep_Banker
			for k, v := range r.bankerList {
				banker := &msg.BankerList{}
				banker.Id = k
				banker.TakeMoney = v
				send.BankerList = append(send.BankerList, banker)
			}
			r.BroadCastMsg(send)
			//log.Debug("BankerTime :%v", r.counter)
			if r.counter == 5 {
				// 产生庄家
				r.PlayerUpBanker()
			}
			if r.counter >= BankerTime {
				r.counter = 0
				BankerChannel <- true
				return
			}
		}
	}()
}

//GrabDealTimerTask 庄家连庄定时器任务
func (r *Room) Banker2TimerTask() {
	r.GameStat = msg.GameStep_Banker2

	for _, v := range r.PlayerList {
		if v != nil && v.Id == r.BankerId {
			v.BankerCount++
		}
	}
	// 抢庄时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Banker2
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	go func() {
		for range r.clock.C {
			r.counter++
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = Banker2Time
			send.GameStep = msg.GameStep_Banker2
			r.BroadCastMsg(send)
			//log.Debug("Banker2Time :%v", r.counter)
			if r.counter >= Banker2Time {
				r.counter = 0
				BankerChannel <- true
				return
			}
		}
	}()
}

//DownBetTimerTask 下注阶段定时器任务
func (r *Room) DownBetTimerTask() {
	//go func() {
	//	select {
	//	case t := <-BankerChannel:
	//		if t == true {
	//			r.DownBetTime()
	//			r.counter = 0
	//			DownBetChannel <- true
	//			return
	//		}
	//	}
	//}()

	// 房间状态
	r.GameStat = msg.GameStep_DownBet

	log.Debug("庄家金额:%v", r.BankerMoney)

	// 下注时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_DownBet
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter
	send.GameTime = DownBetTime
	send.GameStep = msg.GameStep_DownBet
	r.BroadCastMsg(send)

	// 机器开始下注
	r.RobotsDownBet()

	// 定时
	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = DownBetTime
			send.GameStep = msg.GameStep_DownBet
			r.BroadCastMsg(send)
			if r.counter >= DownBetTime {
				r.counter = 0
				DownBetChannel <- true
				return
			}
		}
	}()
}

//DownBetTime 下注计时
func (r *Room) DownBetTime() {
	// 房间状态
	r.GameStat = msg.GameStep_DownBet

	log.Debug("庄家金额:%v", r.BankerMoney)
	// 机器开始下注
	r.RobotsDownBet()

	// 下注时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_DownBet
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 定时
	t := time.NewTicker(time.Second)
	for range t.C {
		r.counter++
		// 发送时间
		send := &msg.SendActTime_S2C{}
		send.StartTime = r.counter
		send.GameTime = DownBetTime
		send.GameStep = msg.GameStep_DownBet
		r.BroadCastMsg(send)
		//log.Debug("DownBetTime :%v", r.counter)
		if r.counter >= DownBetTime {
			return
		}
	}
}

//SettlerTimerTask 结算阶段定时器任务
func (r *Room) SettlerTimerTask() {
	go func() {
		select {
		case t := <-DownBetChannel:
			if t == true {
				//开始比牌结算任务
				r.CompareSettlement()

				//开始新一轮游戏,重复调用StartGameRun函数
				defer r.StartGameRun()
				return
			}
		}
	}()
}

//CompareSettlement 开始比牌结算
func (r *Room) CompareSettlement() {

	r.GameStat = msg.GameStep_Settle

	// 结算时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Settle
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 获取彩源数据
	r.GetCaiYuan()

	// 结算数据
	r.ResultMoney()

	// 发送结算数据
	resultData := &msg.ResultData_S2C{}
	resultData.RoomData = r.RespRoomData()
	r.BroadCastMsg(resultData)

	t := time.NewTicker(time.Second)

	for range t.C {
		r.counter++
		// 发送时间
		send := &msg.SendActTime_S2C{}
		send.StartTime = r.counter
		send.GameTime = SettleTime
		send.GameStep = msg.GameStep_Settle
		r.BroadCastMsg(send)
		//log.Debug("SettleTime :%v", r.counter)
		// 如果时间处理不及时,可以判断定时9秒的时候将处理这个数据然后发送给前端进行处理
		if r.counter >= SettleTime {
			// 踢出房间断线玩家
			r.KickOutPlayer()
			// 处理庄家
			r.HandleBanker()
			// 清理机器人
			r.CleanRobot()
			//根据时间来控制机器人数量
			r.HandleRobot()
			// 清空房间数据,开始下局游戏
			r.CleanRoomData()
			return
		}
	}
}

//ResultMoney 结算数据
func (r *Room) ResultMoney() {
	// 获取开奖结果和类型
	r.GetResultType()

	// 获取所有玩家中奖的金额
	var totalUserWin int32
	if r.LotteryResult.BigSmall == 1 {
		totalUserWin += r.PlayerTotalMoney.SmallDownBet * WinSmall
	} else if r.LotteryResult.BigSmall == 2 {
		totalUserWin += r.PlayerTotalMoney.BigDownBet * WinBig
	}
	if r.LotteryResult.SinDouble == 1 {
		totalUserWin += r.PlayerTotalMoney.SingleDownBet * WinSingle
	} else if r.LotteryResult.SinDouble == 2 {
		totalUserWin += r.PlayerTotalMoney.DoubleDownBet * WinDouble
	}
	if r.LotteryResult.CardType == msg.CardsType_Pair {
		totalUserWin += r.PlayerTotalMoney.PairDownBet * WinPair
	} else if r.LotteryResult.CardType == msg.CardsType_Straight {
		totalUserWin += r.PlayerTotalMoney.StraightDownBet * WinStraight
	} else if r.LotteryResult.CardType == msg.CardsType_Leopard {
		totalUserWin += r.PlayerTotalMoney.LeopardDownBet * WinLeopard
	}

	sur := &SurplusPoolDB{}
	sur.UpdateTime = time.Now()
	sur.TimeNow = time.Now().Format("2006-01-02 15:04:05")
	sur.Rid = r.RoomId
	sur.PlayerNum = GetPlayerCount()

	surPool := FindSurplusPool()
	if surPool != nil {
		sur.HistoryWin = surPool.HistoryWin
		sur.HistoryLose = surPool.HistoryLose
	}

	// 判断注池真实玩家总下注是否大于玩家所赢的钱,大于0庄家获利,否则庄家赔付
	bankerRes := r.PotTotalMoney() - totalUserWin
	log.Debug("房间玩家下注总和:%v,房间玩家赢钱总额:%v", r.PotTotalMoney(), totalUserWin)
	//log.Debug("庄家结算金额:%v", bankerRes)

	for _, v := range r.PlayerList {
		if v != nil && v.IsAction == true {
			if v.IsBanker == true { // 庄家开奖（包括系统坐庄）
				nowTime := time.Now().Unix()
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				reason := "庄家赢钱"
				if bankerRes > 0 { // 庄家获利
					v.WinResultMoney += float64(bankerRes)
					if v.IsRobot == false {
						c4c.BankerWinScore(v, nowTime, v.RoundId, reason)
					}
					sur.HistoryWin += v.WinResultMoney
					sur.TotalWinMoney += v.WinResultMoney

					tax := float64(bankerRes) * taxRate
					v.ResultMoney = float64(bankerRes) - tax
					v.Account += v.ResultMoney
					log.Debug("庄家Id:%v,庄家赢钱:%v", v.Id, v.ResultMoney)
				} else { // 庄家赔付
					v.LoseResultMoney -= float64(bankerRes)
					if v.IsRobot == false {
						c4c.BankerLoseScore(v, nowTime, v.RoundId, reason)
					}
					sur.HistoryLose -= v.LoseResultMoney
					sur.TotalLoseMoney -= v.LoseResultMoney

					v.ResultMoney = v.LoseResultMoney
					v.BankerMoney += v.ResultMoney
					v.Account += v.ResultMoney
					log.Debug("庄家Id:%v,庄家输钱:%v", v.Id, v.ResultMoney)
				}
				r.BankerMoney = v.BankerMoney
			} else { // 玩家开奖
				var totalWin int32
				var taxMoney int32
				var totalLose int32
				totalLose = v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet +
					v.DownBetMoney.SingleDownBet + v.DownBetMoney.DoubleDownBet +
					v.DownBetMoney.PairDownBet + v.DownBetMoney.StraightDownBet + v.DownBetMoney.LeopardDownBet
				if r.LotteryResult.BigSmall == 1 {
					totalWin += v.DownBetMoney.SmallDownBet
					taxMoney += v.DownBetMoney.SmallDownBet * WinSmall
				} else if r.LotteryResult.BigSmall == 2 {
					totalWin += v.DownBetMoney.BigDownBet
					taxMoney += v.DownBetMoney.BigDownBet * WinBig
				}
				if r.LotteryResult.SinDouble == 1 {
					totalWin += v.DownBetMoney.SingleDownBet
					taxMoney += v.DownBetMoney.SingleDownBet * WinSingle
				} else if r.LotteryResult.SinDouble == 2 {
					totalWin += v.DownBetMoney.DoubleDownBet
					taxMoney += v.DownBetMoney.DoubleDownBet * WinDouble
				}
				if r.LotteryResult.CardType == msg.CardsType_Pair {
					totalWin += v.DownBetMoney.PairDownBet
					taxMoney += v.DownBetMoney.PairDownBet * WinPair
				} else if r.LotteryResult.CardType == msg.CardsType_Straight {
					totalWin += v.DownBetMoney.StraightDownBet
					taxMoney += v.DownBetMoney.StraightDownBet * WinStraight
				} else if r.LotteryResult.CardType == msg.CardsType_Leopard {
					totalWin += v.DownBetMoney.LeopardDownBet
					taxMoney += v.DownBetMoney.LeopardDownBet * WinLeopard
				}
				nowTime := time.Now().Unix()
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				if v.IsRobot == false {
					log.Debug("id:%v,totalWin:%v,totalLose:%v", v.Id, totalWin, totalLose)
					log.Debug("downBet:%v", v.DownBetMoney)
				}
				if taxMoney > 0 {
					v.WinResultMoney = float64(taxMoney)
					sur.HistoryWin += v.WinResultMoney
					sur.TotalWinMoney += v.WinResultMoney
					reason := "ResultWinScore"
					if v.IsRobot == false {
						//同时同步赢分和输分
						c4c.UserSyncWinScore(v, nowTime, v.RoundId, reason)
					}
				}
				if totalLose > 0 {
					v.LoseResultMoney = float64(-totalLose + totalWin)
					sur.HistoryLose -= v.LoseResultMoney
					sur.TotalLoseMoney -= v.LoseResultMoney
					reason := "ResultLoseScore"
					//同时同步赢分和输分
					if v.IsRobot == false {
						if v.LoseResultMoney != 0 {
							c4c.UserSyncLoseScore(v, nowTime, v.RoundId, reason)
						}
					}
				}

				tax := float64(taxMoney) * taxRate
				v.ResultMoney = float64(totalWin+taxMoney) - tax
				v.Account += v.ResultMoney
				v.ResultMoney -= float64(totalLose)
				// 记录玩家20句游戏Win次数
				if v.ResultMoney > 0 {
					v.TwentyData = append(v.TwentyData, 2)
				} else {
					v.TwentyData = append(v.TwentyData, 1)
				}
				if len(v.TwentyData) > 20 {
					v.TwentyData = append(v.TwentyData[:0], v.TwentyData[1:]...)
				}
				var count int32
				for _, n := range v.TwentyData {
					if n == 2 {
						count++
					}
				}
				v.WinTotalCount = count
				//log.Debug("玩家Id:%v,玩家输赢:%v,玩家金额:%v", v.Id, v.ResultMoney, v.Account)

				if v.WinTotalCount != 0 || v.LoseResultMoney != 0 {
					data := &PlayerDownBetRecode{}
					data.Id = v.Id
					data.GameId = conf.Server.GameID
					data.RoundId = v.RoundId
					data.RoomId = r.RoomId
					data.DownBetInfo = new(msg.DownBetMoney)
					data.DownBetInfo.BigDownBet = v.DownBetMoney.BigDownBet
					data.DownBetInfo.SmallDownBet = v.DownBetMoney.SmallDownBet
					data.DownBetInfo.SingleDownBet = v.DownBetMoney.SingleDownBet
					data.DownBetInfo.DoubleDownBet = v.DownBetMoney.DoubleDownBet
					data.DownBetInfo.PairDownBet = v.DownBetMoney.PairDownBet
					data.DownBetInfo.StraightDownBet = v.DownBetMoney.StraightDownBet
					data.DownBetInfo.LeopardDownBet = v.DownBetMoney.LeopardDownBet
					data.DownBetTime = nowTime
					data.StartTime = nowTime - 15
					data.EndTime = nowTime + 10
					data.CardResult = new(msg.PotWinList)
					data.CardResult.ResultNum = r.LotteryResult.ResultNum
					data.CardResult.BigSmall = r.LotteryResult.BigSmall
					data.CardResult.SinDouble = r.LotteryResult.SinDouble
					data.CardResult.CardType = r.LotteryResult.CardType
					data.SettlementFunds = v.ResultMoney
					data.SpareCash = v.Account
					data.TaxRate = taxRate
					InsertAccessData(data)
				}

				if v.WinTotalCount != 0 || v.LoseResultMoney != 0 {
					InsertSurplusPool(sur)
				}
			}
		}
	}
	log.Debug("result：%v", r.LotteryResult)
}

//GetResultType 获取结算数据和类型
func (r *Room) GetResultType() {
	num1 := r.Lottery[0]
	num2 := r.Lottery[1]
	num3 := r.Lottery[2]
	// 开奖结果
	r.LotteryResult.ResultNum = int32((num1 + num2 + num3) % 10)
	// 开奖大小
	if r.LotteryResult.ResultNum <= 4 {
		r.LotteryResult.BigSmall = 1
	} else {
		r.LotteryResult.BigSmall = 2
	}
	// 开奖单双
	number := r.LotteryResult.ResultNum % 2
	if number == 1 {
		r.LotteryResult.SinDouble = 1
	} else if number == 0 {
		r.LotteryResult.SinDouble = 2
	}
	// 开奖类型
	numSlice := r.Lottery
	sort.Ints(numSlice)
	r.GetType(numSlice)

	var potWin msg.PotWinList
	potWin.ResultNum = r.LotteryResult.ResultNum
	potWin.BigSmall = r.LotteryResult.BigSmall
	potWin.SinDouble = r.LotteryResult.SinDouble
	potWin.CardType = r.LotteryResult.CardType
	r.PotWinList = append(r.PotWinList, potWin)
	// 判断数据大于10条就删除出一条
	if len(r.PotWinList) > 6 {
		r.PotWinList = append(r.PotWinList[:0], r.PotWinList[1:]...)
	}

	var history msg.HistoryData
	history.TimeFmt = r.resultTime
	for _, v := range r.Lottery {
		history.ResNum = append(history.ResNum, int32(v))
	}
	history.Result = r.LotteryResult.ResultNum
	history.BigSmall = r.LotteryResult.BigSmall
	history.SinDouble = r.LotteryResult.SinDouble
	history.CardType = r.LotteryResult.CardType
	r.HistoryData = append(r.HistoryData, history)
	sort.Slice(r.HistoryData, func(i, j int) bool {
		if r.HistoryData[i].TimeFmt > r.HistoryData[j].TimeFmt {
			return true
		}
		return false
	})
	// 判断数据大于70条就删除出一条
	if len(r.HistoryData) > 70 {
		r.HistoryData = r.HistoryData[:len(r.HistoryData)-1]
	}

	// 存储下注记录
	var downBetHis msg.DownBetHistory
	downBetHis.TimeFmt = r.resultTime
	for _, v := range r.Lottery {
		downBetHis.ResNum = append(downBetHis.ResNum, int32(v))
	}
	downBetHis.Result = r.LotteryResult.ResultNum
	downBetHis.BigSmall = r.LotteryResult.BigSmall
	downBetHis.SinDouble = r.LotteryResult.SinDouble
	downBetHis.CardType = r.LotteryResult.CardType
	downBetHis.Result = r.LotteryResult.ResultNum
	for _, v := range r.PlayerList {
		if v != nil && v.IsAction == true {
			downBetHis.DownBetMoney = new(msg.DownBetMoney)
			downBetHis.DownBetMoney.SmallDownBet = v.DownBetMoney.SmallDownBet
			downBetHis.DownBetMoney.BigDownBet = v.DownBetMoney.BigDownBet
			downBetHis.DownBetMoney.SingleDownBet = v.DownBetMoney.SingleDownBet
			downBetHis.DownBetMoney.DoubleDownBet = v.DownBetMoney.DoubleDownBet
			downBetHis.DownBetMoney.PairDownBet = v.DownBetMoney.PairDownBet
			downBetHis.DownBetMoney.StraightDownBet = v.DownBetMoney.StraightDownBet
			downBetHis.DownBetMoney.LeopardDownBet = v.DownBetMoney.LeopardDownBet
			v.DownBetHistory = append(v.DownBetHistory, downBetHis)
			sort.Slice(v.DownBetHistory, func(i, j int) bool {
				if v.DownBetHistory[i].TimeFmt > v.DownBetHistory[j].TimeFmt {
					return true
				}
				return false
			})
			if len(v.DownBetHistory) > 70 {
				v.DownBetHistory = v.DownBetHistory[:len(v.DownBetHistory)-1]
			}
		}
	}

}
