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
	data.RoomData.GameTime = r.RoomCounter()
	data.LeftTime = r.RoomCounter()
	data.CloseTime = CloseTime
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

	case msg.GameStep_DownBet: //25~54(30s)
		GameTime = DownBetStep + DownBetTime - time.Now().Second()

	case msg.GameStep_Close: // 55~59,00~04 (10s)
		if time.Now().Second() < 5 {
			GameTime = time.Now().Second()
		} else if time.Now().Second() < 60 {
			GameTime = CloseStep + CloseTime + 5 - time.Now().Second()
		}

	default:
		break

	}
	return int32(GameTime)
}

/*
遊戲階段(每分鐘)
取號階段5(14s)
結算階段19(6s)
下注阶段25(30s)
封單階段55(10s)
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

				case GetResStep: // 5s
					r.GameStat = msg.GameStep_GetRes
					common.Debug_log("----------房間%v 奖源阶段----------", r.RoomId)
					r.HandleGetRes() // 获取结算

				case SettleStep: // 19s
					r.GameStat = msg.GameStep_Settle
					common.Debug_log("----------房間%v 开奖阶段----------", r.RoomId)
					if r.PeriodsNum == r.ResultNum { // 判断当前奖期是否与上局奖期相同
						r.Lottery = nil
					}
					// if time.Now().Minute()%10 == 0 || r.Lottery == nil { // 流局处理每10分钟
					if r.Lottery == nil { // 流局处理

						log.Debug("房間%v 当前分钟:%v,当前奖源:%v", r.RoomId, time.Now().Minute(), r.Lottery)
						r.HandleLiuJu() // 当局游戏流局处理

					} else { // 正常结算

						r.CompareSettlement() //开始比牌结算任务
					}

				case DownBetStep: //25s
					common.Debug_log("----------房間%v 下注阶段----------", r.RoomId)
					r.DownBetTimerTask() //下注阶段定时

				case CloseStep: //55s
					r.GameStat = msg.GameStep_Close
					common.Debug_log("----------房間%v 封单阶段----------", r.RoomId)
					r.HandleCloseOver() // 封单时间

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
	data.LeftTime = DownBetTime
	data.CloseTime = CloseTime
	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

	// 机器开始下注
	r.RobotsDownBet()

}

// 封单HandleCloseOver
func (r *Room) HandleCloseOver() {

	r.GameStat = msg.GameStep_Close

	// 封单时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Close
	data.RoomData = r.RespRoomData()
	data.LeftTime = CloseTime
	data.CloseTime = CloseTime
	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

	// 获取派奖前的玩家投注数据
	r.SetPlayerDownBet()

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
	data.LeftTime = GetResTime
	data.CloseTime = CloseTime
	// log.Debug("ActionTime_S2C 房間: %v   playerData:%v  HistoryData:%v ", data.RoomData.RoomId, len(data.RoomData.PlayerData), len(data.RoomData.HistoryData))

	r.BroadCastMsg(data, "ActionTime_S2C")

}

//HandleLiuJu 处理流局数据
func (r *Room) HandleLiuJu() {

	r.GameStat = msg.GameStep_LiuJu

	// 添加流局历史数据(房間非玩家)
	r.LotteryResult.TimeFmt = r.resultTime
	for _, v := range r.Lottery {
		r.LotteryResult.ResNum = append(r.LotteryResult.ResNum, int32(v))
	}
	r.LotteryResult.IsLiuJu = true
	r.HistoryData = append(r.HistoryData, r.LotteryResult)
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
	data.LeftTime = SettleTime
	data.CloseTime = CloseTime
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

	// log.Debug("ResultData_S2C 房間: %v   playerData:%v  HistoryData:%v ", resultData.RoomData.RoomId, len(resultData.RoomData.PlayerData), len(resultData.RoomData.HistoryData))

	// 结算时间
	data := &msg.ActionTime_S2C{}
	data.GameStep = msg.GameStep_Settle
	data.RoomData = RoomData
	data.LeftTime = SettleTime
	data.CloseTime = CloseTime
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

}

type userSettlement struct {
	uBetLoss float64 // 玩家總輸下注(玩家總輸)
	uBetWin  float64 // 玩家總贏下注
	uWinSum  float64 // 玩家總贏
}

type LotteryInfo struct {
	LuckyNum int32         // 公式結果
	CardType msg.CardsType // 1大 2小 3豹子
}

//ResultMoney 结算数据
func (r *Room) ResultMoney() {
	r.SettlementMutex.Lock()
	defer r.SettlementMutex.Unlock()
	if r.GameStat == msg.GameStep_LiuJu || r.IsOpenRoom == false { // 流局结算(与关闭接口有没有冲图TODO)
		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {
				// 返回下注金额
				if v.IsRobot == false {
					r.unlockUserBetMoney(v)
				}
				downBet := float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)
				v.Account += downBet
			}
		}
		log.Debug("玩家流局结算~")
	} else { // 正常结算
		log.Debug("开始处理玩家结算~")

		for _, v := range r.PlayerList {
			if v != nil && v.IsAction == true {

				us := &userSettlement{}
				us.uBetLoss = float64(v.DownBetMoney.SmallDownBet + v.DownBetMoney.BigDownBet + v.DownBetMoney.LeopardDownBet)

				Lottery := LotteryInfo{ // 一般
					LuckyNum: r.LotteryResult.Result.LuckyNum,
					CardType: r.LotteryResult.Result.CardType,
				}
				if v.PackageId == 10 { // 富鑫II
					Lottery = LotteryInfo{
						LuckyNum: r.LotteryResult.ResultFX.LuckyNum,
						CardType: r.LotteryResult.ResultFX.CardType,
					}
				}
				// 只會有一個中獎區
				if Lottery.CardType == msg.CardsType_Leopard { // 豹子
					// 中豹子 大.小下注額退一半給玩家
					refund := float64(v.DownBetMoney.SmallDownBet+v.DownBetMoney.BigDownBet) / 2
					v.Account += refund
					us.uBetLoss -= float64(v.DownBetMoney.LeopardDownBet) + refund
					us.uBetWin += float64(v.DownBetMoney.LeopardDownBet)
					us.uWinSum += float64(v.DownBetMoney.LeopardDownBet * WinLeopard)

				} else if Lottery.CardType == msg.CardsType_Small { // 小

					us.uBetLoss -= float64(v.DownBetMoney.SmallDownBet)
					us.uBetWin += float64(v.DownBetMoney.SmallDownBet)
					us.uWinSum += float64(v.DownBetMoney.SmallDownBet * WinSmall)

				} else if Lottery.CardType == msg.CardsType_Big { // 大

					us.uBetLoss -= float64(v.DownBetMoney.BigDownBet)
					us.uBetWin += float64(v.DownBetMoney.BigDownBet)
					us.uWinSum += float64(v.DownBetMoney.BigDownBet * WinBig)

				}

				nowTime := time.Now().Unix() //todo
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				if v.IsRobot == false {
					AddTurnoverRecord("UserUnLockMoney", common.AmountFlowReq{
						UserID:    v.Id,
						Money:     v.LockMoney,
						RoundID:   r.RoundID,
						Order:     bson.NewObjectId().Hex(),
						PackageID: v.PackageId,
						Reason:    "撤回投注解锁资金",
						TimeStamp: time.Now().Unix(),
					})
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
								PackageID:  v.PackageId,
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
							PackageID:  v.PackageId,
							BetMoney:   us.uBetWin,
							RoundID:    r.RoundID,
							Order:      bson.NewObjectId().Hex(),
							Reason:     reason,
							TimeStamp:  nowTime,
						})
					}
				}

				var tax float64
				if v.IsRobot == false {

					tax = minusTax(us.uWinSum, v.PackageId) // todo
				} else {
					tax = (us.uWinSum) * 0.05

				}
				// oldAccount := v.Account
				v.Account = v.Account + us.uBetWin + us.uWinSum - tax
				// if v.IsRobot == false {
				// 	log.Debug("[结算]玩家:%v,贏分下注:%v 輸分下注:%v 玩家獲利(扣税):%v 原本余额:%v 最新余额:%v", v.Id, us.uBetWin, us.uBetLoss, us.uWinSum-tax, oldAccount, v.Account)
				// }
				v.ResultMoney = (us.uWinSum - tax) - us.uBetLoss // 玩家总赢(扣税)+玩家输分下住
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
				v.IsAction = false
				if (v.WinResultMoney != 0 || v.LoseResultMoney != 0) && v.IsRobot == false { //todo
					log.Debug("玩家:%v,贏分下注:%v 輸分下注:%v 玩家獲利:%v 玩家输赢:%v,玩家金额:%v", v.Id, us.uBetWin, us.uBetLoss, us.uWinSum, v.ResultMoney, v.Account)
					// 插入盈余池数据
					// InsertSurplusPool(sur)

					// 插入玩家下注记录
					data := &PlayerDownBetRecode{}
					data.Id = common.Int32ToStr(v.Id)
					data.GameId = conf.Server.GameID
					data.RoundId = r.RoundID
					data.RoomId = r.RoomId
					data.PackageId = v.PackageId
					data.DownBetInfo = v.DownBetMoney
					data.DownBetTime = nowTime
					data.StartTime = nowTime - 55
					data.EndTime = nowTime + 5
					data.Lottery = r.Lottery
					data.CardResult = &r.LotteryResult
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
	num1 := r.Lottery[0] // 萬
	num2 := r.Lottery[1] // 千
	num3 := r.Lottery[2] // 百
	num4 := r.Lottery[3] // 十
	num5 := r.Lottery[4] // 個
	common.Debug_log("获取开奖类型~")

	// 一般開獎方式:（万位+十位）x（千位-个位）-百位  ，絕對值取個位
	res := (num1+num4)*(num2-num5) - num3
	data := strconv.Itoa(res)
	data = data[len(data)-1:]
	LuckyNum, _ := strconv.Atoi(data)
	// 开奖结果
	r.LotteryResult.Result.LuckyNum = int32(LuckyNum)
	// 开奖大小
	if r.LotteryResult.Result.LuckyNum <= 4 {
		r.LotteryResult.Result.CardType = 1
	} else {
		r.LotteryResult.Result.CardType = 2
	}

	// 富鑫II結算方式
	// 一般開獎方式:千位+百位+個位  ，絕對值取個位
	resfx := num2 + num5 + num3
	datafx := strconv.Itoa(resfx)
	datafx = datafx[len(datafx)-1:]
	LuckyNumfx, _ := strconv.Atoi(datafx)
	// 开奖结果
	r.LotteryResult.ResultFX.LuckyNum = int32(LuckyNumfx)
	// 开奖大小
	if r.LotteryResult.ResultFX.LuckyNum <= 4 {
		r.LotteryResult.ResultFX.CardType = 1
	} else {
		r.LotteryResult.ResultFX.CardType = 2
	}

	r.ResultNum = r.PeriodsNum
	// 是否為 豹子(千位、百位、个位)
	r.GetType(r.Lottery)

	r.LotteryResult.TimeFmt = r.resultTime
	for _, v := range r.Lottery {
		r.LotteryResult.ResNum = append(r.LotteryResult.ResNum, int32(v))
	}
	r.LotteryResult.IsLiuJu = false

	c := r.LotteryResult
	fmt.Printf("~~~~~~~~~~~~~~~~開獎結果~~~~~~~~~~~~~~~~\n房間:%v\n期數:%v\n時間戳:%v\n獎號:%v\n一般結果:\n(%v+%v)*(%v-%v)-%v=%v\n%v\n富鑫結果:\n%v+%v+%v=%v\n%v\n", r.RoomId, r.ResultNum, c.TimeFmt, c.ResNum, num1, num4, num2, num5, num3, c.Result.LuckyNum, c.Result.CardType, num5, num3, num2, c.ResultFX.LuckyNum, c.ResultFX.CardType)
	r.HistoryData = append(r.HistoryData, r.LotteryResult)
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
			// downBetHis.ResNum = r.LotteryResult.ResNum
			downBetHis.Result = r.LotteryResult.Result
			downBetHis.ResultFX = r.LotteryResult.ResultFX
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
