package internal

import (
	"caidaxiao/conf"
	"caidaxiao/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"runtime"
	"sort"
	"strconv"
	"time"
)

//JoinGameRoom 加入游戏房间
func (r *Room) JoinGameRoom(p *Player) {
	//插入玩家信息
	//if p.IsRobot == false {
	//	p.FindPlayerInfo()
	//}

	r.SetUserRoom(p)

	// 将用户添加到用户列表
	r.PlayerList = append(r.PlayerList, p)

	// 玩家列表更新
	uptPlayerList := &msg.UptPlayerList_S2C{}
	uptPlayerList.PlayerList = r.RespUptPlayerList()
	r.BroadCastMsg(uptPlayerList)

	// 只要不小于两人,就属于游戏状态
	p.Status = msg.PlayerStatus_PlayGame

	// 获取桌面显示的6个玩家
	if len(r.PlayerList) >= 6 {
		num := len(r.PlayerList) - 6
		r.TablePlayer = append(r.TablePlayer, r.PlayerList[:len(r.PlayerList)-num]...)
	}

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

}

func (r *Room) GetRoomType() {
	t := time.NewTicker(time.Second)
	go func() {
		for {
			fmt.Println("时间：", time.Now().Second())
			fmt.Println("go数量:", runtime.NumGoroutine())
			select {
			case <-t.C:
				if time.Now().Second() == DownBetStep {
					fmt.Println("下注阶段")
					// 下注阶段定时
					r.DownBetTimerTask()
				}
				if time.Now().Second() == CloseStep {
					r.GameStat = msg.GameStep_Close
					fmt.Println("封单阶段")
					// 封单时间
					r.HandleCloseOver()
				}
				if time.Now().Second() == GetResStep {
					r.GameStat = msg.GameStep_GetRes
					fmt.Println("奖源阶段")
					// 获取结算
					r.HandleGetRes()
				}
				if time.Now().Second() == SettleStep {
					r.GameStat = msg.GameStep_Settle
					fmt.Println("开奖阶段")
					if time.Now().Minute() == 0 || time.Now().Minute() == 30 || r.Lottery == nil { // 流局处理
						log.Debug("当前分钟:%v,当前奖源:%v", time.Now().Minute(), r.Lottery)
						// 当局游戏流局处理
						r.HandleLiuJu()
					} else { // 正常结算
						//开始比牌结算任务
						r.CompareSettlement()
					}
				}
			}
		}
	}()
}

//DownBetTimerTask 下注阶段定时器任务
func (r *Room) DownBetTimerTask() {

	r.GameStat = msg.GameStep_DownBet

	// 玩家列表更新
	uptPlayerList := &msg.UptPlayerList_S2C{}
	uptPlayerList.PlayerList = r.RespUptPlayerList()
	r.BroadCastMsg(uptPlayerList)

	// 获取桌面显示的6个玩家
	if len(r.PlayerList) >= 6 {
		num := len(r.PlayerList) - 6
		r.TablePlayer = append(r.TablePlayer, r.PlayerList[:len(r.PlayerList)-num]...)
	}

	// 下注时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_DownBet
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 发送时间
	//send := &msg.SendActTime_S2C{}
	//send.StartTime = 0
	//send.GameTime = DownBetTime
	//send.GameStep = msg.GameStep_DownBet
	//r.BroadCastMsg(send)

	// 机器开始下注
	r.RobotsDownBet()

	// 定时
	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			log.Debug("下注时间:%v", r.counter)
			r.counter++
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = DownBetTime
			send.GameStep = msg.GameStep_DownBet
			r.BroadCastMsg(send)
			if r.GameStat == msg.GameStep_Close {
				r.counter = 0
				return
			}
		}
	}()
}

// HandleCloseOver
func (r *Room) HandleCloseOver() {

	r.GameStat = msg.GameStep_Close

	// 封单时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Close
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 发送时间
	//send := &msg.SendActTime_S2C{}
	//send.StartTime = r.counter
	//send.GameTime = CloseTime
	//send.GameStep = msg.GameStep_Close
	//r.BroadCastMsg(send)

	// 获取派奖前的玩家投注数据
	r.SetPlayerDownBet()

	// 定时
	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			log.Debug("封单时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = CloseTime
			send.GameStep = msg.GameStep_Close
			r.BroadCastMsg(send)
			if time.Now().Second() == 0 {
				r.resultTime = time.Now().Format("2006-01-02 15:04:05")
			}
			if r.GameStat == msg.GameStep_GetRes {
				r.counter = 0
				return
			}
		}
	}()
}

func (r *Room) HandleGetRes() {

	r.GameStat = msg.GameStep_GetRes

	// 获取彩源
	r.GetCaiYuan()

	// 奖源时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_GetRes
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter
	send.GameTime = GetResTime
	send.GameStep = msg.GameStep_GetRes
	r.BroadCastMsg(send)

	// 定时
	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			log.Debug("奖源时间:%v,房间状态:%v", r.counter, r.GameStat)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = GetResTime
			send.GameStep = msg.GameStep_GetRes
			r.BroadCastMsg(send)
			if r.GameStat == msg.GameStep_Settle || r.GameStat == msg.GameStep_LiuJu {
				r.counter = 0
				return
			}
		}
	}()
}

//HandleLiuJu 处理流局数据
func (r *Room) HandleLiuJu() {

	r.GameStat = msg.GameStep_LiuJu

	// 结算时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_LiuJu
	data.RoomData = r.RespRoomData()
	r.BroadCastMsg(data)

	// 添加流局历史数据
	var history msg.HistoryData
	history.TimeFmt = r.resultTime
	for _, v := range r.Lottery {
		history.ResNum = append(history.ResNum, int32(v))
	}
	history.Result = r.LotteryResult.ResultNum
	history.BigSmall = r.LotteryResult.BigSmall
	history.SinDouble = r.LotteryResult.SinDouble
	history.CardType = r.LotteryResult.CardType
	history.IsLiuJu = true
	r.HistoryData = append(r.HistoryData, history)
	sort.Slice(r.HistoryData, func(i, j int) bool {
		if r.HistoryData[i].TimeFmt > r.HistoryData[j].TimeFmt {
			return true
		}
		return false
	})


	// 判断数据大于50条就删除出一条
	if len(r.HistoryData) > 50 {
		r.HistoryData = r.HistoryData[:len(r.HistoryData)-1]
	}

	// 结算数据
	r.ResultMoney()

	// 发送结算数据
	resultData := &msg.ResultData_S2C{}
	resultData.RoomData = r.RespRoomData()
	r.BroadCastMsg(resultData)

	// 获取投注统计
	r.SeRoomTotalBet()
	// 踢出房间断线玩家
	r.KickOutPlayer()
	// 处理庄家
	//r.HandleBanker()
	// 清理机器人
	r.CleanRobot()
	//根据时间来控制机器人数量
	r.HandleRobot()
	// 清空房间数据,开始下局游戏
	r.CleanRoomData()

	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			log.Debug("流局时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = SettleTime
			send.GameStep = msg.GameStep_LiuJu
			r.BroadCastMsg(send)

			if r.GameStat == msg.GameStep_DownBet {
				r.counter = 0
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

	// 获取开奖结果和类型
	r.GetResultType()

	// 结算数据
	r.ResultMoney()

	// 发送结算数据
	resultData := &msg.ResultData_S2C{}
	resultData.RoomData = r.RespRoomData()
	r.BroadCastMsg(resultData)

	// 获取投注统计
	r.SeRoomTotalBet()
	// 踢出房间断线玩家
	r.KickOutPlayer()
	// 处理庄家
	//r.HandleBanker()
	// 清理机器人
	r.CleanRobot()
	//根据时间来控制机器人数量
	r.HandleRobot()
	// 清空房间数据,开始下局游戏
	r.CleanRoomData()

	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			log.Debug("结算时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = SettleTime
			send.GameStep = msg.GameStep_Settle
			r.BroadCastMsg(send)
			if r.GameStat == msg.GameStep_DownBet {
				r.counter = 0
				return
			}
		}
	}()
}

//ResultMoney 结算数据
func (r *Room) ResultMoney() {

	if r.GameStat == msg.GameStep_LiuJu { // 流局结算
		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {
				// 返回下注金额
				downBet := float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)
				v.Account += downBet
			}
		}
	} else { // 正常结算
		sur := &SurplusPoolDB{}
		sur.UpdateTime = time.Now()
		sur.TimeNow = time.Now().Format("2006-01-02 15:04:05")
		sur.Rid = r.RoomId
		sur.PlayerNum = GetPlayerCount() //todo

		surPool := FindSurplusPool()
		if surPool != nil {
			sur.HistoryWin = surPool.HistoryWin
			sur.HistoryLose = surPool.HistoryLose
		}

		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {
				var totalWin float64
				var taxMoney float64
				var totalLose float64

				totalLose = float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)
				if r.LotteryResult.CardType == msg.CardsType_Leopard {
					totalWin += float64(v.DownBetMoney.LeopardDownBet)
					taxMoney += float64(v.DownBetMoney.LeopardDownBet * WinLeopard)
					totalLose -= float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet)
					money := float64(v.DownBetMoney.SmallDownBet+v.DownBetMoney.BigDownBet) / 2
					totalLose += money
					v.Account += money
				} else if r.LotteryResult.BigSmall == 1 {
					totalWin += float64(v.DownBetMoney.SmallDownBet)
					taxMoney += float64(v.DownBetMoney.SmallDownBet * WinSmall)
				} else if r.LotteryResult.BigSmall == 2 {
					totalWin += float64(v.DownBetMoney.BigDownBet)
					taxMoney += float64(v.DownBetMoney.BigDownBet * WinBig)
				}

				if v.IsRobot == false {
					log.Debug("id:%v,totalWin:%v,totalLose:%v", v.Id, totalWin, totalLose)
					log.Debug("downBet:%v", v.DownBetMoney)
				}

				nowTime := time.Now().Unix() //todo
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				if taxMoney > 0 {
					v.WinResultMoney = taxMoney
					sur.HistoryWin += v.WinResultMoney
					sur.TotalWinMoney += v.WinResultMoney
					reason := "ResultWinScore" //todo
					if v.IsRobot == false {
						//同时同步赢分和输分
						c4c.UserSyncWinScore(v, nowTime, v.RoundId, reason, totalWin)
					}
				}
				if totalLose > 0 {
					v.LoseResultMoney = -totalLose + totalWin
					sur.HistoryLose -= v.LoseResultMoney
					sur.TotalLoseMoney -= v.LoseResultMoney
					reason := "ResultLoseScore" //todo
					//同时同步赢分和输分
					if v.IsRobot == false {
						if v.LoseResultMoney != 0 {
							c4c.UserSyncLoseScore(v, nowTime, v.RoundId, reason, 0-v.LoseResultMoney)
						}
					}
				}

				tax := (taxMoney) * taxRate
				v.ResultMoney = (totalWin + taxMoney) - tax
				v.Account += v.ResultMoney
				v.ResultMoney -= totalLose
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
				log.Debug("玩家Id:%v,玩家输赢:%v,玩家金额:%v", v.Id, v.ResultMoney, v.Account)

				if v.WinTotalCount != 0 || v.LoseResultMoney != 0 { //todo
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
					data.PeriodsNum = r.PeriodsNum
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
	num4 := r.Lottery[3]
	num5 := r.Lottery[4]

	// 结算方式:（万位+十位）x（千位-个位）-百位
	res := (num1+num4)*(num2-num5) - num3
	data := strconv.Itoa(res)
	data = data[len(data)-1:]
	resNum, _ := strconv.Atoi(data)
	// 开奖结果
	r.LotteryResult.ResultNum = int32(resNum)
	// 开奖大小
	if r.LotteryResult.ResultNum <= 4 {
		r.LotteryResult.BigSmall = 1
	} else {
		r.LotteryResult.BigSmall = 2
	}
	// 开奖类型 豹子(千位、百位、个位)
	r.GetType(r.Lottery)

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
	history.IsLiuJu = false
	r.HistoryData = append(r.HistoryData, history)
	sort.Slice(r.HistoryData, func(i, j int) bool {
		if r.HistoryData[i].TimeFmt > r.HistoryData[j].TimeFmt {
			return true
		}
		return false
	})
	// 判断数据大于50条就删除出一条
	if len(r.HistoryData) > 50 {
		r.HistoryData = r.HistoryData[:len(r.HistoryData)-1]
	}

	// 去重
	//r.HistoryData = removeDuplicate(r.HistoryData)

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
			if len(v.DownBetHistory) > 50 {
				v.DownBetHistory = v.DownBetHistory[:len(v.DownBetHistory)-1]
			}
		}
	}
}
