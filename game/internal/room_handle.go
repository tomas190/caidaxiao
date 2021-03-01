package internal

import (
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
	r.UpdatePlayerList()
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
	if r.PlayerLength() < 2 {
		// 房间游戏不能开始,房间设为等待状态
		r.RoomStat = RoomStatusNone

		log.Debug("房间人数不够，不能重新开始游戏~")
		return
	}

	r.RoomStat = RoomStatusRun
	r.GameStat = msg.GameStep_Banker

	// 抢庄时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Banker
	data.StartTime = BankerTime
	r.BroadCastMsg(data)

	// 抢庄阶段定时
	r.GrabDealTimerTask()
	// 下注阶段定时
	r.DownBetTimerTask()
	// 机器开始下注
	r.RobotsDownBet()
	// 结算阶段定时
	r.SettlerTimerTask()
}

//GrabDealTimerTask 庄家阶段定时器任务
func (r *Room) GrabDealTimerTask() {
	log.Debug("------开始抢庄阶段------")
	go func() {
		for range r.clock.C {
			r.counter++
			log.Debug("BankerTime :%v", r.counter)
			if r.counter == 7 {
				// 产生庄家
				r.PlayerUpBanker()
			}
			if r.counter == BankerTime {
				r.counter = 0
				BankerChannel <- true
				return
			}
		}
	}()
}

//DownBetTimerTask 下注阶段定时器任务
func (r *Room) DownBetTimerTask() {
	log.Debug("------开始下注阶段------")
	go func() {
		select {
		case t := <-BankerChannel:
			if t == true {
				r.DownBetTime()
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
	// 下注时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_DownBet
	data.StartTime = DownBetTime
	r.BroadCastMsg(data)

	// 定时
	t := time.NewTicker(time.Second)
	for range t.C {
		r.counter++
		log.Debug("DownBetTime :%v", r.counter)
		if r.counter == DownBetTime {
			break
		}
	}
}

//SettlerTimerTask 结算阶段定时器任务
func (r *Room) SettlerTimerTask() {
	log.Debug("------开始结算阶段------")
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
	actionTime := &msg.ActionTime_S2C{}
	actionTime.GameStep = msg.GameStep_Settle
	actionTime.StartTime = SettleTime
	r.BroadCastMsg(actionTime)

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
		log.Debug("SettleTime :%v", r.counter)
		// 如果时间处理不及时,可以判断定时9秒的时候将处理这个数据然后发送给前端进行处理
		if r.counter == SettleTime {
			// 踢出房间断线玩家
			r.KickOutPlayer()
			//根据时间来控制机器人数量
			r.HandleRobot()

			// 玩家列表更新
			r.UpdatePlayerList()
			uptPlayerList := &msg.UptPlayerList_S2C{}
			uptPlayerList.PlayerList = r.RespUptPlayerList()
			r.BroadCastMsg(uptPlayerList)

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

	// 判断注池真实玩家总下注是否大于玩家所赢的钱,大于0庄家获利,否则庄家赔付
	bankerRes := r.PotTotalMoney() - totalUserWin

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
					v.ResultMoney = float64(bankerRes) - (float64(bankerRes) * taxRate)
					v.Account += v.ResultMoney
				} else { // 庄家赔付
					v.LoseResultMoney -= float64(bankerRes)
					if v.IsRobot == false {
						c4c.BankerLoseScore(v, nowTime, v.RoundId, reason)
					}
					v.ResultMoney = v.LoseResultMoney
					v.Account += v.LoseResultMoney
				}
			} else { // 玩家开奖
				var taxMoney int32
				var totalWin int32
				var totalLose int32
				totalLose = v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet +
					v.DownBetMoney.SingleDownBet + v.DownBetMoney.DoubleDownBet +
					v.DownBetMoney.PairDownBet + v.DownBetMoney.StraightDownBet + v.DownBetMoney.LeopardDownBet
				if r.LotteryResult.BigSmall == 1 {
					taxMoney += r.PlayerTotalMoney.SmallDownBet * WinSmall
					totalWin += r.PlayerTotalMoney.SmallDownBet
				} else if r.LotteryResult.BigSmall == 2 {
					taxMoney += r.PlayerTotalMoney.BigDownBet * WinBig
					totalWin += r.PlayerTotalMoney.BigDownBet
				}
				if r.LotteryResult.SinDouble == 1 {
					taxMoney += r.PlayerTotalMoney.SingleDownBet * WinSingle
					totalWin += r.PlayerTotalMoney.SingleDownBet
				} else if r.LotteryResult.SinDouble == 2 {
					taxMoney += r.PlayerTotalMoney.DoubleDownBet * WinDouble
					totalWin += r.PlayerTotalMoney.DoubleDownBet
				}
				if r.LotteryResult.CardType == msg.CardsType_Pair {
					taxMoney += r.PlayerTotalMoney.PairDownBet * WinPair
					totalWin += r.PlayerTotalMoney.PairDownBet
				} else if r.LotteryResult.CardType == msg.CardsType_Straight {
					taxMoney += r.PlayerTotalMoney.StraightDownBet * WinStraight
					totalWin += r.PlayerTotalMoney.StraightDownBet
				} else if r.LotteryResult.CardType == msg.CardsType_Leopard {
					taxMoney += r.PlayerTotalMoney.LeopardDownBet * WinLeopard
					totalWin += r.PlayerTotalMoney.LeopardDownBet
				}
				nowTime := time.Now().Unix()
				if taxMoney > 0 {
					v.WinResultMoney = float64(taxMoney)
					log.Debug("玩家金额: %v, 赢了Win: %v", v.Account, v.WinResultMoney)
					reason := "ResultWinScore"
					if v.IsRobot == false {
						//同时同步赢分和输分
						c4c.UserSyncWinScore(v, nowTime, v.RoundId, reason)
					}
				}
				if totalLose > 0 {
					v.LoseResultMoney = float64(-totalLose + totalWin)
					reason := "ResultLoseScore"
					//同时同步赢分和输分
					if v.LoseResultMoney != 0 {
						c4c.UserSyncLoseScore(v, nowTime, v.RoundId, reason)
					}
				}
				tax := float64(taxMoney) * taxRate
				v.ResultMoney = float64(totalWin+taxMoney) - tax
				v.Account += v.ResultMoney
				v.ResultMoney -= float64(totalLose)

				v.bankerMoney += v.ResultMoney
			}
		}
	}
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
	if len(r.PotWinList) > 10 {
		r.PotWinList = append(r.PotWinList[:0], r.PotWinList[1:]...)
	}
}
