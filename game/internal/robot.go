package internal

import (
	"caidaxiao/msg"
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

//机器人问题:
//1、机器人没钱怎么充值,不能再房间就直接充值,不然可以被其他用户看见
//2、机器人怎么下注，如果在桌面6个位置上，是否设置机器的下注速度和选择注池
//3、机器人选择注池的输赢,都要进行计算，只是不和盈余池牵扯，主要是前端做展示
//4、如果机器人金额如果小于50或不能参加游戏,则踢出房间删除机器人，在生成新的机器人加入该房间。

//Init 初始机器人控制中心
func (rc *RobotsCenter) Init() {
	log.Debug("-------------- RobotsCenter Init~! ---------------")
	rc.mapRobotList = make(map[uint32]*Player)
}

//CreateRobot 创建一个机器人
func (rc *RobotsCenter) CreateRobot() *Player {
	r := &Player{}
	r.Init()

	r.IsRobot = true
	//生成随机ID
	r.Id = RandomID()
	//生成随机头像IMG
	r.HeadImg = RandomIMG()
	//生成随机机器人NickName
	r.NickName = r.Id
	//r.NickName = RandomName()
	//生成机器人金币随机数
	r.Account = RandomAccount()

	return r
}

//RobotsDownBet 机器人进行下注
func (r *Room) RobotsDownBet() {
	// 线程下注
	go func() {
		time.Sleep(time.Second)
		rData := &RobotDATA{}
		rData.RoomId = r.RoomId
		rData.RoomTime = time.Now().Unix()
		rData.RobotNum = r.RobotLength()
		rData.BigPot = new(ChipDownBet)
		rData.SmallPot = new(ChipDownBet)
		rData.SinglePot = new(ChipDownBet)
		rData.DoublePot = new(ChipDownBet)
		rData.PairPot = new(ChipDownBet)
		rData.StraightPot = new(ChipDownBet)
		rData.LeopardPot = new(ChipDownBet)
		for {
			for _, v := range r.PlayerList {
				if v != nil && v.IsRobot == true {
					// 间隔时间
					timerSlice := []int32{50, 150, 20, 300, 30}
					rand.Seed(time.Now().UnixNano())
					num := rand.Intn(len(timerSlice))
					time.Sleep(time.Millisecond * time.Duration(timerSlice[num]))

					// 判断当前是否下注阶段，否则插入机器数据退出
					if r.GameStat == msg.GameStep_DownBet {
						// 判断机器人下注大于1000就不下注
						if (v.DownBetMoney.BigDownBet + v.DownBetMoney.SmallDownBet +
							v.DownBetMoney.SingleDownBet + v.DownBetMoney.DoubleDownBet +
							v.DownBetMoney.PairDownBet + v.DownBetMoney.StraightDownBet +
							v.DownBetMoney.LeopardDownBet) > 1000 {
							log.Debug("机器的下注金额大于1000~")
							continue
						}
						v.IsAction = true

						pot := RobotRandPot()
						bet := RobotRandBet()
						// 筹码大于100只投一个
						if bet < 100 {
							randNum := RandInRange(1, 5)
							for i := 0; i < randNum; i++ {
								time.Sleep(time.Millisecond * 20)
								// 判断机器人的下注筹码是否足够
								if v.Account < float64(bet) {
									log.Debug("机器的下注金额不足~")
									continue
								}
								var downBetMoney float64
								// 判断注池限红
								if pot == int32(msg.PotType_BigPot) {
									downBetMoney = float64(bet * WinBig)
								}
								if pot == int32(msg.PotType_SmallPot) {
									downBetMoney = float64(bet * WinSmall)
								}
								if pot == int32(msg.PotType_SinglePot) {
									downBetMoney = float64(bet * WinSingle)
								}
								if pot == int32(msg.PotType_DoublePot) {
									downBetMoney = float64(bet * WinDouble)
								}
								if pot == int32(msg.PotType_PairPot) {
									downBetMoney = float64(bet * WinPair)
								}
								if pot == int32(msg.PotType_StraightPot) {
									downBetMoney = float64(bet * WinStraight)
								}
								if pot == int32(msg.PotType_LeopardPot) {
									downBetMoney = float64(bet * WinLeopard)
								}
								// 各注池下注金额加上对应的倍数
								totalMoney := r.PotMoneyCount.BigDownBet*WinBig +
									r.PotMoneyCount.SmallDownBet*WinSmall +
									r.PotMoneyCount.SingleDownBet*WinSingle +
									r.PotMoneyCount.DoubleDownBet*WinDouble +
									r.PotMoneyCount.PairDownBet*WinPair +
									r.PotMoneyCount.StraightDownBet*WinStraight +
									r.PotMoneyCount.LeopardDownBet*WinLeopard
								if float64(totalMoney)+downBetMoney > 20000 { // r.BankerMoney
									//log.Debug("玩家下注已限红~")
									continue
								}

								v.Account -= float64(bet)
								v.TotalDownBet += bet

								// 记录机器人下注的注池和筹码
								if pot == int32(msg.PotType_BigPot) {
									v.DownBetMoney.BigDownBet += bet
									r.PotMoneyCount.BigDownBet += bet
									if bet == 1 {
										rData.BigPot.Chip1 += 1
									} else if bet == 5 {
										rData.BigPot.Chip5 += 1
									} else if bet == 10 {
										rData.BigPot.Chip10 += 1
									} else if bet == 50 {
										rData.BigPot.Chip50 += 1
									} else if bet == 100 {
										rData.BigPot.Chip100 += 1
									} else if bet == 500 {
										rData.BigPot.Chip500 += 1
									} else if bet == 1000 {
										rData.BigPot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_SmallPot) {
									v.DownBetMoney.SmallDownBet += bet
									r.PotMoneyCount.SmallDownBet += bet
									if bet == 1 {
										rData.SmallPot.Chip1 += 1
									} else if bet == 5 {
										rData.SmallPot.Chip5 += 1
									} else if bet == 10 {
										rData.SmallPot.Chip10 += 1
									} else if bet == 50 {
										rData.SmallPot.Chip50 += 1
									} else if bet == 100 {
										rData.SmallPot.Chip100 += 1
									} else if bet == 500 {
										rData.SmallPot.Chip500 += 1
									} else if bet == 1000 {
										rData.SmallPot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_SinglePot) {
									v.DownBetMoney.SingleDownBet += bet
									r.PotMoneyCount.SingleDownBet += bet
									if bet == 1 {
										rData.SinglePot.Chip1 += 1
									} else if bet == 5 {
										rData.SinglePot.Chip5 += 1
									} else if bet == 10 {
										rData.SinglePot.Chip10 += 1
									} else if bet == 50 {
										rData.SinglePot.Chip50 += 1
									} else if bet == 100 {
										rData.SinglePot.Chip100 += 1
									} else if bet == 500 {
										rData.SinglePot.Chip500 += 1
									} else if bet == 1000 {
										rData.SinglePot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_DoublePot) {
									v.DownBetMoney.DoubleDownBet += bet
									r.PotMoneyCount.DoubleDownBet += bet
									if bet == 1 {
										rData.DoublePot.Chip1 += 1
									} else if bet == 5 {
										rData.DoublePot.Chip5 += 1
									} else if bet == 10 {
										rData.DoublePot.Chip10 += 1
									} else if bet == 50 {
										rData.DoublePot.Chip50 += 1
									} else if bet == 100 {
										rData.DoublePot.Chip100 += 1
									} else if bet == 500 {
										rData.DoublePot.Chip500 += 1
									} else if bet == 1000 {
										rData.DoublePot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_PairPot) {
									v.DownBetMoney.PairDownBet += bet
									r.PotMoneyCount.PairDownBet += bet
									if bet == 1 {
										rData.PairPot.Chip1 += 1
									} else if bet == 5 {
										rData.PairPot.Chip5 += 1
									} else if bet == 10 {
										rData.PairPot.Chip10 += 1
									} else if bet == 50 {
										rData.PairPot.Chip50 += 1
									} else if bet == 100 {
										rData.PairPot.Chip100 += 1
									} else if bet == 500 {
										rData.PairPot.Chip500 += 1
									} else if bet == 1000 {
										rData.PairPot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_StraightPot) {
									v.DownBetMoney.StraightDownBet += bet
									r.PotMoneyCount.StraightDownBet += bet
									if bet == 1 {
										rData.StraightPot.Chip1 += 1
									} else if bet == 5 {
										rData.StraightPot.Chip5 += 1
									} else if bet == 10 {
										rData.StraightPot.Chip10 += 1
									} else if bet == 50 {
										rData.StraightPot.Chip50 += 1
									} else if bet == 100 {
										rData.StraightPot.Chip100 += 1
									} else if bet == 500 {
										rData.StraightPot.Chip500 += 1
									} else if bet == 1000 {
										rData.StraightPot.Chip1000 += 1
									}
								}
								if pot == int32(msg.PotType_LeopardPot) {
									v.DownBetMoney.LeopardDownBet += bet
									r.PotMoneyCount.LeopardDownBet += bet
									if bet == 1 {
										rData.LeopardPot.Chip1 += 1
									} else if bet == 5 {
										rData.LeopardPot.Chip5 += 1
									} else if bet == 10 {
										rData.LeopardPot.Chip10 += 1
									} else if bet == 50 {
										rData.LeopardPot.Chip50 += 1
									} else if bet == 100 {
										rData.LeopardPot.Chip100 += 1
									} else if bet == 500 {
										rData.LeopardPot.Chip500 += 1
									} else if bet == 1000 {
										rData.LeopardPot.Chip1000 += 1
									}
								}
								// 返回玩家行动数据
								action := &msg.PlayerAction_S2C{}
								action.Id = v.Id
								action.DownBet = bet
								action.DownPot = msg.PotType(pot)
								action.IsAction = v.IsAction
								action.Account = v.Account
								r.BroadCastMsg(action)

								// 广播房间更新注池金额
								potChange := &msg.PotChangeMoney_S2C{}
								potChange.PotMoneyCount = new(msg.DownBetMoney)
								potChange.PotMoneyCount.BigDownBet = r.PotMoneyCount.BigDownBet
								potChange.PotMoneyCount.SmallDownBet = r.PotMoneyCount.SmallDownBet
								potChange.PotMoneyCount.SingleDownBet = r.PotMoneyCount.SingleDownBet
								potChange.PotMoneyCount.DoubleDownBet = r.PotMoneyCount.DoubleDownBet
								potChange.PotMoneyCount.PairDownBet = r.PotMoneyCount.PairDownBet
								potChange.PotMoneyCount.StraightDownBet = r.PotMoneyCount.StraightDownBet
								potChange.PotMoneyCount.LeopardDownBet = r.PotMoneyCount.LeopardDownBet
								r.BroadCastMsg(potChange)
							}
						} else {
							// 判断机器人的下注筹码是否足够
							if v.Account < float64(bet) {
								continue
							}
							var downBetMoney float64
							// 判断注池限红
							if pot == int32(msg.PotType_BigPot) {
								downBetMoney = float64(bet * WinBig)
							}
							if pot == int32(msg.PotType_SmallPot) {
								downBetMoney = float64(bet * WinSmall)
							}
							if pot == int32(msg.PotType_SinglePot) {
								downBetMoney = float64(bet * WinSingle)
							}
							if pot == int32(msg.PotType_DoublePot) {
								downBetMoney = float64(bet * WinDouble)
							}
							if pot == int32(msg.PotType_PairPot) {
								downBetMoney = float64(bet * WinPair)
							}
							if pot == int32(msg.PotType_StraightPot) {
								downBetMoney = float64(bet * WinStraight)
							}
							if pot == int32(msg.PotType_LeopardPot) {
								downBetMoney = float64(bet * WinLeopard)
							}
							// 各注池下注金额加上对应的倍数
							totalMoney := r.PotMoneyCount.BigDownBet*WinBig +
								r.PotMoneyCount.SmallDownBet*WinSmall +
								r.PotMoneyCount.SingleDownBet*WinSingle +
								r.PotMoneyCount.DoubleDownBet*WinDouble +
								r.PotMoneyCount.PairDownBet*WinPair +
								r.PotMoneyCount.StraightDownBet*WinStraight +
								r.PotMoneyCount.LeopardDownBet*WinLeopard
							if float64(totalMoney)+downBetMoney > 20000 { // r.BankerMoney
								//log.Debug("玩家下注已限红~")
								continue
							}

							v.Account -= float64(bet)
							v.TotalDownBet += bet

							// 记录机器人下注的注池和筹码
							if pot == int32(msg.PotType_BigPot) {
								v.DownBetMoney.BigDownBet += bet
								r.PotMoneyCount.BigDownBet += bet
								if bet == 1 {
									rData.BigPot.Chip1 += 1
								} else if bet == 5 {
									rData.BigPot.Chip5 += 1
								} else if bet == 10 {
									rData.BigPot.Chip10 += 1
								} else if bet == 50 {
									rData.BigPot.Chip50 += 1
								} else if bet == 100 {
									rData.BigPot.Chip100 += 1
								} else if bet == 500 {
									rData.BigPot.Chip500 += 1
								} else if bet == 1000 {
									rData.BigPot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_SmallPot) {
								v.DownBetMoney.SmallDownBet += bet
								r.PotMoneyCount.SmallDownBet += bet
								if bet == 1 {
									rData.SmallPot.Chip1 += 1
								} else if bet == 5 {
									rData.SmallPot.Chip5 += 1
								} else if bet == 10 {
									rData.SmallPot.Chip10 += 1
								} else if bet == 50 {
									rData.SmallPot.Chip50 += 1
								} else if bet == 100 {
									rData.SmallPot.Chip100 += 1
								} else if bet == 500 {
									rData.SmallPot.Chip500 += 1
								} else if bet == 1000 {
									rData.SmallPot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_SinglePot) {
								v.DownBetMoney.SingleDownBet += bet
								r.PotMoneyCount.SingleDownBet += bet
								if bet == 1 {
									rData.SinglePot.Chip1 += 1
								} else if bet == 5 {
									rData.SinglePot.Chip5 += 1
								} else if bet == 10 {
									rData.SinglePot.Chip10 += 1
								} else if bet == 50 {
									rData.SinglePot.Chip50 += 1
								} else if bet == 100 {
									rData.SinglePot.Chip100 += 1
								} else if bet == 500 {
									rData.SinglePot.Chip500 += 1
								} else if bet == 1000 {
									rData.SinglePot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_DoublePot) {
								v.DownBetMoney.DoubleDownBet += bet
								r.PotMoneyCount.DoubleDownBet += bet
								if bet == 1 {
									rData.DoublePot.Chip1 += 1
								} else if bet == 5 {
									rData.DoublePot.Chip5 += 1
								} else if bet == 10 {
									rData.DoublePot.Chip10 += 1
								} else if bet == 50 {
									rData.DoublePot.Chip50 += 1
								} else if bet == 100 {
									rData.DoublePot.Chip100 += 1
								} else if bet == 500 {
									rData.DoublePot.Chip500 += 1
								} else if bet == 1000 {
									rData.DoublePot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_PairPot) {
								v.DownBetMoney.PairDownBet += bet
								r.PotMoneyCount.PairDownBet += bet
								if bet == 1 {
									rData.PairPot.Chip1 += 1
								} else if bet == 5 {
									rData.PairPot.Chip5 += 1
								} else if bet == 10 {
									rData.PairPot.Chip10 += 1
								} else if bet == 50 {
									rData.PairPot.Chip50 += 1
								} else if bet == 100 {
									rData.PairPot.Chip100 += 1
								} else if bet == 500 {
									rData.PairPot.Chip500 += 1
								} else if bet == 1000 {
									rData.PairPot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_StraightPot) {
								v.DownBetMoney.StraightDownBet += bet
								r.PotMoneyCount.StraightDownBet += bet
								if bet == 1 {
									rData.StraightPot.Chip1 += 1
								} else if bet == 5 {
									rData.StraightPot.Chip5 += 1
								} else if bet == 10 {
									rData.StraightPot.Chip10 += 1
								} else if bet == 50 {
									rData.StraightPot.Chip50 += 1
								} else if bet == 100 {
									rData.StraightPot.Chip100 += 1
								} else if bet == 500 {
									rData.StraightPot.Chip500 += 1
								} else if bet == 1000 {
									rData.StraightPot.Chip1000 += 1
								}
							}
							if pot == int32(msg.PotType_LeopardPot) {
								v.DownBetMoney.LeopardDownBet += bet
								r.PotMoneyCount.LeopardDownBet += bet
								if bet == 1 {
									rData.LeopardPot.Chip1 += 1
								} else if bet == 5 {
									rData.LeopardPot.Chip5 += 1
								} else if bet == 10 {
									rData.LeopardPot.Chip10 += 1
								} else if bet == 50 {
									rData.LeopardPot.Chip50 += 1
								} else if bet == 100 {
									rData.LeopardPot.Chip100 += 1
								} else if bet == 500 {
									rData.LeopardPot.Chip500 += 1
								} else if bet == 1000 {
									rData.LeopardPot.Chip1000 += 1
								}
							}
							// 返回玩家行动数据
							action := &msg.PlayerAction_S2C{}
							action.Id = v.Id
							action.DownBet = bet
							action.DownPot = msg.PotType(pot)
							action.IsAction = v.IsAction
							action.Account = v.Account
							r.BroadCastMsg(action)

							// 广播房间更新注池金额
							potChange := &msg.PotChangeMoney_S2C{}
							potChange.PotMoneyCount = new(msg.DownBetMoney)
							potChange.PotMoneyCount.BigDownBet = r.PotMoneyCount.BigDownBet
							potChange.PotMoneyCount.SmallDownBet = r.PotMoneyCount.SmallDownBet
							potChange.PotMoneyCount.SingleDownBet = r.PotMoneyCount.SingleDownBet
							potChange.PotMoneyCount.DoubleDownBet = r.PotMoneyCount.DoubleDownBet
							potChange.PotMoneyCount.PairDownBet = r.PotMoneyCount.PairDownBet
							potChange.PotMoneyCount.StraightDownBet = r.PotMoneyCount.StraightDownBet
							potChange.PotMoneyCount.LeopardDownBet = r.PotMoneyCount.LeopardDownBet
							r.BroadCastMsg(potChange)
							//log.Debug("机器Id: %v,下注: %v", v.Id, v.DownBetMoney)
						}
					} else {
						InsertRobotData(rData) //todo
						return
					}
				}
			}
		}
	}()
}

//RobotRandPot 随机机器下注注池
func RobotRandPot() int32 {
	num := RandInRange(1, 10001)
	var pot int32
	if num >= 1 && num <= 2074 {
		pot = 1
	} else if num >= 2075 && num <= 4148 {
		pot = 2
	} else if num >= 4149 && num <= 6222 {
		pot = 3
	} else if num >= 6223 && num <= 8296 {
		pot = 4
	} else if num >= 8297 && num <= 9679 {
		pot = 5
	} else if num >= 9680 && num <= 9938 {
		pot = 6
	} else if num >= 9939 && num <= 10000 {
		pot = 7
	}
	return pot
}

//RandNumber 随机机器下注金额
func RobotRandBet() int32 {
	num := RandInRange(1, 1001)
	var bet int32
	if num >= 1 && num <= 480 {
		bet = 1
	} else if num >= 481 && num <= 695 {
		bet = 5
	} else if num >= 696 && num <= 847 {
		bet = 10
	} else if num >= 848 && num <= 915 {
		bet = 50
	} else if num >= 916 && num <= 963 {
		bet = 100
	} else if num >= 964 && num <= 985 {
		bet = 500
	} else if num >= 986 && num <= 1000 {
		bet = 1000
	}
	return bet
}

//Start 机器人开工~！
func (rc *RobotsCenter) Start() {
	rand.Seed(time.Now().UnixNano())
	num := RandInRange(15, 25)
	hall.LoadHallRobots(num)
}

//生成随机机器人ID
func RandomID() string {
	for {
		RobotId := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(800000000))
		if RobotId[0:1] != "0" {
			return RobotId
		}
	}
}

//生成随机机器人头像IMG
func RandomIMG() string {
	slice := []string{
		"1.png", "2.png", "3.png", "4.png", "5.png", "6.png", "7.png", "8.png", "9.png", "10.png",
		"11.png", "12.png", "13.png", "14.png", "15.png", "16.png", "17.png", "18.png", "19.png", "20.png",
	}
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(slice))

	return slice[num]
}

//生成随机机器人NickName
func RandomName() string {
	for {
		randNum := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(800000000))
		if randNum[0:1] != "0" {
			return randNum
		}
	}
}

func RandomAccount() float64 {
	rand.Intn(int(time.Now().Unix()))
	money := RandInRange(200, 5000)
	return float64(money)
}

func RandomBankerAccount() float64 {
	rand.Intn(int(time.Now().Unix()))
	money := RandInRange(2000, 20000)
	return float64(money)
}
