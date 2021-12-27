package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type PrizeRecord struct {
	Issue    string `json:"issue"`    // 奖期号
	Code     string `json:"code"`     // 开奖号码
	OpenDate string `json:"opendate"` // 开奖时间
}

const (
	DownBetTime = 30 // 下注總时间 30秒
	CloseTime   = 10 // 封单總时间 10秒
	GetResTime  = 14 // 开奖總时间 14秒
	SettleTime  = 6  // 结算總时间 6秒

)

const ( //切換狀態時間點
	DownBetStep = 25 // 下注时间點
	CloseStep   = 55 // 封单时间點
	GetResStep  = 5  // 开奖时间點
	SettleStep  = 19 // 结算时间點
)

const (
	WinBig     int32 = 1  //大 1倍
	WinSmall   int32 = 1  //小 1倍
	WinLeopard int32 = 66 //豹 66倍
)

const (
	taxRate float64 = 0.05 //税率
)

const (
	QIQFFC = "https://manycai.com/K2601968389c853/QIQFFC-1.json"
	HNFFC  = "https://manycai.com/K2601968389c853/hn60-1.json"
)

var (
	packageTax  map[uint16]float64
	keyReqPrize sync.Mutex //取號鎖
)

type Room struct {
	RoomId      string    // 房间号
	PlayerList  []*Player // 玩家列表
	TablePlayer []*Player // 玩家列表

	PackageId     uint16
	GodGambleName int32            // 赌神id
	BankerId      int32            // 庄家ID
	BankerMoney   float64          // 庄家金额
	bankerList    map[string]int32 // 抢庄列表
	IsConBanker   bool             // 是否继续连庄

	resultTime       string            // 结算时间
	Lottery          []int             // 开奖数据
	LotteryResult    msg.LotteryData   // 开奖结果
	PeriodsNum       string            // 开奖期数
	ResultNum        string            // 期数
	RoundID          string            // 房間回合
	PeriodsTime      string            // 开奖时间
	GameStat         msg.GameStep      // 游戏状态
	PotMoneyCount    msg.DownBetMoney  // 注池下注总金额(用于客户端显示)
	PlayerTotalMoney *msg.DownBetMoney // 所有真实玩家注池下注(用于计算金额)
	// PotWinList       []*msg.PotWinList  // [packageid]游戏开奖记录
	HistoryData []msg.LotteryData // 历史开奖数据
	counter     int32             // 已经过去多少秒
	clock       *time.Ticker      // 计时器
	UserLeave   []int32           // 用户是否在房间
	IsOpenRoom  bool              // 是否开启房间

	RoomMinBet int32 // 房间限定下注最小金额
	RoomMaxBet int32 // 房间限定下注最大金额

	userRoomMutex   sync.RWMutex // 玩家進出鎖
	userBetMutex    sync.RWMutex // 玩家下注鎖
	SettlementMutex sync.RWMutex // 玩家結算鎖
}

func (r *Room) Init() {
	//roomId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	//r.RoomId = roomId
	r.PlayerList = nil
	r.TablePlayer = nil

	r.GodGambleName = 0
	r.BankerId = 0
	r.BankerMoney = 0
	r.bankerList = make(map[string]int32)
	r.IsConBanker = false

	r.resultTime = ""
	r.Lottery = nil
	r.LotteryResult = msg.LotteryData{}
	r.LotteryResult.Result = &msg.LotteryResult{}
	r.LotteryResult.ResultFX = &msg.LotteryResultFX{}
	r.PeriodsNum = ""
	r.PeriodsTime = ""
	r.GameStat = msg.GameStep_XX_Step
	r.PlayerTotalMoney = &msg.DownBetMoney{}
	r.PotMoneyCount = msg.DownBetMoney{}
	// r.PotWinList = make([]*msg.PotWinList, 0)
	r.HistoryData = make([]msg.LotteryData, 0)

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	r.UserLeave = make([]int32, 0)
	r.IsOpenRoom = true
	r.RoomMinBet = 1
	r.RoomMaxBet = 5000

	roomidCount := SearchCMD{
		DBName: dbName,
		CName:  RoomStatusDB,
		Query:  bson.M{"room_id": r.RoomId},
	}

	if FindCountByQuery(roomidCount) > 0 {
		roomidLoad := SearchCMD{
			DBName: dbName,
			CName:  RoomStatusDB,
			Query:  bson.M{"room_id": r.RoomId},
		}
		RoomStatus := &RoomStatus{}
		if FindOneItem(roomidLoad, RoomStatus) {
			r.IsOpenRoom = RoomStatus.IsOpen
			r.RoomMinBet = RoomStatus.MinBet
			r.RoomMaxBet = RoomStatus.MaxBet
		}
	}

}

//BroadCastExcept 向当前玩家之外的玩家广播
// func (r *Room) BroadCastExcept(msg interface{}, p *Player) {
// 	for _, v := range r.PlayerList {
// 		if v != nil && v.Id != p.Id {
// 			v.SendMsg(msg, "沒在用")
// 		}
// 	}
// }

//BroadCastMsg 进行广播消息
func (r *Room) BroadCastMsg(msg interface{}, event string) {
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == false {
			v.SendMsg(msg, event)
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
	rd.RoundId = r.RoundID
	// rd.GameTime = r.counter //舊的
	rd.GameTime = r.RoomCounter()
	rd.GameStep = r.GameStat
	for _, v := range r.Lottery {
		rd.ResultInt = append(rd.ResultInt, int32(v))
	}
	rd.PeriodsNum = r.PeriodsNum
	rd.PotMoneyCount = new(msg.DownBetMoney)
	rd.PotMoneyCount.BigDownBet = r.PotMoneyCount.BigDownBet
	rd.PotMoneyCount.SmallDownBet = r.PotMoneyCount.SmallDownBet
	rd.PotMoneyCount.SingleDownBet = r.PotMoneyCount.SingleDownBet
	rd.PotMoneyCount.DoubleDownBet = r.PotMoneyCount.DoubleDownBet
	rd.PotMoneyCount.PairDownBet = r.PotMoneyCount.PairDownBet
	rd.PotMoneyCount.StraightDownBet = r.PotMoneyCount.StraightDownBet
	rd.PotMoneyCount.LeopardDownBet = r.PotMoneyCount.LeopardDownBet
	// rd.HistoryData = r.HistoryData // 要針對不同品牌處理
	for _, v := range r.HistoryData {
		lotterydata := &msg.LotteryData{
			TimeFmt:  v.TimeFmt,
			ResNum:   v.ResNum,
			Result:   v.Result,
			ResultFX: v.ResultFX,
			IsLiuJu:  v.IsLiuJu,
		}

		rd.HistoryData = append(rd.HistoryData, lotterydata)
	}
	// fmt.Printf("//////////////////////////////////\nRoom：%v RoomData：\n%v\n//////////////////////////////////\n", r.RoomId, rd.HistoryData)

	// if len(rd.PotWinList) != 0 && len(rd.HistoryData) != 0 && len(rd.ResultInt) != 0 {
	// 	var num = 0
	// 	for i := 0; i < 5; i++ {
	// 		if rd.HistoryData[0].ResNum[i] == rd.ResultInt[i] {
	// 			num++
	// 		}
	// 	}
	// 	if num == 5 {
	// 		common.Debug_log("room:%v \n左上 rd.HistoryData：%v \n中間 rd.ResultInt：%v \n第六 rd.PotWinList：%v", r.RoomId, rd.HistoryData[0].ResNum, rd.ResultInt, rd.PotWinList[0].ResultNum)
	// 	} else {
	// 		var wireteString = fmt.Sprintln("room:", r.RoomId, " \n左上 rd.HistoryData：", rd.HistoryData[0].ResNum, " \n中間 rd.ResultInt：", rd.ResultInt, " \n不一樣!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// 		var filename = "./errorlottery.txt"
	// 		var f *os.File
	// 		var err1 error
	// 		if checkFileIsExist(filename) {
	// 			//如果檔案存在
	// 			f, err1 = os.OpenFile(filename, os.O_APPEND, 0666) //開啟檔案
	// 			fmt.Println("檔案存在")
	// 		} else {
	// 			f, err1 = os.Create(filename) //建立檔案
	// 			fmt.Println("檔案不存在")
	// 		}
	// 		defer f.Close()
	// 		n, err1 := io.WriteString(f, wireteString) //寫入檔案(字串)
	// 		if err1 != nil {
	// 			panic(err1)
	// 		}
	// 		fmt.Printf("寫入 %d 個位元組n", n)
	// 	}
	// }

	// 这里只需要遍历桌面玩家，站起玩家不显示出来
	for _, v := range r.PlayerList {
		if v != nil {
			pd := &msg.PlayerData{}
			pd.PlayerInfo = new(msg.PlayerInfo)
			pd.PlayerInfo.Id = common.Int32ToStr(v.Id)
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
			pd.DownBetHistory = v.DownBetHistory
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
			pd.PlayerInfo.Id = common.Int32ToStr(v.Id)
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
			pd.DownBetHistory = v.DownBetHistory
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

//GetGodGableId 获取赌神ID
func (r *Room) GetGodGableId() {
	var GodSlice []*Player
	GodSlice = append(GodSlice, r.PlayerList...)

	var WinCount []*Player
	for _, v := range GodSlice {
		if v != nil && v.WinTotalCount != 0 {
			WinCount = append(WinCount, v)
		}
	}
	if len(WinCount) == 0 {
		//log.Debug("---------- 没有获取到赌神 ~")
		return
	}

	for i := 0; i < len(GodSlice); i++ {
		for j := 1; j < len(GodSlice)-i; j++ {
			if GodSlice[j].TotalDownBet > GodSlice[j-1].TotalDownBet {
				GodSlice[j], GodSlice[j-1] = GodSlice[j-1], GodSlice[j]
			}
		}
	}

	for i := 0; i < len(GodSlice); i++ {
		for j := 1; j < len(GodSlice)-i; j++ {
			if GodSlice[j].WinTotalCount > GodSlice[j-1].WinTotalCount {
				//交换
				GodSlice[j], GodSlice[j-1] = GodSlice[j-1], GodSlice[j]
			}
		}
	}
	r.GodGambleName = GodSlice[0].Id
}

//UpdatePlayerList 玩家列表排序
func (r *Room) UpdatePlayerList() {
	// 获取赌神id
	r.GetGodGableId()

	//首先
	//临时切片
	var playerSlice []*Player
	//1、赌神
	for _, v := range r.PlayerList {
		if v != nil && v.Id == r.GodGambleName {
			playerSlice = append(playerSlice, v)
		}
	}
	//2、玩家下注总金额
	var p1 []*Player //所有下注过的用户
	var p2 []*Player //所有下注金额为0的用户
	for _, v := range r.PlayerList {
		if v != nil && v.Id != r.GodGambleName {
			if v.TotalDownBet != 0 {
				p1 = append(p1, v)
			} else {
				p2 = append(p2, v)
			}
		}
	}
	//根据玩家总下注进行排序
	for i := 0; i < len(p1); i++ {
		for j := 1; j < len(p1)-i; j++ {
			if p1[j].TotalDownBet > p1[j-1].TotalDownBet {
				//交换
				p1[j], p1[j-1] = p1[j-1], p1[j]
			}
		}
	}
	//将用户总下注金额顺序追加到临时切片
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
	//将用户余额排序追加到临时切片
	playerSlice = append(playerSlice, p2...)

	//将房间列表置为空,将更新的数据追加到房间列表
	r.PlayerList = nil
	r.PlayerList = append(r.PlayerList, playerSlice...)
}

//GetCaiYuan 获取彩源开奖结果
func (r *Room) GetCaiYuan() {

	go func() {
		for {
			time.Sleep(time.Millisecond * 2000)
			ReqAgain := r.CaiYunApi()
			if ReqAgain != true || r.GameStat == msg.GameStep_Settle || r.GameStat == msg.GameStep_LiuJu {
				return
			}

		}
	}()
}

func (r *Room) CaiYunApi() bool {

	var caiYuan string
	if r.RoomId == "1" {
		caiYuan = HNFFC
	} else if r.RoomId == "2" {
		caiYuan = QIQFFC

	}
	keyReqPrize.Lock()
	resp, err := http.Get(caiYuan)

	defer func() {
		keyReqPrize.Unlock()
		if resp != nil && resp.Body != nil {
			errC := resp.Body.Close() //必須調用否則可能產生記憶體洩漏
			if errC != nil {
				log.Debug("close resp err = %v, urlReq = %s", errC, caiYuan)
			}
		}
	}()

	if err != nil {
		log.Debug("Get-err-1 %+v , urlReq = %s", err, caiYuan)
		return true
	}
	resp.Close = true

	cpinfo := make([]*PrizeRecord, 1)
	err = json.NewDecoder(resp.Body).Decode(&cpinfo)
	if err != nil {
		errMap := make(map[string]interface{})
		err2 := json.NewDecoder(resp.Body).Decode(&errMap)
		if err2 == nil && errMap["error"] != nil {
			errMsg, ok := errMap["error"].(string)
			if !ok {
				errMsg = "Unknown Error"
			}
			err2 = errors.New(errMsg)
		}
		log.Debug("reqPrizeSource-err: %v, urlReq = %s", err2, caiYuan)
		// return nil, err2
		return true
	}

	opendate := cpinfo[0].OpenDate // 开奖时间
	issue := cpinfo[0].Issue       // 彩票期数
	code := cpinfo[0].Code         // 中奖号码
	log.Debug("獎源:%v \n开奖时间:%v 彩票期数:%v 中奖号码:%v", caiYuan, opendate, issue, code)

	t := time.Now()
	nMinute := t.Minute()
	m := getMinute(cpinfo[0].OpenDate)

	if nMinute == m { // 判斷是最新的期号
		r.resultTime = opendate
		r.PeriodsTime = opendate
		r.PeriodsNum = issue
		codeString := code
		codeSlice := strings.Split(codeString, `,`)
		//codeSlice = append(codeSlice[:0], codeSlice[2:]...)
		var codeData []int
		for _, v := range codeSlice {
			num, _ := strconv.Atoi(v)
			codeData = append(codeData, num)
		}
		r.Lottery = codeData
		return false
	}
	return true
}

//CleanRoomData 清空房间数据,开始下一句游戏
func (r *Room) CleanRoomData() {
	r.TablePlayer = nil
	r.resultTime = ""
	r.Lottery = nil
	r.LotteryResult = msg.LotteryData{}
	r.LotteryResult.Result = &msg.LotteryResult{}
	r.LotteryResult.ResultFX = &msg.LotteryResultFX{}
	r.PlayerTotalMoney = &msg.DownBetMoney{}
	r.PotMoneyCount = msg.DownBetMoney{}
	r.counter = 0
	r.UserLeave = []int32{}
	r.bankerList = make(map[string]int32)
	// 清空玩家数据
	r.CleanPlayerData()
}

//CleanPlayerData 清空玩家数据,开始下一句游戏
func (r *Room) CleanPlayerData() {
	for _, v := range r.PlayerList {
		if v != nil {
			v.DownBetMoney = &msg.DownBetMoney{}
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
			if v != nil && v.Id == uid && v.IsAction == false {
				// 玩家断线的话，退出房间信息，也要断开链接
				if v.IsOnline == true {
					v.PlayerExitRoom()
				} else {
					v.PlayerExitRoom()
					hall.UserRecord.Delete(v.Id)
					// c4c.UserLogoutCenter(v.Id, v.Password, v.Token) //, p.PassWord
					sendLogout(v.Id) // 登出
					leaveHall := &msg.Logout_S2C{}
					v.SendMsg(leaveHall, "Logout_S2C")
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
	p.DownBetMoney = &msg.DownBetMoney{}
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.TotalDownBet = 0
	p.WinTotalCount = 0
	p.TwentyData = nil
	p.DownBetHistory = make([]*msg.DownBetHistory, 0)
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
	leave.PlayerInfo.Id = common.Int32ToStr(p.Id)
	leave.PlayerInfo.NickName = p.NickName
	leave.PlayerInfo.HeadImg = p.HeadImg
	leave.PlayerInfo.Account = p.Account
	p.SendMsg(leave, "LeaveRoom_S2C")

	r.DeleteUserRoom(p)
}

// func (r *Room) HandleBanker() {
// 	for _, v := range r.PlayerList {
// 		if v != nil && v.IsRobot == false && v.IsBanker == true {
// 			// 如果玩家点击下庄就处理庄家下庄
// 			if v.IsDownBanker == true {
// 				v.BankerMoney = 0
// 				v.BankerCount = 0
// 				v.BankerStatus = 0
// 				v.IsBanker = false
// 				v.IsDownBanker = false
// 				r.IsConBanker = false
// 				nowTime := time.Now().Unix()
// 				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
// 				reason := "庄家申请下庄"
// 				c4c.BankerStatus(v, 0, nowTime, v.RoundId, reason)
// 			}
// 			// 判断庄家金额是否小于2000或者庄家连庄3次以上就下庄
// 			if v.BankerMoney < 2000 || v.BankerCount >= 3 {
// 				v.BankerMoney = 0
// 				v.BankerCount = 0
// 				v.BankerStatus = 0
// 				v.IsBanker = false
// 				v.IsDownBanker = false
// 				r.IsConBanker = false
// 				nowTime := time.Now().Unix()
// 				v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
// 				reason := "庄家申请下庄"
// 				c4c.BankerStatus(v, 0, nowTime, v.RoundId, reason)
// 			}
// 		}
// 	}
// 	// 清空机器庄家
// 	r.ClearRobotBanker()
// }

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
			}
		}
	}

	robotNum := r.RobotLength()
	log.Debug("机器人当前数量:%v,handleNum当局指定人数:%v", robotNum, handleNum)
	if robotNum < handleNum { // 加
		for {
			robot := gRobotCenter.CreateRobot()
			r.JoinGameRoom(robot)
			robotNum = r.RobotLength()
			if robotNum == handleNum {
				log.Debug("房间:%v,現机器人数量:%v ", r.RoomId, r.RobotLength())
				break
			}
		}
	} else if robotNum > handleNum { // 减
		for _, v := range r.PlayerList {
			if v != nil && v.IsRobot == true {
				r.ExitFromRoom(v)
				robotNum = r.RobotLength()
				if robotNum == handleNum {
					log.Debug("房间:%v,現机器人数量:%v", r.RoomId, r.RobotLength())
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

// func (r *Room) PlayerUpBanker() {
// 	if len(r.bankerList) == 0 {
// 		p := gRobotCenter.CreateRobot()
// 		p.Account = RandomBankerAccount()
// 		p.IsBanker = true
// 		r.BankerId = p.Id
// 		p.BankerCount++
// 		hall.UserRoom.Store(p.Id, r.RoomId)
// 		r.PlayerList = append(r.PlayerList, p)

// 		bankerMoney := []int32{2000, 5000, 10000, 20000}
// 		num := RandInRange(0, 4)

// 		p.BankerMoney = float64(bankerMoney[num])
// 		r.BankerMoney = float64(bankerMoney[num])
// 		log.Debug("机器人当庄:%v,当庄金额:%v", p.Id, p.BankerMoney)

// 		data := &msg.BankerData_S2C{}
// 		data.Banker = p.RespPlayerData()
// 		data.TakeMoney = bankerMoney[num]
// 		r.BroadCastMsg(data, "BankerData_S2C")
// 	} else if len(r.bankerList) >= 1 {
// 		b2000 := make([]string, 0)
// 		b5000 := make([]string, 0)
// 		b10000 := make([]string, 0)
// 		b20000 := make([]string, 0)
// 		for k := range r.bankerList {
// 			if r.bankerList[k] == 2000 {
// 				b2000 = append(b2000, k)
// 			}
// 			if r.bankerList[k] == 5000 {
// 				b5000 = append(b5000, k)
// 			}
// 			if r.bankerList[k] == 10000 {
// 				b10000 = append(b10000, k)
// 			}
// 			if r.bankerList[k] == 20000 {
// 				b20000 = append(b20000, k)
// 			}
// 		}
// 		if len(b20000) >= 1 {
// 			if len(b20000) == 1 {
// 				r.SetBanker(b20000[0], 20000)
// 			} else {
// 				num := RandInRange(0, len(b20000))
// 				r.SetBanker(b20000[num], 20000)
// 			}
// 			return
// 		}
// 		if len(b10000) >= 1 {
// 			if len(b10000) == 1 {
// 				r.SetBanker(b10000[0], 10000)
// 			} else {
// 				num := RandInRange(0, len(b10000))
// 				r.SetBanker(b10000[num], 10000)
// 			}
// 			return
// 		}
// 		if len(b5000) >= 1 {
// 			if len(b5000) == 1 {
// 				r.SetBanker(b5000[0], 5000)
// 			} else {
// 				num := RandInRange(0, len(b5000))
// 				r.SetBanker(b5000[num], 5000)
// 			}
// 			return
// 		}
// 		if len(b2000) >= 1 {
// 			if len(b2000) == 1 {
// 				r.SetBanker(b2000[0], 2000)
// 			} else {
// 				num := RandInRange(0, len(b2000))
// 				r.SetBanker(b2000[num], 2000)
// 			}
// 			return
// 		}
// 	}
// }

// 设定庄家并发送数据
// func (r *Room) SetBanker(id string, takeMoney int32) {
// 	for _, v := range r.PlayerList {
// 		if v != nil && v.Id == id {
// 			v.IsBanker = true
// 			v.BankerMoney = float64(takeMoney)
// 			v.BankerCount++
// 			v.BankerStatus = msg.BankerStatus_BankerUp
// 			r.BankerId = id
// 			r.BankerMoney = float64(takeMoney)
// 			r.IsConBanker = true
// 			nowTime := time.Now().Unix()
// 			v.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), r.RoomId)
// 			reason := "庄家申请上庄"
// 			c4c.BankerStatus(v, 1, nowTime, v.RoundId, reason)
// 			log.Debug("玩家当庄:%v,当庄金额:%v", v.Id, v.BankerMoney)

// 			data := &msg.BankerData_S2C{}
// 			data.Banker = v.RespPlayerData()
// 			data.TakeMoney = takeMoney
// 			r.BroadCastMsg(data, "BankerData_S2C")
// 		}
// 	}
// }

// 清除机器庄家
// func (r *Room) ClearRobotBanker() {
// 	for _, v := range r.PlayerList {
// 		if v != nil {
// 			if v.IsRobot == true && v.IsBanker == true {
// 				v.IsBanker = false
// 				v.BankerMoney = 0
// 				v.BankerCount = 0
// 				v.BankerStatus = 0
// 			}
// 		}
// 	}
// }

// 获取派奖前玩家投注的数据
func (r *Room) SetPlayerDownBet() {
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == false && v.IsAction == true {
			data := &PlayerDownBet{}
			data.Id = v.Id
			data.RoomId = r.RoomId
			data.GameId = conf.Server.GameID
			PeriodsNum, _ := FFCPeriodsAdd(r.PeriodsNum, r.PeriodsTime)
			data.PeriodsNum = PeriodsNum // 當期開始時間
			data.PeriodsTime = r.PeriodsTime
			if r.RoomId == "1" {
				data.LotteryType = "hn60"
			} else if r.RoomId == "2" {
				data.LotteryType = "qiqffc"
			}
			data.DownBetInfo = v.DownBetMoney
			data.DownBetTime = time.Now().Format("2006-01-02 15:04:05")
			InsertPlayerDownBet(PeriodsNum[0:6], data) //todo
		}
	}
}

// 相對應棋號下一期分分彩(1440期) 返回:下一期 ,當期結束時間
func FFCPeriodsAdd(oldPeriod string, oldPeriodTime string) (string, string) {
	PeriodsArr := strings.Split(oldPeriod, "-")
	NewPeriodsNum := common.Str2Int(PeriodsArr[1]) + 1
	NewTimeStamp := common.TimestrToTimestamp(oldPeriodTime, 5) + 60
	NewTimeStr := common.TimeFormatDate(NewTimeStamp)
	if NewPeriodsNum > 1440 { //隔日切換日期
		NewPeriodsNum = 1
		return fmt.Sprintf("%s-%04d", common.DateFromTimeStamp(NewTimeStamp), NewPeriodsNum), NewTimeStr
	}
	return fmt.Sprintf("%s-%04d", PeriodsArr[0], NewPeriodsNum), NewTimeStr
}

// 获取投注统计
func (r *Room) SeRoomTotalBet() {
	data := &RoomTotalBet{}
	data.RoomId = r.RoomId
	data.GameId = conf.Server.GameID
	data.PeriodsNum = r.PeriodsNum
	data.PeriodsTime = r.PeriodsTime

	PeriodsNumArr := strings.Split(r.PeriodsNum, "-")
	date := PeriodsNumArr[0]

	if r.RoomId == "1" {
		data.LotteryType = "hn60"
	} else if r.RoomId == "2" {
		data.LotteryType = "qiqffc"
	}
	data.PotTotalMoney = r.PlayerTotalMoney
	InsertRoomTotalBet(data, date[0:6]) //todo
}

func (r *Room) SetUserRoom(p *Player) {
	r.userRoomMutex.Lock()
	hall.UserRoom.Store(p.Id, r.RoomId)
	r.userRoomMutex.Unlock()
}

func (r *Room) DeleteUserRoom(p *Player) {
	r.userRoomMutex.Lock()
	defer r.userRoomMutex.Unlock()
	hall.UserRoom.Delete(p.Id)
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
