package internal

import (
	"caidaxiao/conf"
	"caidaxiao/msg"
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RoomStatus int32

const (
	RoomStatusNone RoomStatus = 1 // 房间等待状态
	RoomStatusRun  RoomStatus = 2 // 房间运行状态
	RoomStatusOver RoomStatus = 3 // 房间结束状态
)

const (
	BankerTime  = 8  // 庄家时间 8秒
	Banker2Time = 3  // 庄家连庄 3秒
	DownBetTime = 23 // 下注时间 23秒
	SettleTime  = 29 // 结算时间 29秒
)

const (
	WinBig      int32 = 1  //大 1倍
	WinSmall    int32 = 1  //小 1倍
	WinSingle   int32 = 1  //单 1倍
	WinDouble   int32 = 1  //双 1倍
	WinPair     int32 = 2  //对 2倍
	WinStraight int32 = 6  //顺 6倍
	WinLeopard  int32 = 15 //豹 15倍
)

const (
	taxRate float64 = 0.05 //税率
)

// 游戏阶段channel
var BankerChannel chan bool
var DownBetChannel chan bool

type Room struct {
	RoomId      string    // 房间号
	PlayerList  []*Player // 玩家列表
	TablePlayer []*Player // 玩家列表

	BankerId    string           // 庄家ID
	BankerMoney float64          // 庄家金额
	bankerList  map[string]int32 // 抢庄列表
	IsConBanker bool             // 是否继续连庄

	resultTime       string            // 结算时间
	Lottery          []int             // 开奖数据
	LotteryResult    msg.PotWinList    // 开奖结果
	PeriodsNum       string            // 开奖期数
	RoomStat         RoomStatus        // 房间状态
	GameStat         msg.GameStep      // 游戏状态
	PotMoneyCount    msg.DownBetMoney  // 注池下注总金额(用于客户端显示)
	PlayerTotalMoney msg.DownBetMoney  // 所有真实玩家注池下注(用于计算金额)
	PotWinList       []msg.PotWinList  // 游戏开奖记录
	HistoryData      []msg.HistoryData // 历史开奖数据
	counter          int32             // 已经过去多少秒
	clock            *time.Ticker      // 计时器

	UserLeave []string // 用户是否在房间
}

func (r *Room) Init() {
	//roomId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	//r.RoomId = roomId
	r.PlayerList = nil
	r.TablePlayer = nil

	r.BankerId = ""
	r.BankerMoney = 0
	r.bankerList = make(map[string]int32)
	r.IsConBanker = false

	r.resultTime = ""
	r.Lottery = nil
	r.LotteryResult = msg.PotWinList{}
	r.PeriodsNum = ""
	r.RoomStat = RoomStatusNone
	r.GameStat = msg.GameStep_XX_Step
	r.PlayerTotalMoney = msg.DownBetMoney{}
	r.PotMoneyCount = msg.DownBetMoney{}
	r.PotWinList = make([]msg.PotWinList, 0)
	r.HistoryData = make([]msg.HistoryData, 0)

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	r.UserLeave = make([]string, 0)

	BankerChannel = make(chan bool)
	DownBetChannel = make(chan bool)
}

//BroadCastExcept 向当前玩家之外的玩家广播
func (r *Room) BroadCastExcept(msg interface{}, p *Player) {
	for _, v := range r.PlayerList {
		if v != nil && v.Id != p.Id {
			v.SendMsg(msg)
		}
	}
}

//BroadCastMsg 进行广播消息
func (r *Room) BroadCastMsg(msg interface{}) {
	for _, v := range r.PlayerList {
		if v != nil {
			v.SendMsg(msg)
		}
	}
}

//PlayerLen 房间当前人数
func (r *Room) PlayerLength() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil {
			num++
		}
	}
	return num
}

//PlayerLen 房间当前真实玩家人数
func (r *Room) PlayerTrueLength() int32 {
	var num int32
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == false {
			num++
		}
	}
	return num
}

//PotTotalMoney 房间真实玩家总下注和
func (r *Room) PotTotalMoney() int32 {
	totalMoney := r.PlayerTotalMoney.BigDownBet + r.PlayerTotalMoney.SmallDownBet +
		r.PlayerTotalMoney.SingleDownBet + r.PlayerTotalMoney.DoubleDownBet +
		r.PlayerTotalMoney.PairDownBet + r.PlayerTotalMoney.StraightDownBet + r.PlayerTotalMoney.LeopardDownBet
	return totalMoney
}

//LoadRoomRobots 装载机器人
func (r *Room) LoadRoomRobots(num int) {
	log.Debug("房间: %v ----- 装载 %v个机器人", r.RoomId, num)
	for i := 0; i < num; i++ {
		time.Sleep(time.Millisecond)
		robot := gRobotCenter.CreateRobot()
		r.JoinGameRoom(robot)
	}
}

//RespRoomData 返回房间数据
func (r *Room) RespRoomData() *msg.RoomData {
	rd := &msg.RoomData{}
	rd.RoomId = r.RoomId
	rd.GameTime = r.counter
	rd.GameStep = r.GameStat
	for _, v := range r.Lottery {
		rd.ResultInt = append(rd.ResultInt, int32(v))
	}
	rd.PotMoneyCount = new(msg.DownBetMoney)
	rd.PotMoneyCount.BigDownBet = r.PotMoneyCount.BigDownBet
	rd.PotMoneyCount.SmallDownBet = r.PotMoneyCount.SmallDownBet
	rd.PotMoneyCount.SingleDownBet = r.PotMoneyCount.SingleDownBet
	rd.PotMoneyCount.DoubleDownBet = r.PotMoneyCount.DoubleDownBet
	rd.PotMoneyCount.PairDownBet = r.PotMoneyCount.PairDownBet
	rd.PotMoneyCount.StraightDownBet = r.PotMoneyCount.StraightDownBet
	rd.PotMoneyCount.LeopardDownBet = r.PotMoneyCount.LeopardDownBet
	for _, v := range r.PotWinList {
		pot := &msg.PotWinList{}
		pot.CardType = v.CardType
		pot.BigSmall = v.BigSmall
		pot.SinDouble = v.SinDouble
		pot.ResultNum = v.ResultNum
		rd.PotWinList = append(rd.PotWinList, pot)
	}
	for _, v := range r.HistoryData {
		his := &msg.HistoryData{}
		his.TimeFmt = v.TimeFmt
		his.ResNum = v.ResNum
		his.Result = v.Result
		his.BigSmall = v.BigSmall
		his.SinDouble = v.SinDouble
		his.CardType = v.CardType
		rd.HistoryData = append(rd.HistoryData, his)
	}
	// 这里只需要遍历桌面玩家，站起玩家不显示出来
	for _, v := range r.PlayerList {
		if v != nil {
			pd := &msg.PlayerData{}
			pd.PlayerInfo = new(msg.PlayerInfo)
			pd.PlayerInfo.Id = v.Id
			pd.PlayerInfo.NickName = v.NickName
			pd.PlayerInfo.HeadImg = v.HeadImg
			pd.PlayerInfo.Account = v.Account
			pd.BankerMoney = v.BankerMoney
			pd.BankerCount = v.BankerCount
			pd.DownBetMoney = new(msg.DownBetMoney)
			pd.DownBetMoney.BigDownBet = v.DownBetMoney.BigDownBet
			pd.DownBetMoney.SmallDownBet = v.DownBetMoney.SmallDownBet
			pd.DownBetMoney.SingleDownBet = v.DownBetMoney.SingleDownBet
			pd.DownBetMoney.DoubleDownBet = v.DownBetMoney.DoubleDownBet
			pd.DownBetMoney.PairDownBet = v.DownBetMoney.PairDownBet
			pd.DownBetMoney.StraightDownBet = v.DownBetMoney.StraightDownBet
			pd.DownBetMoney.LeopardDownBet = v.DownBetMoney.LeopardDownBet
			for _, v := range v.DownBetHistory {
				his := &msg.DownBetHistory{}
				his.TimeFmt = v.TimeFmt
				his.ResNum = v.ResNum
				his.Result = v.Result
				his.BigSmall = v.BigSmall
				his.SinDouble = v.SinDouble
				his.CardType = v.CardType
				his.DownBetMoney = v.DownBetMoney
				pd.DownBetHistory = append(pd.DownBetHistory, his)
			}
			pd.TotalDownBet = v.TotalDownBet
			pd.WinTotalCount = v.WinTotalCount
			pd.ResultMoney = v.ResultMoney
			pd.IsAction = v.IsAction
			pd.IsBanker = v.IsBanker
			pd.IsRobot = v.IsRobot
			rd.PlayerData = append(rd.PlayerData, pd)
		}
	}
	for _, v := range r.TablePlayer {
		if v != nil {
			pd := &msg.PlayerData{}
			pd.PlayerInfo = new(msg.PlayerInfo)
			pd.PlayerInfo.Id = v.Id
			pd.PlayerInfo.NickName = v.NickName
			pd.PlayerInfo.HeadImg = v.HeadImg
			pd.PlayerInfo.Account = v.Account
			pd.BankerMoney = v.BankerMoney
			pd.BankerCount = v.BankerCount
			pd.DownBetMoney = new(msg.DownBetMoney)
			pd.DownBetMoney.BigDownBet = v.DownBetMoney.BigDownBet
			pd.DownBetMoney.SmallDownBet = v.DownBetMoney.SmallDownBet
			pd.DownBetMoney.SingleDownBet = v.DownBetMoney.SingleDownBet
			pd.DownBetMoney.DoubleDownBet = v.DownBetMoney.DoubleDownBet
			pd.DownBetMoney.PairDownBet = v.DownBetMoney.PairDownBet
			pd.DownBetMoney.StraightDownBet = v.DownBetMoney.StraightDownBet
			pd.DownBetMoney.LeopardDownBet = v.DownBetMoney.LeopardDownBet
			for _, v := range v.DownBetHistory {
				his := &msg.DownBetHistory{}
				his.TimeFmt = v.TimeFmt
				his.ResNum = v.ResNum
				his.Result = v.Result
				his.BigSmall = v.BigSmall
				his.SinDouble = v.SinDouble
				his.CardType = v.CardType
				his.DownBetMoney = v.DownBetMoney
				pd.DownBetHistory = append(pd.DownBetHistory, his)
			}
			pd.TotalDownBet = v.TotalDownBet
			pd.WinTotalCount = v.WinTotalCount
			pd.ResultMoney = v.ResultMoney
			pd.IsAction = v.IsAction
			pd.IsBanker = v.IsBanker
			pd.IsRobot = v.IsRobot
			rd.TablePlayer = append(rd.TablePlayer, pd)
		}
	}
	return rd
}

//RespRoomData 返回房间数据
func (r *Room) RespUptPlayerList() []*msg.PlayerData {
	var playerSlice []*msg.PlayerData
	for _, v := range r.PlayerList {
		if v != nil {
			pd := &msg.PlayerData{}
			pd.PlayerInfo = new(msg.PlayerInfo)
			pd.PlayerInfo.Id = v.Id
			pd.PlayerInfo.NickName = v.NickName
			pd.PlayerInfo.HeadImg = v.HeadImg
			pd.PlayerInfo.Account = v.Account
			pd.DownBetMoney = new(msg.DownBetMoney)
			pd.DownBetMoney.BigDownBet = v.DownBetMoney.BigDownBet
			pd.DownBetMoney.SmallDownBet = v.DownBetMoney.SmallDownBet
			pd.DownBetMoney.SingleDownBet = v.DownBetMoney.SingleDownBet
			pd.DownBetMoney.DoubleDownBet = v.DownBetMoney.DoubleDownBet
			pd.DownBetMoney.PairDownBet = v.DownBetMoney.PairDownBet
			pd.DownBetMoney.StraightDownBet = v.DownBetMoney.StraightDownBet
			pd.DownBetMoney.LeopardDownBet = v.DownBetMoney.LeopardDownBet
			pd.TotalDownBet = v.TotalDownBet
			pd.WinTotalCount = v.WinTotalCount
			pd.ResultMoney = v.ResultMoney
			pd.IsAction = v.IsAction
			pd.IsBanker = v.IsBanker
			playerSlice = append(playerSlice, pd)
		}
	}
	return playerSlice
}

//PlayerListSort 玩家列表排序(进入房间、退出房间、重新开始)
func (r *Room) UpdatePlayerList() {
	// 首先
	// 临时切片
	var playerSlice []*Player

	//1、玩家下注总金额
	var p1 []*Player //所有下注过的用户
	var p2 []*Player //所有下注金额为0的用户
	for _, v := range r.PlayerList {
		if v != nil {
			if v.TotalDownBet != 0 {
				p1 = append(p1, v)
			} else {
				p2 = append(p2, v)
			}
		}
	}
	//2、根据玩家总下注进行排序
	for i := 0; i < len(p1); i++ {
		for j := 1; j < len(p1)-i; j++ {
			if p1[j].TotalDownBet > p1[j-1].TotalDownBet {
				//交换
				p1[j], p1[j-1] = p1[j-1], p1[j]
			}
		}
	}
	// 将用户总下注金额顺序追加到临时切片
	playerSlice = append(playerSlice, p1...)
	//3、玩家金额,总下注为0,按用户金额排序
	for i := 0; i < len(p2); i++ {
		for j := 1; j < len(p2)-i; j++ {
			if p2[j].Account > p2[j-1].Account {
				//交换
				p2[j], p2[j-1] = p2[j-1], p2[j]
			}
		}
	}
	// 将用户余额排序追加到临时切片
	playerSlice = append(playerSlice, p2...)

	// 将房间列表置为空,将更新的数据追加到房间列表
	r.PlayerList = nil
	r.PlayerList = append(r.PlayerList, playerSlice...)
}

//GetCaiYuan 获取彩源开奖结果
func (r *Room) GetCaiYuan() {
	var dataRes *http.Response
	if r.RoomId == "1" {
		//caiYuan := "http://free.manycai.com/K2601968389c853/hn60-1.json"
		res, err := http.Get(conf.Server.CaiYuan)
		if err != nil {
			log.Debug("再次获取随机数值失败: %v", err)
			return
		}
		dataRes = res
	} else if r.RoomId == "2" {
		caiYuan := "http://free.manycai.com/K2601968389c853/PTXFFC-1.json"
		res, err := http.Get(caiYuan)
		if err != nil {
			log.Debug("再次获取随机数值失败: %v", err)
			return
		}
		dataRes = res
	}

	log.Debug("res:%v", dataRes)
	result, err := ioutil.ReadAll(dataRes.Body)
	defer dataRes.Body.Close()
	if err != nil {
		log.Error("解析随机数值失败: %v", err)
		return
	}

	var users interface{}
	err2 := json.Unmarshal(result, &users)
	if err2 != nil {
		log.Error("解码随机数值失败: %v", err)
		return
	}

	log.Debug("读取的彩源数据: %v", users)

	data, ok := users.([]interface{})
	if ok {
		for _, v := range data {
			lottery := v.(map[string]interface{})
			opendate := lottery["opendate"] // 开奖时间
			log.Debug("开奖时间:%v", opendate)
			issue := lottery["issue"] // 彩票期数
			log.Debug("彩票期数:%v", issue)
			lotterycode := lottery["lotterycode"] // 彩票代码
			log.Debug("彩票代码:%v", lotterycode)
			code := lottery["code"] // 中奖号码
			log.Debug("中奖号码:%v", code)

			r.resultTime = opendate.(string)
			r.PeriodsNum = issue.(string)
			codeString := code.(string)
			codeSlice := strings.Split(codeString, `,`)
			codeSlice = append(codeSlice[:0], codeSlice[2:]...)
			var codeData []int
			for _, v := range codeSlice {
				num, _ := strconv.Atoi(v)
				codeData = append(codeData, num)
			}
			r.Lottery = codeData
		}
	}
}

//CleanRoomData 清空房间数据,开始下一句游戏
func (r *Room) CleanRoomData() {
	r.TablePlayer = nil
	r.Lottery = nil
	r.LotteryResult = msg.PotWinList{}
	r.RoomStat = RoomStatusOver
	r.GameStat = msg.GameStep_XX_Step
	r.PlayerTotalMoney = msg.DownBetMoney{}
	r.PotMoneyCount = msg.DownBetMoney{}
	r.counter = 0
	r.UserLeave = []string{}
	r.bankerList = make(map[string]int32)
	// 清空玩家数据
	r.CleanPlayerData()
}

//CleanPlayerData 清空玩家数据,开始下一句游戏
func (r *Room) CleanPlayerData() {
	for _, v := range r.PlayerList {
		if v != nil {
			v.DownBetMoney = msg.DownBetMoney{}
			v.IsAction = false
			v.WinResultMoney = 0
			v.LoseResultMoney = 0
			v.ResultMoney = 0
			v.IsDownBanker = false
		}
	}
}

//CleanRobot 清理机器金额小于100的
func (r *Room) CleanRobot() {
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			if v.Account < 100 {
				r.ExitFromRoom(v)
			}
		}
	}
}

//KickOutPlayer 踢出房间断线玩家
func (r *Room) KickOutPlayer() {
	// 清理断线玩家
	for _, uid := range r.UserLeave {
		for _, v := range r.PlayerList {
			if v != nil && v.Id == uid {
				// 玩家断线的话，退出房间信息，也要断开链接
				if v.IsOnline == true {
					v.PlayerExitRoom()
				} else {
					if v.IsBanker == true {
						r.IsConBanker = false
						nowTime := time.Now().Unix()
						v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
						reason := "庄家申请下庄"
						c4c.BankerStatus(v, 0, nowTime, v.RoundId, reason)
					}
					v.PlayerExitRoom()
					hall.UserRecord.Delete(v.Id)
					c4c.UserLogoutCenter(v.Id, v.Password, v.Token) //, p.PassWord
					leaveHall := &msg.Logout_S2C{}
					v.SendMsg(leaveHall)
					v.IsOnline = false
					log.Debug("踢出房间断线玩家 : %v", v.Id)
				}
			}
		}
	}
}

//ExitFromRoom 从房间退出处理
func (r *Room) ExitFromRoom(p *Player) {

	//清空用户数据
	p.Status = msg.PlayerStatus_XX_Status
	p.BankerMoney = 0
	p.BankerCount = 0
	p.BankerStatus = 0
	p.DownBetMoney = msg.DownBetMoney{}
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.TotalDownBet = 0
	p.WinTotalCount = 0
	p.TwentyData = nil
	p.DownBetHistory = make([]msg.DownBetHistory, 0)
	p.IsBanker = false
	p.IsDownBanker = false
	p.IsAction = false

	//从房间列表删除玩家信息,更新房间列表
	for k, v := range r.PlayerList {
		if v != nil && v.Id == p.Id {
			if v.IsRobot == false {
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...) //这里两个同样的用户名退出，会报错
				log.Debug("%v 玩家从房间列表删除成功 ~", v.Id)
			} else {
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...)
				//log.Debug("%v 机器从房间列表删除成功 ~", v.Id)
			}
		}
	}

	// 发送退出房间
	leave := &msg.LeaveRoom_S2C{}
	leave.PlayerInfo = new(msg.PlayerInfo)
	leave.PlayerInfo.Id = p.Id
	leave.PlayerInfo.NickName = p.NickName
	leave.PlayerInfo.HeadImg = p.HeadImg
	leave.PlayerInfo.Account = p.Account
	p.SendMsg(leave)

	// 玩家列表更新
	uptPlayerList := &msg.UptPlayerList_S2C{}
	uptPlayerList.PlayerList = r.RespUptPlayerList()
	r.BroadCastMsg(uptPlayerList)

	delete(hall.UserRoom, p.Id)
}

func (r *Room) HandleBanker() {
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == false && v.IsBanker == true {
			// 如果玩家点击下庄就处理庄家下庄
			if v.IsDownBanker == true {
				v.BankerMoney = 0
				v.BankerCount = 0
				v.BankerStatus = 0
				v.IsBanker = false
				v.IsDownBanker = false
				r.IsConBanker = false
				nowTime := time.Now().Unix()
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				reason := "庄家申请下庄"
				c4c.BankerStatus(v, 0, nowTime, v.RoundId, reason)
			}
			// 判断庄家金额是否小于2000或者庄家连庄3次以上就下庄
			if v.BankerMoney < 2000 || v.BankerCount >= 3 {
				v.BankerMoney = 0
				v.BankerCount = 0
				v.BankerStatus = 0
				v.IsBanker = false
				v.IsDownBanker = false
				r.IsConBanker = false
				nowTime := time.Now().Unix()
				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
				reason := "庄家申请下庄"
				c4c.BankerStatus(v, 0, nowTime, v.RoundId, reason)
			}
		}
	}
	// 清空机器庄家
	r.ClearRobotBanker()
}

//HandleRobot 处理机器人
func (r *Room) HandleRobot() {
	timeNow := time.Now().Hour()
	var handleNum int
	var nextNum int
	switch timeNow {
	case 1:
		handleNum = 75
		nextNum = 68
		break
	case 2:
		handleNum = 68
		nextNum = 60
		break
	case 3:
		handleNum = 60
		nextNum = 51
		break
	case 4:
		handleNum = 51
		nextNum = 41
		break
	case 5:
		handleNum = 41
		nextNum = 30
		break
	case 6:
		handleNum = 30
		nextNum = 17
		break
	case 7:
		handleNum = 17
		nextNum = 15
		break
	case 8:
		handleNum = 15
		nextNum = 17
		break
	case 9:
		handleNum = 17
		nextNum = 30
		break
	case 10:
		handleNum = 30
		nextNum = 41
		break
	case 11:
		handleNum = 41
		nextNum = 51
		break
	case 12:
		handleNum = 51
		nextNum = 60
		break
	case 13:
		handleNum = 60
		nextNum = 68
		break
	case 14:
		handleNum = 68
		nextNum = 75
		break
	case 15:
		handleNum = 75
		nextNum = 80
		break
	case 16:
		handleNum = 80
		nextNum = 84
		break
	case 17:
		handleNum = 84
		nextNum = 87
		break
	case 18:
		handleNum = 87
		nextNum = 89
		break
	case 19:
		handleNum = 89
		nextNum = 90
		break
	case 20:
		handleNum = 90
		nextNum = 89
		break
	case 21:
		handleNum = 89
		nextNum = 87
		break
	case 22:
		handleNum = 87
		nextNum = 84
		break
	case 23:
		handleNum = 84
		nextNum = 80
		break
	case 0:
		handleNum = 80
		nextNum = 75
		break
	}

	t2 := time.Now().Minute()
	m1 := float64(t2) / 60
	m2, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", m1), 64)
	if handleNum > nextNum { // -
		m3 := handleNum - nextNum
		m4 := m2 * float64(m3)
		m5 := math.Floor(m4)
		handleNum -= int(m5)
	} else if handleNum < nextNum { // +
		m3 := nextNum - handleNum
		m4 := m2 * float64(m3)
		m5 := math.Floor(m4)
		handleNum += int(m5)
	}

	var minP int
	var maxP int
	getNum := float64(handleNum) * 0.2
	maNum := math.Floor(getNum)
	minP = handleNum - int(maNum)
	maxP = handleNum + int(maNum)

	num := RandInRange(0, 100)
	if num >= 0 && num < 50 {
		num2 := handleNum - minP
		num3 := RandInRange(0, num2)
		handleNum += num3
	} else if num >= 50 && num < 100 {
		num2 := maxP - handleNum
		num3 := RandInRange(0, num2)
		handleNum -= num3
	}

	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			rNum := 1 / float64((v.WinTotalCount+1)*2)
			//log.Debug("rNum:%v", rNum)
			rNum2 := int(rNum * 1000)
			rNum3 := RandInRange(0, 1000)
			//log.Debug("rNum2:%v,rNum3:%v", rNum2, rNum3)
			if rNum3 <= rNum2 {
				r.ExitFromRoom(v)
				time.Sleep(time.Millisecond * 10)
			}
		}
	}

	robotNum := r.RobotLength()
	log.Debug("机器人当前数量:%v,handleNum当局指定人数:%v", robotNum, handleNum)
	if robotNum < handleNum { // 加
		for {
			robot := gRobotCenter.CreateRobot()
			r.JoinGameRoom(robot)
			time.Sleep(time.Millisecond * 10)
			robotNum = r.RobotLength()
			if robotNum == handleNum {
				log.Debug("房间:%v,加机器人数量:%v", r.RoomId, r.RobotLength())
				break
			}
		}
	} else if robotNum > handleNum { // 减
		for _, v := range r.PlayerList {
			if v != nil && v.IsRobot == true {
				r.ExitFromRoom(v)
				time.Sleep(time.Millisecond * 10)
				robotNum = r.RobotLength()
				if robotNum == handleNum {
					log.Debug("房间:%v,减机器人数量:%v", r.RoomId, r.RobotLength())
					break
				}
			}
		}
	}
}

func (r *Room) RobotLength() int {
	var num int
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			num++
		}
	}
	return num
}

func (r *Room) PlayerUpBanker() {
	if len(r.bankerList) == 0 {
		p := gRobotCenter.CreateRobot()
		p.Account = RandomBankerAccount()
		p.IsBanker = true
		r.BankerId = p.Id
		p.BankerCount++
		hall.UserRoom[p.Id] = r.RoomId
		r.PlayerList = append(r.PlayerList, p)

		bankerMoney := []int32{2000, 5000, 10000, 20000}
		num := RandInRange(0, 4)

		p.BankerMoney = float64(bankerMoney[num])
		r.BankerMoney = float64(bankerMoney[num])
		log.Debug("机器人当庄:%v,当庄金额:%v", p.Id, p.BankerMoney)

		data := &msg.BankerData_S2C{}
		data.Banker = p.RespPlayerData()
		data.TakeMoney = bankerMoney[num]
		r.BroadCastMsg(data)
	} else if len(r.bankerList) >= 1 {
		b2000 := make([]string, 0)
		b5000 := make([]string, 0)
		b10000 := make([]string, 0)
		b20000 := make([]string, 0)
		for k := range r.bankerList {
			if r.bankerList[k] == 2000 {
				b2000 = append(b2000, k)
			}
			if r.bankerList[k] == 5000 {
				b5000 = append(b5000, k)
			}
			if r.bankerList[k] == 10000 {
				b10000 = append(b10000, k)
			}
			if r.bankerList[k] == 20000 {
				b20000 = append(b20000, k)
			}
		}
		if len(b20000) >= 1 {
			if len(b20000) == 1 {
				r.SetBanker(b20000[0], 20000)
			} else {
				num := RandInRange(0, len(b20000))
				r.SetBanker(b20000[num], 20000)
			}
			return
		}
		if len(b10000) >= 1 {
			if len(b10000) == 1 {
				r.SetBanker(b10000[0], 10000)
			} else {
				num := RandInRange(0, len(b10000))
				r.SetBanker(b10000[num], 10000)
			}
			return
		}
		if len(b5000) >= 1 {
			if len(b5000) == 1 {
				r.SetBanker(b5000[0], 5000)
			} else {
				num := RandInRange(0, len(b5000))
				r.SetBanker(b5000[num], 5000)
			}
			return
		}
		if len(b2000) >= 1 {
			if len(b2000) == 1 {
				r.SetBanker(b2000[0], 2000)
			} else {
				num := RandInRange(0, len(b2000))
				r.SetBanker(b2000[num], 2000)
			}
			return
		}
	}
}

// 设定庄家并发送数据
func (r *Room) SetBanker(id string, takeMoney int32) {
	for _, v := range r.PlayerList {
		if v != nil && v.Id == id {
			v.IsBanker = true
			v.BankerMoney = float64(takeMoney)
			v.BankerCount++
			v.BankerStatus = msg.BankerStatus_BankerUp
			r.BankerId = id
			r.BankerMoney = float64(takeMoney)
			r.IsConBanker = true
			nowTime := time.Now().Unix()
			v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
			reason := "庄家申请上庄"
			c4c.BankerStatus(v, 1, nowTime, v.RoundId, reason)
			log.Debug("玩家当庄:%v,当庄金额:%v", v.Id, v.BankerMoney)

			data := &msg.BankerData_S2C{}
			data.Banker = v.RespPlayerData()
			data.TakeMoney = takeMoney
			r.BroadCastMsg(data)
		}
	}
}

// 清除机器庄家
func (r *Room) ClearRobotBanker() {
	for _, v := range r.PlayerList {
		if v != nil {
			if v.IsRobot == true && v.IsBanker == true {
				v.IsBanker = false
				v.BankerMoney = 0
				v.BankerCount = 0
				v.BankerStatus = 0
			}
		}
	}
}
