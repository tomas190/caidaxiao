package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

var (
	moneyWinToNotice float64 = 100                   // 獲勝多少需要廣播
	taxPercent               = 0.05                  // 预设税收
	mapTaxPercent            = make(map[int]float64) // 根据平台列表税收map
)

//JoinGameRoom 加入游戏房间
func (r *Room) JoinGameRoom(p *Player) {

	r.SetUserRoom(p)

	// 将用户添加到用户列表
	if PlayerExists(r.PlayerList, p) {
		r.PlayerList = append(r.PlayerList, p)
	}

	// 只要不小于两人,就属于游戏状态
	p.Status = msg.PlayerStatus_PlayGame

	if p.IsRobot == true { //todo
		return
	}

	// 获取桌面显示的6个玩家
	if len(r.PlayerList) >= 6 {
		num := len(r.PlayerList) - 6
		r.TablePlayer = nil
		r.TablePlayer = append(r.TablePlayer, r.PlayerList[:len(r.PlayerList)-num]...)
	}

	//返回前端房间信息
	data := &msg.JoinRoom_S2C{}
	roomData := r.RespRoomData()
	data.RoomData = roomData
	if r.GameStat == msg.GameStep_DownBet {
		data.RoomData.GameTime = DownBetTime - r.counter
	} else if r.GameStat == msg.GameStep_Close {
		data.RoomData.GameTime = CloseTime - r.counter
	} else if r.GameStat == msg.GameStep_GetRes {
		data.RoomData.GameTime = GetResTime - r.counter
	} else if r.GameStat == msg.GameStep_Settle {
		data.RoomData.GameTime = SettleTime - r.counter
	} else if r.GameStat == msg.GameStep_LiuJu {
		data.RoomData.GameTime = SettleTime - r.counter
	}

	// log.Debug("房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	p.SendMsg(data, "JoinRoom_S2C")

}

// 計算房間剩餘時間(每個階段還有多少時間)
func (r *Room) RoomCounter() int32 {
	var GameTime int
	switch r.GameStat {

	case msg.GameStep_GetRes: // 5~18 (14s)
		GameTime = GetResStep + GetResTime - time.Now().Second()

	case msg.GameStep_Settle, msg.GameStep_LiuJu: //19~24(6)
		GameTime = SettleStep + SettleTime - time.Now().Second()

	case msg.GameStep_DownBet: //25~44(20s)
		GameTime = DownBetStep + DownBetTime - time.Now().Second()

	case msg.GameStep_Close: // 45~59,00~04 (20s)
		if time.Now().Second() < 5 {
			GameTime = time.Now().Second()
		} else if time.Now().Second() < 60 {
			GameTime = CloseStep + CloseTime + 5 - time.Now().Second()
		}

	default:
		break

	}

	// if r.GameStat == msg.GameStep_DownBet {
	// GameTime = DownBetTime - r.counter
	// } else if r.GameStat == msg.GameStep_Close {
	// GameTime = CloseTime - r.counter
	// } else if r.GameStat == msg.GameStep_GetRes {
	// GameTime = GetResTime - r.counter
	// } else if r.GameStat == msg.GameStep_Settle {
	// GameTime = SettleTime - r.counter
	// } else if r.GameStat == msg.GameStep_LiuJu {
	// GameTime = SettleTime - r.counter
	// }
	return int32(GameTime)
}

/*
遊戲階段(每分鐘)
取號階段5(14s)
結算階段19(6s)
下注阶段25(20s)
封單階段45(20s)
*/
func (r *Room) GetRoomType() {

	go func() {
		t := time.NewTicker(time.Second)
		for {
			// log.Debug("时间:%v", time.Now().Second())
			//log.Debug("go数量:%v", runtime.NumGoroutine())
			select {
			case <-t.C:
				switch time.Now().Second() {
				case DownBetStep:
					log.Debug("----------下注阶段----------")
					// 下注阶段定时
					r.DownBetTimerTask()
					break
				case CloseStep:
					r.GameStat = msg.GameStep_Close
					log.Debug("----------封单阶段----------")
					// 封单时间
					r.HandleCloseOver()
					break
				case GetResStep:
					r.GameStat = msg.GameStep_GetRes
					log.Debug("----------奖源阶段----------")
					// 获取结算
					r.HandleGetRes()
				case SettleStep:
					r.GameStat = msg.GameStep_Settle
					log.Debug("----------开奖阶段----------")
					if r.PeriodsNum == r.ResultNum { // 判断当前奖期是否与上局奖期相同
						r.Lottery = nil
					}
					if r.Lottery == nil { // 流局处理
						log.Debug("当前分钟:%v,当前奖源:%v", time.Now().Minute(), r.Lottery)
						// 当局游戏流局处理
						r.HandleLiuJu()
					} else { // 正常结算
						//开始比牌结算任务
						r.CompareSettlement()
					}
					break
				default:
					break
				}
			}
		}
	}()
}

//DownBetTimerTask 下注阶段定时器任务
func (r *Room) DownBetTimerTask() {

	r.GameStat = msg.GameStep_DownBet
	r.RoundID = r.RoomId + bson.NewObjectId().Hex()
	// 更新玩家列表
	r.UpdatePlayerList()

	// 获取桌面显示的6个玩家
	if len(r.PlayerList) >= 6 {
		num := len(r.PlayerList) - 6
		r.TablePlayer = nil
		r.TablePlayer = append(r.TablePlayer, r.PlayerList[:len(r.PlayerList)-num]...)
	}

	// 下注时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_DownBet
	data.RoomData = r.RespRoomData()
	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

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
			//log.Debug("下注时间:%v", r.counter)
			r.counter++
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = DownBetTime
			send.GameStep = msg.GameStep_DownBet
			r.BroadCastMsg(send, "SendActTime_S2C")
			if r.GameStat == msg.GameStep_Close {
				r.counter = 0
				return
			}
		}
	}()
}

// 封单HandleCloseOver
func (r *Room) HandleCloseOver() {

	r.GameStat = msg.GameStep_Close

	// 封单时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Close
	data.RoomData = r.RespRoomData()

	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter
	send.GameTime = CloseTime
	send.GameStep = msg.GameStep_Close

	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(send, "SendActTime_S2C-1")

	// 获取派奖前的玩家投注数据
	r.SetPlayerDownBet()

	// 定时

	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			r.counter++
			//log.Debug("封单时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = CloseTime
			send.GameStep = msg.GameStep_Close
			r.BroadCastMsg(send, "SendActTime_S2C-2")
			if r.GameStat == msg.GameStep_GetRes { // 6
				r.counter = 0
				return
			}
		}
	}()
}

//取奖号时间
func (r *Room) HandleGetRes() {

	r.GameStat = msg.GameStep_GetRes

	// 获取彩源
	r.GetCaiYuan()

	// 奖源时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_GetRes
	data.RoomData = r.RespRoomData()

	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter // 0
	send.GameTime = GetResTime
	send.GameStep = msg.GameStep_GetRes

	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(send, "SendActTime_S2C-3")

	// sur = &SurplusPoolDB{}
	// sur.UpdateTime = time.Now()
	// sur.TimeNow = time.Now().Format("2006-01-02 15:04:05")
	// sur.Rid = r.RoomId
	// sur.PlayerNum = GetPlayerCount() //SUR Todo

	// surPool := FindSurplusPool()
	// if surPool != nil {
	// 	sur.HistoryWin = surPool.HistoryWin
	// 	sur.HistoryLose = surPool.HistoryLose
	// }

	// 定时
	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			//log.Debug("奖源时间:%v,房间状态:%v", r.counter, r.GameStat)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = GetResTime
			send.GameStep = msg.GameStep_GetRes
			r.BroadCastMsg(send, "SendActTime_S2C-4")
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

	// 添加流局历史数据(房間非玩家)
	history := &msg.HistoryData{}
	if r.resultTime == "" {
		r.resultTime = getNextTime()
	}
	history.TimeFmt = r.resultTime // 結算時間
	for _, v := range r.Lottery {  // 獎號(slice)
		history.ResNum = append(history.ResNum, int32(v))
	}
	history.Result = r.LotteryResult.ResultNum    // 獎號
	history.BigSmall = r.LotteryResult.BigSmall   // 大小
	history.SinDouble = r.LotteryResult.SinDouble // 沒有用到
	history.CardType = r.LotteryResult.CardType   // 豹子
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

	// log.Debug("ResultData_S2C 房間:%v    playerData:%v  HistoryData:%v ", resultData.RoomData.RoomId, len(resultData.RoomData.PlayerData), len(resultData.RoomData.HistoryData))

	r.BroadCastMsg(resultData, "ResultData_S2C")

	// 结算时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_LiuJu
	data.RoomData = r.RespRoomData()

	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

	// 获取投注统计
	r.SeRoomTotalBet()
	// 踢出房间断线玩家
	r.KickOutPlayer()
	// 清理机器人
	r.CleanRobot()
	//根据时间来控制机器人数量
	r.HandleRobot()
	// 清空房间数据,开始下局游戏
	r.CleanRoomData()

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter //0
	send.GameTime = SettleTime
	send.GameStep = msg.GameStep_LiuJu
	r.BroadCastMsg(send, "SendActTime_S2C-5")

	t := time.NewTicker(time.Second)
	go func() {
		for range t.C {
			r.counter++
			//log.Debug("流局时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = SettleTime
			send.GameStep = msg.GameStep_LiuJu
			r.BroadCastMsg(send, "SendActTime_S2C-6")

			if r.GameStat == msg.GameStep_DownBet { //3
				r.counter = 0
				return
			}
		}
	}()
}

//CompareSettlement 开始比牌结算
func (r *Room) CompareSettlement() {

	r.GameStat = msg.GameStep_Settle

	// 获取开奖结果和类型
	r.GetResultType()

	log.Debug("开始进行结算数据~")

	// 结算数据
	r.ResultMoney()

	RoomData := r.RespRoomData()
	// 发送结算数据
	resultData := &msg.ResultData_S2C{}
	resultData.RoomData = RoomData

	// log.Debug("ResultData_S2C 房間: %v   playerData:%v  HistoryData:%v ", resultData.RoomData.RoomId, len(resultData.RoomData.PlayerData), len(resultData.RoomData.HistoryData))

	r.BroadCastMsg(resultData, "ResultData_S2C")

	// 结算时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Settle
	data.RoomData = RoomData
	r.BroadCastMsg(data, "ActionTime_S2C")

	// 获取投注统计
	r.SeRoomTotalBet()
	// 踢出房间断线玩家
	r.KickOutPlayer()
	// 清理机器人
	r.CleanRobot()
	//根据时间来控制机器人数量
	r.HandleRobot()
	// 清空房间数据,开始下局游戏
	r.CleanRoomData()

	// 发送时间
	send := &msg.SendActTime_S2C{}
	send.StartTime = r.counter //0
	send.GameTime = SettleTime
	send.GameStep = msg.GameStep_Settle
	r.BroadCastMsg(send, "SendActTime_S2C-7")

	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			r.counter++
			//log.Debug("结算时间:%v", r.counter)
			// 发送时间
			send := &msg.SendActTime_S2C{}
			send.StartTime = r.counter
			send.GameTime = SettleTime
			send.GameStep = msg.GameStep_Settle
			r.BroadCastMsg(send, "SendActTime_S2C-8")
			if r.GameStat == msg.GameStep_DownBet { //3秒
				r.counter = 0
				return
			}
		}
	}()
}

type userSettlement struct {
	uBetLoss float64 // 玩家總輸下注(玩家總輸)
	uBetWin  float64 // 玩家總贏下注
	uWinSum  float64 // 玩家總贏
}

//ResultMoney 结算数据
func (r *Room) ResultMoney() {
	r.SettlementMutex.Lock()
	defer r.SettlementMutex.Unlock()
	if r.GameStat == msg.GameStep_LiuJu { // 流局结算
		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {
				// 返回下注金额
				r.unlockUserBetMoney(v)
				downBet := float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)
				v.Account += downBet
			}
		}
		log.Debug("玩家流局结算~")
	} else { // 正常结算
		log.Debug("开始处理玩家结算~")

		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {
				// var totalWin float64  //clear
				// var taxMoney float64  //clear
				// var totalLose float64 //clear
				us := &userSettlement{}
				us.uBetLoss = float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)

				// totalLose = float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet) //clear
				// 只會有一個中獎區
				if r.LotteryResult.CardType == msg.CardsType_Leopard { // 豹子
					// 中豹子 大.小下注額退一半給玩家
					refund := float64(v.DownBetMoney.SmallDownBet+v.DownBetMoney.BigDownBet) / 2
					v.Account += refund
					us.uBetLoss -= float64(v.DownBetMoney.LeopardDownBet) - refund
					us.uBetWin += float64(v.DownBetMoney.LeopardDownBet)
					us.uWinSum += float64(v.DownBetMoney.LeopardDownBet * WinLeopard)

					// totalWin += float64(v.DownBetMoney.LeopardDownBet) //clear
					// taxMoney += float64(v.DownBetMoney.LeopardDownBet * WinLeopard)//clear
					// totalLose -= float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet)//clear
					// money := float64(v.DownBetMoney.SmallDownBet+v.DownBetMoney.BigDownBet) / 2//clear
					// totalLose += money//clear
					// v.Account += money //clear
				} else if r.LotteryResult.BigSmall == 1 { // 小

					us.uBetLoss -= float64(v.DownBetMoney.SmallDownBet)
					us.uBetWin += float64(v.DownBetMoney.SmallDownBet)
					us.uWinSum += float64(v.DownBetMoney.SmallDownBet * WinSmall)

					// totalWin += float64(v.DownBetMoney.SmallDownBet)            //clear
					// taxMoney += float64(v.DownBetMoney.SmallDownBet * WinSmall) //clear
				} else if r.LotteryResult.BigSmall == 2 { // 大

					us.uBetLoss -= float64(v.DownBetMoney.BigDownBet)
					us.uBetWin += float64(v.DownBetMoney.BigDownBet)
					us.uWinSum += float64(v.DownBetMoney.BigDownBet * WinBig)

					// totalWin += float64(v.DownBetMoney.BigDownBet)          //clear
					// taxMoney += float64(v.DownBetMoney.BigDownBet * WinBig) //clear
				}

				// if v.IsRobot == false {
				// 	log.Debug("玩家:%v,贏分下注:%v 輸分下注:%v 玩家獲利:%v", v.Id, us.uBetWin, us.uBetLoss, us.uWinSum)
				// }

				nowTime := time.Now().Unix() //todo
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				if v.IsRobot == false {
					AddTurnoverRecord("UserUnLockMoney", common.AmountFlowReq{
						UserID:    v.Id,
						Money:     v.LockMoney,
						RoundID:   r.RoundID,
						Order:     bson.NewObjectId().Hex(),
						Reason:    "撤回投注解锁资金",
						TimeStamp: time.Now().Unix(),
					})
				}
				if us.uWinSum > 0 {
					v.WinResultMoney = us.uWinSum

					if v.IsRobot == false {
						ServerSurPool.TotalWin += us.uWinSum
						reason := "彩源猜大小赢钱" //todo
						//同时同步赢分和输分
						// c4c.UserSyncWinScore(v, nowTime, v.RoundId, reason, us.uBetWin)

						AddTurnoverRecord("UserWinMoney", common.AmountFlowReq{
							UserID:     v.Id,
							UserName:   v.NickName,
							Money:      v.WinResultMoney, //本局盈虧(未扣稅)
							RoomNumber: r.RoomId,
							BetMoney:   us.uBetWin,
							RoundID:    r.RoundID,
							Order:      bson.NewObjectId().Hex(),
							Reason:     reason,
							TimeStamp:  nowTime,
						})
					}
				}

				if us.uBetLoss > 0 {
					v.LoseResultMoney = -us.uBetLoss

					//同时同步赢分和输分
					if v.IsRobot == false {
						ServerSurPool.TotalLost += us.uBetLoss
						reason := "彩源猜大小输钱" //todo

						if v.LoseResultMoney != 0 {
							// c4c.UserSyncLoseScore(v, nowTime, v.RoundId, reason, us.uBetLoss)
							AddTurnoverRecord("UserLoseMoney", common.AmountFlowReq{
								UserID:     v.Id,
								UserName:   v.NickName,
								RoomNumber: r.RoomId,
								Money:      v.LoseResultMoney,
								BetMoney:   us.uBetLoss,
								RoundID:    r.RoundID,
								Order:      bson.NewObjectId().Hex(),
								Reason:     reason,
								TimeStamp:  nowTime,
							})
						}
					}
				}

				var tax float64
				if v.IsRobot == false {

					tax = minusTax(us.uWinSum, v.PackageId) // todo
				} else {
					tax = (us.uWinSum) * 0.05

				}

				v.Account += us.uBetWin + us.uWinSum - tax
				// if v.IsRobot == false {
				// 	log.Debug("玩家:%v,贏分下注:%v 輸分下注:%v 玩家獲利:%v", v.Id, us.uBetWin, us.uBetLoss, us.uWinSum)
				// }
				v.ResultMoney = us.uBetWin + us.uWinSum - tax - us.uBetLoss

				// 玩家獲利一定金額廣播
				if v.ResultMoney >= moneyWinToNotice && v.IsRobot == false {
					sendNotice(v.Id, v.NickName, v.ResultMoney)
				}

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
				unlockMoney(v)
				if (v.WinResultMoney != 0 || v.LoseResultMoney != 0) && v.IsRobot == false { //todo
					log.Debug("玩家:%v,贏分下注:%v 輸分下注:%v 玩家獲利:%v \n 玩家输赢:%v,玩家金额:%v", v.Id, us.uBetWin, us.uBetLoss, us.uWinSum, v.ResultMoney, v.Account)
					// 插入盈余池数据
					// InsertSurplusPool(sur)

					// 插入玩家下注记录
					data := &PlayerDownBetRecode{}
					data.Id = common.Int32ToStr(v.Id)
					data.GameId = conf.Server.GameID
					data.RoundId = r.RoundID
					data.RoomId = r.RoomId
					data.DownBetInfo = v.DownBetMoney
					// data.DownBetInfo = new(msg.DownBetMoney)
					// data.DownBetInfo.BigDownBet = v.DownBetMoney.BigDownBet
					// data.DownBetInfo.SmallDownBet = v.DownBetMoney.SmallDownBet
					// data.DownBetInfo.SingleDownBet = v.DownBetMoney.SingleDownBet
					// data.DownBetInfo.DoubleDownBet = v.DownBetMoney.DoubleDownBet
					// data.DownBetInfo.PairDownBet = v.DownBetMoney.PairDownBet
					// data.DownBetInfo.StraightDownBet = v.DownBetMoney.StraightDownBet
					// data.DownBetInfo.LeopardDownBet = v.DownBetMoney.LeopardDownBet
					data.DownBetTime = nowTime
					data.StartTime = nowTime - 55
					data.EndTime = nowTime + 5
					data.Lottery = r.Lottery
					data.CardResult = &r.LotteryResult
					// data.CardResult = new(msg.PotWinList)
					// data.CardResult.ResultNum = r.LotteryResult.ResultNum
					// data.CardResult.BigSmall = r.LotteryResult.BigSmall
					// data.CardResult.SinDouble = r.LotteryResult.SinDouble
					// data.CardResult.CardType = r.LotteryResult.CardType
					data.SettlementFunds = v.ResultMoney
					data.SpareCash = v.Account
					data.TaxRate = getTaxPercent(v.PackageId)
					data.PeriodsNum = r.PeriodsNum
					InsertAccessData(data)

					// 插入玩家数据
					gameData := &PlayerGameData{}
					gameData.UserId = common.Int32ToStr(v.Id)
					gameData.RoomId = r.RoomId
					gameData.DownBetInfo = v.DownBetMoney
					// gameData.DownBetInfo = new(msg.DownBetMoney)
					// gameData.DownBetInfo.BigDownBet = v.DownBetMoney.BigDownBet
					// gameData.DownBetInfo.SmallDownBet = v.DownBetMoney.SmallDownBet
					// gameData.DownBetInfo.LeopardDownBet = v.DownBetMoney.LeopardDownBet
					gameData.DownBetTime = nowTime
					gameData.StartTime = nowTime - 55
					gameData.EndTime = nowTime + 5
					gameData.SettlementFunds = v.ResultMoney
					gameData.TotalWin = v.WinResultMoney
					gameData.TotalLose = v.LoseResultMoney
					gameData.PackageId = v.PackageId
					InsertPlayerGame(gameData)
				}
			}
		}

		// 更新盈餘池
		ServerSurPool.updatePoolBalance()
	}
	// log.Debug("result：%v", r.LotteryResult)
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
	log.Debug("获取开奖类型~")

	r.ResultNum = r.PeriodsNum

	potWin := &msg.PotWinList{}
	potWin.ResultNum = r.LotteryResult.ResultNum
	potWin.BigSmall = r.LotteryResult.BigSmall
	potWin.SinDouble = r.LotteryResult.SinDouble
	potWin.CardType = r.LotteryResult.CardType
	r.PotWinList = append(r.PotWinList, potWin)
	// 判断数据大于10条就删除出一条
	if len(r.PotWinList) > 6 {
		r.PotWinList = append(r.PotWinList[:0], r.PotWinList[1:]...)
	}

	history := &msg.HistoryData{}
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

	for _, v := range r.PlayerList {
		if v != nil && v.IsAction == true && v.IsRobot == false {
			downBetHis := &msg.DownBetHistory{}
			downBetHis.TimeFmt = r.resultTime
			for _, v := range r.Lottery {
				downBetHis.ResNum = append(downBetHis.ResNum, int32(v))
			}
			downBetHis.Result = r.LotteryResult.ResultNum
			downBetHis.BigSmall = r.LotteryResult.BigSmall
			downBetHis.SinDouble = r.LotteryResult.SinDouble
			downBetHis.CardType = r.LotteryResult.CardType
			downBetHis.Result = r.LotteryResult.ResultNum

			playerHis := &msg.DownBetMoney{}

			playerHis.SmallDownBet = v.DownBetMoney.SmallDownBet
			playerHis.BigDownBet = v.DownBetMoney.BigDownBet
			playerHis.SingleDownBet = v.DownBetMoney.SingleDownBet
			playerHis.DoubleDownBet = v.DownBetMoney.DoubleDownBet
			playerHis.PairDownBet = v.DownBetMoney.PairDownBet
			playerHis.StraightDownBet = v.DownBetMoney.StraightDownBet
			playerHis.LeopardDownBet = v.DownBetMoney.LeopardDownBet
			downBetHis.DownBetMoney = playerHis
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
	log.Debug("获取历史下注记录~")
}

// 勝利提示(獲利一定金額會發布通知)
func sendNotice(userID int32, userName string, money float64) {
	common.GetInstance().Login.Go("NoticeBroadcast", common.AmountFlowReq{
		Money:    money,
		UserID:   userID,
		UserName: userName,
	})
}

// 平台扣稅比例
func minusTax(money float64, packageID int) float64 {
	if money <= 0 {
		return money
	}
	return money * getTaxPercent(packageID)
}

// 获取平台对应税收
func getTaxPercent(packageID int) float64 {
	t, ok := mapTaxPercent[packageID] //tax
	if !ok {
		return taxPercent
	}
	// common.Debug_log("存在对应ID(%d)的Tax：%.2f", packageID, t)
	return t
}

// 避免玩家切後台重覆加入列表
func PlayerExists(PlayerList []*Player, Player *Player) bool {
	for _, v := range PlayerList {
		if v.Id == Player.Id {
			return false
		}
	}
	return true
}
