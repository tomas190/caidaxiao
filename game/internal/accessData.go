package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type pageData struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

type GameDataReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	RoomId    string `form:"room_id" json:"room_id"`
	RoundId   string `form:"round_id" json:"round_id"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Page      string `form:"page" json:"page"`
	Limit     string `form:"limit" json:"limit"`
}

type GameData struct {
	Time            int64       `json:"time"`
	TimeFmt         string      `json:"time_fmt"`
	StartTime       int64       `json:"start_time"`
	EndTime         int64       `json:"end_time"`
	PlayerId        string      `json:"player_id"`
	RoundId         string      `json:"round_id"`
	RoomId          string      `json:"room_id"`
	PackageId       int         `json:"package_id"` // 玩家品牌
	TaxRate         float64     `json:"tax_rate"`
	Lottery         interface{} `json:"lottery"`          // 开奖号码
	Card            interface{} `json:"card"`             // 开牌信息
	BetInfo         interface{} `json:"bet_info"`         // 玩家下注信息
	SettlementFunds interface{} `json:"settlement_funds"` // 结算信息 输赢结果
	SpareCash       interface{} `json:"spare_cash"`       // 剩余金额
	CreatedAt       int64       `json:"created_at"`       // 下注时间
	PeriodsNum      string      `json:"periods_num"`      // 获奖期数
}

type GetSurPool struct {
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       float64 `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
	CountAfterWin                  float64 `json:"random_count_after_win"`       // 玩家贏錢重新開獎次數
	PercentageAfterWin             float64 `json:"random_percentage_after_win"`  // 玩家贏錢重新開獎機率
	CountAfterLose                 float64 `json:"random_count_after_lose"`      // 玩家輸錢重新開獎次數
	PercertageAfterLose            float64 `json:"random_percentage_after_lose"` // 玩家輸錢重新開獎機率
}

type UpSurPool struct {
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`         // （该游戏全部实际的玩家历史总输 _ （该游戏全部实际的玩家历史总赢 * 100%）
	CoefficientToTotalPlayer       float64 `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"` //玩家赠送金额
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
	RandomCountAfterWin            float64 `json:"random_count_after_win" bson:"random_count_after_win"`             // 玩家贏錢重新開獎次數
	RandomCountAfterLose           float64 `json:"random_count_after_lose" bson:"random_count_after_lose"`           // 玩家輸錢重新開獎次數
	RandomPercentageAfterWin       float64 `json:"random_percentage_after_win" bson:"random_percentage_after_win"`   // 玩家贏錢重新開獎機率
	RandomPercentageAfterLose      float64 `json:"random_percentage_after_lose" bson:"random_percentage_after_lose"` // 玩家輸錢重新開獎機率
}

type GRobotData struct {
	RoomId      string       `json:"room_id" bson:"room_id"`
	RoomTime    int64        `json:"room_time" bson:"room_time"`
	RobotNum    int          `json:"robot_num" bson:"robot_num"`
	BigPot      *ChipDownBet `json:"big_pot" bson:"big_pot"`
	SmallPot    *ChipDownBet `json:"small_pot" bson:"small_pot"`
	SinglePot   *ChipDownBet `json:"single_pot" bson:"single_pot"`
	DoublePot   *ChipDownBet `json:"double_pot" bson:"double_pot"`
	PairPot     *ChipDownBet `json:"pair_pot" bson:"pair_pot"`
	StraightPot *ChipDownBet `json:"straight_pot" bson:"straight_pot"`
	LeopardPot  *ChipDownBet `json:"leopard_pot" bson:"leopard_pot"`
}

type CaiYuanReq struct {
	GameId     string `form:"game_id" json:"game_id"`
	PrizeType  string `form:"prize_type" json:"prize_type"`
	PeriodsNum string `form:"periods_num" json:"periods_num"`
	Page       string `form:"page" json:"page"`
	Limit      string `form:"limit" json:"limit"`
}

type RoomType struct {
	GameId string `form:"game_id" json:"game_id"`
	RoomId string `form:"room_id" json:"room_id"`
	IsOpen string `form:"room_status" json:"room_status"`
}

type GamePayReq struct {
	UserId    string `form:"user_id" json:"user_id"`
	MinBet    string `form:"min_bet" json:"min_bet"`
	MaxBet    string `form:"max_bet" json:"max_bet"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Lottery   string `form:"lottery" json:"lottery"`
	Limit     string `form:"limit" json:"limit"`
}

type GameWinReq struct {
	UserId      string `form:"user_id" json:"user_id"`
	LevelAmount string `form:"level_amount" json:"level_amount"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
}

type GamePayResp struct {
	GameCount int     `json:"game_count" bson:"game_count"`
	TotalWin  float64 `json:"total_win" bson:"total_win"`
	TotalLose float64 `json:"total_lose" bson:"total_lose"`
}

type GameLimitBet struct {
	UserId  string `form:"user_id" bson:"user_id" json:"user_id"`
	GameId  string `form:"game_id" bson:"game_id" json:"game_id"`
	MinBet  string `form:"min_bet" bson:"min_bet" json:"min_bet"`
	MaxBet  string `form:"max_bet" bson:"max_bet" json:"max_bet"`
	TimeFmt string `form:"time_fmt" bson:"time_fmt" json:"time_fmt"`
}

type GameRoomLimitBet struct {
	Room   string `form:"room_id" bson:"room_id" json:"room_id"`
	GameId string `form:"game_id" bson:"game_id" json:"game_id"`
	MinBet string `form:"min_bet" bson:"min_bet" json:"min_bet"`
	MaxBet string `form:"max_bet" bson:"max_bet" json:"max_bet"`
}

type PlayerGameInfoReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	PackageID string `form:"package_id" json:"package_id"`
	RoomID    string `form:"room_id" json:"room_id"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Page      string `form:"page" json:"page"`
	Limit     string `form:"limit" json:"limit"`
}

type PlayerProfitInfo struct {
	PlayerID  string  `json:"player_id"`  // 玩家id
	TotalWin  float64 `json:"total_win"`  // 总赢
	TotalLose float64 `json:"total_lose"` // 总输
	Profit    float64 `json:"profit"`     // 输赢差
}

// type UserRoundDataRes struct {
// 	Code int            `json:"code"`
// 	Msg  string         `json:"msg"`
// 	Data DataRoundsData `json:"data"`
// }
// type DataRoundsData struct {
// 	Total int        `json:"total"`
// 	List  []GameData `json:"list"`
// }

const (
	SuccCode = 0
	ErrCode  = -1
)

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	http.HandleFunc("/api/getGameData", getAccessData)
	// 查询子游戏盈余池数据
	http.HandleFunc("/api/getSurplusOne", getSurplusOne)
	// 修改盈余池数据
	http.HandleFunc("/api/uptSurplusConf", uptSurplusOne)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)
	// 获取机器人数据
	http.HandleFunc("/api/getRobotData", getRobotData)
	// 获取彩源玩家投注数据(系統部)
	http.HandleFunc("/api/getUsersPlayInfo", getPlayerDownBet)
	// 获取彩源房间投注统计(系統部)
	http.HandleFunc("/api/getPrizeTotalBet", getRoomTotalBet)
	// 接口操作关闭或开启房源
	http.HandleFunc("/api/changeRoomStatus", HandleRoomType)
	// 分分彩包赔活动
	http.HandleFunc("/api/HandleBaoPay", HandleBaoPay)
	// 分分彩连赢活动（河内分分彩）
	http.HandleFunc("/api/HandleHeNeiWin", HandleHeNeiWin)
	// 分分彩连赢活动（奇趣分分彩）
	http.HandleFunc("/api/HandleQiQuWin", HandleQiQuWin)
	// 设定玩家下注限紅
	http.HandleFunc("/api/setUserLimitBet", setUserLimitBet)
	// 获取玩家下注限紅
	http.HandleFunc("/api/getUserLimitBet", getUserLimitBet)
	// 踢除玩家並退资金
	http.HandleFunc("/api/kickUser", kickUser)
	// 退资金
	http.HandleFunc("/api/logoutUser", logoutUser)
	// 解鎖资金
	http.HandleFunc("/api/unLockUserMoney", unLockUserMoney)
	// 查看玩家获利状况
	http.HandleFunc("/api/PlayerProfit", getPlayerGameInfo)
	// 查看获利状况(单一玩家or品牌)
	http.HandleFunc("/api/getStatementTotal", getStatementTotal)
	// 设定房间下注限紅
	http.HandleFunc("/api/setRoomLimitBet", setRoomLimitBet)
	// 查詢房间下注限紅
	http.HandleFunc("/api/getRoomLimitBet", getRoomLimitBet)
	// 視訊直播消費扣款
	http.HandleFunc("/api/userZhiBoReward", userZhiBoReward)
	// 线上玩家
	http.HandleFunc("/api/getOnlineTotal", getOnlineTotal)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

type playergamepageData struct {
	Total         int         `json:"total"`
	List          interface{} `json:"list"`
	ServerWinLoss interface{} `json:"server_winloss"`
}

type serverWinLoss struct {
	TotalWin  float64 `json:"total_win"`  // 剩余金额
	TotalLose float64 `json:"total_lose"` // 下注时间
	Profit    float64 `json:"profit"`     // 获奖期数
}

///api/PlayerProfit
// id 可选  (指定玩家或所有玩家)
// game_id 必填
// package_id 可选 (指定渠道或所有渠道)
// start_time 必填
// end_time 必填
// page 可选 (预设1)
// limit 可选 (预设50)
// 玩家 总输  总赢  输赢差额
func getPlayerGameInfo(w http.ResponseWriter, r *http.Request) {
	var req PlayerGameInfoReq
	var msg = ""
	var result playergamepageData
	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.PackageID = r.FormValue("package_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	// log.Debug("获取分页数据:%v", req.Page)

	defer func() {
		js, err := json.Marshal(NewResp(SuccCode, msg, result))
		if err != nil && msg != "" {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: msg, Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}()

	selector := bson.M{}

	if req.Id != "" {
		selector["user_id"] = req.Id
	}

	if req.GameId != conf.Server.GameID {
		msg = "game_id错误"
		return
	}

	if req.PackageID != "" {
		packageid, err := strconv.Atoi(req.PackageID)
		if err != nil {
			msg = "packageid错误"
			return
		}
		selector["package_id"] = packageid
	}

	sTime, _ := strconv.Atoi(req.StartTime)
	eTime, _ := strconv.Atoi(req.EndTime)
	if sTime == 0 || eTime == 0 || sTime > eTime {
		msg = "请填入正确时间参数"
		return
	}
	selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}

	if req.Page == "" {
		req.Page = "1"
	}

	page, err := strconv.Atoi(req.Page)
	if err != nil {
		msg = "page参数不合法"
		log.Debug("req.Page轉int :%v Error:%v", req.Page, err.Error())
		return
	}
	if page < 1 {
		msg = "page参数不合法"
		log.Debug("page参数不合法 :%v Error:%v", page, err.Error())
		return
	}

	var limits int
	if len(req.Limit) == 0 {
		limits = 50
	} else {
		limits, _ = strconv.Atoi(req.Limit)
	}

	recodes, err := PlayerGameInfo(selector)
	if err != nil {
		msg = "獲取數據錯誤"
		log.Debug("獲取數據錯誤:%v", err)
		return
	}

	var gameDataMap = map[string]*PlayerProfitInfo{}
	// common.Debug_log("资料总长:%v", len(recodes))
	for _, v := range recodes { //加总
		// common.Debug_log("v.TotalWin:%v, v.TotalLose%v", v.TotalWin, v.TotalLose)
		pd, ok := gameDataMap[v.UserId]
		if ok {

			pd.TotalWin = pd.TotalWin + v.TotalWin
			pd.TotalLose = pd.TotalLose + v.TotalLose // 负数
			pd.Profit = pd.TotalWin + pd.TotalLose
		} else {
			gameDataMap[v.UserId] = &PlayerProfitInfo{
				PlayerID:  v.UserId,
				TotalWin:  v.TotalWin,
				TotalLose: v.TotalLose,
				Profit:    v.TotalWin + v.TotalLose,
			}
		}
	}
	// common.Debug_log("map长度:%v", len(gameDataMap))
	var playerProfitData []*PlayerProfitInfo
	for _, v := range gameDataMap { //放进阵列
		playerProfitData = append(playerProfitData, v)
	}
	sort.Slice(playerProfitData, func(i, j int) bool {
		if playerProfitData[i].Profit > playerProfitData[j].Profit {
			return true
		}
		return false
	})

	playerProfitArr := splitArray(playerProfitData, limits)
	playerProfitRsp := playerProfitArr[page-1]

	serverWinLoss := &serverWinLoss{}
	for _, ppd := range playerProfitData {
		serverWinLoss.TotalLose = serverWinLoss.TotalLose + ppd.TotalLose
		serverWinLoss.TotalWin = serverWinLoss.TotalWin + ppd.TotalWin
	}
	serverWinLoss.Profit = serverWinLoss.TotalLose + serverWinLoss.TotalWin

	result.Total = len(playerProfitData)
	result.ServerWinLoss = serverWinLoss
	result.List = playerProfitRsp

}

//slice切分
func splitArray(arr []*PlayerProfitInfo, limits int) [][]*PlayerProfitInfo {
	var data = make([][]*PlayerProfitInfo, 0)
	for i := 1; i <= int(math.Floor(float64(len(arr)/limits)))+1; i++ {
		low := limits * (i - 1)
		if low > len(arr) {
			return [][]*PlayerProfitInfo{}
		}
		high := limits * i
		if high > len(arr) {
			high = len(arr)
		}
		// fmt.Println(arr[low:high])
		data = append(data, arr[low:high])
		//intss[i-1] = append(intss[i-1],ints[low:high])

	}
	return data
}

type StatementTotalData struct {
	LoseStatementTotal float64 `json:"lose_statement_total"` // 1. 传package_id时品牌下的总输 2. 传id时单玩家总输
	WinStatementTotal  float64 `json:"win_statement_total"`  // 1.传package_id时品牌下的总赢 2.传id时单玩家总赢
	GameID             string  `json:"game_id"`              // 游戏id
	GameName           string  `json:"game_name"`            //游戏名称
	BetMoney           float64 `json:"bet_money"`            //1. 传package_id时该品牌的有效投注 2. 传id时单玩家的有效投注
	Count              []int   `json:"count"`                //在线人数 去重后的uid 也就是查询时间段内的有记录的玩家ID
}

///api/getStatementTotal
// id 可选  (指定玩家或所有玩家)
// package_id 可选 (指定渠道或所有渠道)
// game_id 必填
// start_time 必填
// end_time 必填
// room_id 可选
// 玩家 总输  总赢  输赢差额
func getStatementTotal(w http.ResponseWriter, r *http.Request) {
	var req PlayerGameInfoReq
	var msg = ""
	var result = StatementTotalData{}
	req.Id = r.FormValue("id")
	req.PackageID = r.FormValue("package_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.RoomID = r.FormValue("room_id")
	// log.Debug("获取分页数据:%v", req.Page)

	defer func() {
		js, err := json.Marshal(NewResp(SuccCode, msg, result))
		if err != nil && msg != "" {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: msg, Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}()

	selector := bson.M{}
	if (req.Id != "" && req.PackageID != "") || (req.Id == "" && req.PackageID == "") { // 两个都不为空
		msg = "错误:id.packageid同时存在或皆为空"
		return
	}

	if req.Id != "" {
		selector["user_id"] = req.Id
	}

	if req.PackageID != "" {
		packageid, err := strconv.Atoi(req.PackageID)
		if err != nil {
			msg = "packageid错误"
			return
		}
		if packageid != 777 { // 777為所有渠道
			selector["package_id"] = packageid
		}

	}

	if req.RoomID != "" { // 依據彩種計算總流水
		common.Debug_log(req.RoomID)
		selector["room_id"] = req.RoomID
	}

	sTime, _ := strconv.Atoi(req.StartTime)
	eTime, _ := strconv.Atoi(req.EndTime)
	if sTime == 0 || eTime == 0 || sTime > eTime {
		msg = "请填入正确时间参数"
		return
	}
	selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}

	recodes, err := PlayerGameInfo(selector)
	if err != nil {
		msg = "獲取數據錯誤"
		log.Debug("獲取數據錯誤:%v", err)
		return
	}
	for _, v := range recodes { //加总
		// common.Debug_log("v.TotalWin:%v, v.TotalLose%v", v.TotalWin, v.TotalLose)
		result.BetMoney = result.BetMoney + float64(v.DownBetInfo.BigDownBet+v.DownBetInfo.SmallDownBet+v.DownBetInfo.LeopardDownBet)
		result.LoseStatementTotal = result.LoseStatementTotal + v.TotalLose
		result.WinStatementTotal = result.WinStatementTotal + v.TotalWin
		uid := common.Str2Int(v.UserId)
		if !common.SearchSliInt(result.Count, uid) { //不在列表中就加入
			result.Count = append(result.Count, uid)
		}

	}
	result.GameID = conf.Server.GameID
	result.GameName = "分分彩猜大小"

}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq
	var msg = ""
	var result pageData
	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.RoomId = r.FormValue("room_id")
	req.RoundId = r.FormValue("round_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	defer func() {
		js, err := json.Marshal(NewResp(SuccCode, msg, result))
		if err != nil && msg != "" {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: msg, Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}()

	selector := bson.M{}

	if req.Id != "" {
		selector["id"] = req.Id
	}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoomId != "" {
		selector["room_id"] = req.RoomId
	}

	if req.RoundId != "" {
		selector["round_id"] = req.RoundId
	}

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	} else if sTime != 0 && eTime == 0 {
		selector["start_time"] = bson.M{"$gte": sTime}
	} else if eTime != 0 && sTime == 0 {
		selector["end_time"] = bson.M{"$lte": eTime}
	}

	if req.Page == "" {
		req.Page = "1"
	}

	page, err := strconv.Atoi(req.Page)
	if err != nil {
		msg = "page参数不合法"
		log.Debug("req.Page轉int :%v Error:%v", req.Page, err.Error())
		return
	}
	if page < 1 {
		msg = "page参数不合法"
		log.Debug("page参数不合法 :%v Error:%v", page, err.Error())
		return
	}

	var limits int
	if len(req.Limit) == 0 {
		limits = 50
	} else {
		limits, _ = strconv.Atoi(req.Limit)
	}

	recodes, count, err := GetDownRecodeList(page, limits, selector, "-down_bet_time")
	if err != nil {
		msg = "獲取數據錯誤"
		log.Debug("獲取數據錯誤:%v", err)
		return
	}

	var gameData []GameData
	for i := 0; i < len(recodes); i++ {
		var gd GameData
		pr := recodes[i]
		gd.Time = pr.DownBetTime
		gd.TimeFmt = FormatTime(pr.DownBetTime, "2006-01-02 15:04:05")
		gd.StartTime = pr.StartTime
		gd.EndTime = pr.EndTime
		gd.PlayerId = pr.Id
		gd.PackageId = pr.PackageId
		gd.RoomId = pr.RoomId
		gd.RoundId = pr.RoundId
		gd.Lottery = pr.Lottery
		gd.BetInfo = pr.DownBetInfo
		gd.Card = pr.CardResult
		gd.SettlementFunds = pr.SettlementFunds
		gd.SpareCash = pr.SpareCash
		gd.TaxRate = pr.TaxRate
		gd.CreatedAt = pr.DownBetTime
		gd.PeriodsNum = pr.PeriodsNum
		gameData = append(gameData, gd)
	}

	result.Total = count
	result.List = gameData

}

// 查询子游戏盈余池数据
func getSurplusOne(w http.ResponseWriter, r *http.Request) {

	GameId := r.FormValue("game_id")

	if GameId != conf.Server.GameID {
		log.Debug("game_id错误:%v   %v", GameId, conf.Server.GameID)
		return
	}

	var getSur GetSurPool
	getSur.PlayerTotalLose = ServerSurPool.TotalLost
	getSur.PlayerTotalWin = ServerSurPool.TotalWin
	getSur.PercentageToTotalWin = ServerSurPool.KillPercent
	getSur.TotalPlayer = int32(ServerSurPool.SumUser)
	getSur.CoefficientToTotalPlayer = ServerSurPool.MoneyPrizeOneUser
	getSur.FinalPercentage = ServerSurPool.FinalPercentage
	getSur.PlayerTotalLoseWin = ServerSurPool.UserLostMinusWin
	getSur.SurplusPool = ServerSurPool.PoolBalance
	getSur.PlayerLoseRateAfterSurplusPool = ServerSurPool.LoseRateAfterSurplus
	getSur.DataCorrection = ServerSurPool.DataCorrection
	getSur.CountAfterWin = ServerSurPool.CountAfterWin
	getSur.PercentageAfterWin = ServerSurPool.PercentageAfterWin
	getSur.CountAfterLose = ServerSurPool.CountAfterLose
	getSur.PercertageAfterLose = ServerSurPool.PercertageAfterLose

	js, err := json.Marshal(NewResp(SuccCode, "", getSur))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func uptSurplusOne(w http.ResponseWriter, r *http.Request) {

	gameID := r.PostFormValue("game_id")
	if gameID != conf.Server.GameID {
		log.Debug("game_id错误 %v", gameID)
		return
	}

	rateSur, errR := strconv.ParseFloat(r.PostFormValue("player_lose_rate_after_surplus_pool"), 64)
	percentage, errP := strconv.ParseFloat(r.PostFormValue("percentage_to_total_win"), 64)
	coefficient, errC := strconv.ParseFloat(r.PostFormValue("coefficient_to_total_player"), 64)
	final, errF := strconv.ParseFloat(r.PostFormValue("final_percentage"), 64)
	correction, errD := strconv.ParseFloat(r.PostFormValue("data_correction"), 64)
	CountAfterWin, errW := strconv.ParseFloat(r.PostFormValue("random_count_after_win"), 64)
	PercentageAfterWin, errPW := strconv.ParseFloat(r.PostFormValue("random_percentage_after_win"), 64)
	CountAfterLose, errL := strconv.ParseFloat(r.PostFormValue("random_count_after_lose"), 64)
	PercentageAfterLose, errPL := strconv.ParseFloat(r.PostFormValue("random_percentage_after_lose"), 64)

	if errR == nil {
		ServerSurPool.LoseRateAfterSurplus = rateSur
	}
	if errP == nil {
		ServerSurPool.KillPercent = percentage
	}
	if errC == nil {
		ServerSurPool.MoneyPrizeOneUser = coefficient
	}
	if errF == nil {
		ServerSurPool.FinalPercentage = final
	}
	if errD == nil {
		ServerSurPool.DataCorrection = correction
	}
	if errW == nil {
		ServerSurPool.CountAfterWin = CountAfterWin
	}
	if errPW == nil {
		ServerSurPool.PercentageAfterWin = PercentageAfterWin
	}
	if errL == nil {
		ServerSurPool.CountAfterLose = CountAfterLose
	}
	if errPL == nil {
		ServerSurPool.PercertageAfterLose = PercentageAfterLose
	}

	if errP == nil || errC == nil || errF == nil || errD == nil {
		itemPoolBalance := (ServerSurPool.TotalLost - ServerSurPool.TotalWin*ServerSurPool.KillPercent - ServerSurPool.SumUser*ServerSurPool.MoneyPrizeOneUser + ServerSurPool.DataCorrection) * ServerSurPool.FinalPercentage
		p, errSur := strconv.ParseFloat(fmt.Sprintf("%.2f", itemPoolBalance), 64)
		if errSur == nil {
			ServerSurPool.PoolBalance = p
		} else {
			common.Debug_log("errSur=%v", errSur)
		}
	}

	SaveServerConfig()

	var upt UpSurPool
	upt.PlayerLoseRateAfterSurplusPool = ServerSurPool.LoseRateAfterSurplus
	upt.PercentageToTotalWin = ServerSurPool.KillPercent
	upt.CoefficientToTotalPlayer = ServerSurPool.MoneyPrizeOneUser
	upt.FinalPercentage = ServerSurPool.FinalPercentage
	upt.DataCorrection = ServerSurPool.DataCorrection
	upt.RandomCountAfterWin = ServerSurPool.CountAfterWin
	upt.RandomPercentageAfterWin = ServerSurPool.PercentageAfterWin
	upt.RandomCountAfterLose = ServerSurPool.CountAfterLose
	upt.RandomPercentageAfterLose = ServerSurPool.PercertageAfterLose

	js, err := json.Marshal(NewResp(SuccCode, "", upt))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func reqPlayerLeave(w http.ResponseWriter, r *http.Request) {
	// Id := r.FormValue("id")
	// log.Debug("reqPlayerLeave 踢出玩家:%v", Id)
	// rid, _ := hall.UserRoom.Load(Id)
	// v, _ := hall.RoomRecord.Load(rid)
	// if v != nil {
	// 	room := v.(*Room)
	// 	user, _ := hall.UserRecord.Load(Id)
	// 	if user != nil {
	// 		p := user.(*Player)
	// 		room.IsConBanker = false
	// 		hall.UserRecord.Delete(p.Id)
	// 		p.PlayerExitRoom()
	// 		sendLogout(p.Id) // 登出
	// 		// c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
	// 		leaveHall := &msg.Logout_S2C{}
	// 		p.SendMsg(leaveHall, "Logout_S2C")

	// 		js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
	// 		if err != nil {
	// 			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
	// 			return
	// 		}
	// 		w.Write(js)
	// 	}
	// }
}

func getRobotData(w http.ResponseWriter, r *http.Request) {
	_ = r
	recodes, err := GetRobotData()
	if err != nil {
		return
	}

	var rData []GRobotData
	for i := 0; i < len(recodes); i++ {
		var rd GRobotData
		rd.BigPot = new(ChipDownBet)
		rd.SmallPot = new(ChipDownBet)
		rd.SinglePot = new(ChipDownBet)
		rd.DoublePot = new(ChipDownBet)
		rd.PairPot = new(ChipDownBet)
		rd.StraightPot = new(ChipDownBet)
		rd.LeopardPot = new(ChipDownBet)
		pr := recodes[i]
		log.Debug("获取机器数据:%v", pr)
		rd.RoomId = pr.RoomId
		rd.RoomTime = pr.RoomTime
		rd.RobotNum = pr.RobotNum
		rd.BigPot = pr.BigPot
		rd.SmallPot = pr.SmallPot
		rd.SinglePot = pr.SinglePot
		rd.DoublePot = pr.DoublePot
		rd.PairPot = pr.PairPot
		rd.StraightPot = pr.StraightPot
		rd.LeopardPot = pr.LeopardPot
		rData = append(rData, rd)
	}

	js, err := json.Marshal(NewResp(SuccCode, "", rData))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getPlayerDownBet(w http.ResponseWriter, r *http.Request) {
	var req CaiYuanReq

	req.GameId = r.FormValue("game_id")
	// req.RoomId = r.FormValue("room_id")
	req.PrizeType = r.FormValue("prize_type")
	req.PeriodsNum = r.FormValue("periods_num")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	selector := bson.M{}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.PrizeType != "" {
		selector["lottery_type"] = req.PrizeType
	}

	// if req.PeriodsNum != "" {
	// 	selector["periods_num"] = req.PeriodsNum
	// }

	var date string
	if req.PeriodsNum != "" {
		selector["periods_num"] = req.PeriodsNum
		PeriodsNumArr := strings.Split(req.PeriodsNum, "-")
		date = PeriodsNumArr[0]
	} else {
		return
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)

	recodes, count, err := GetPlayerDownBet(date[0:6], page, limits, selector, "-down_bet_time")
	if err != nil {
		return
	}

	var result pageData
	result.Total = count
	result.List = recodes

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// func getRoomTotalBet(w http.ResponseWriter, r *http.Request) {
// 	var req CaiYuanReq

// 	req.GameId = r.FormValue("game_id")
// 	// req.RoomId = r.FormValue("room_id")
// 	req.PrizeType = r.FormValue("prize_type")
// 	req.PeriodsNum = r.FormValue("periods_num")
// 	req.Page = r.FormValue("page")
// 	req.Limit = r.FormValue("limit")
// 	log.Debug("获取分页数据:%v", req.Page)

// 	selector := bson.M{}

// 	if req.GameId != "" {
// 		selector["game_id"] = req.GameId
// 	}

// 	if req.PrizeType != "" {
// 		selector["lottery_type"] = req.PrizeType
// 	}

// 	if req.PeriodsNum != "" {
// 		selector["periods_num"] = req.PeriodsNum
// 	}

// 	page, _ := strconv.Atoi(req.Page)

// 	limits, _ := strconv.Atoi(req.Limit)
// 	//if limits != 0 {
// 	//	selector["limit"] = limits
// 	//}

// 	recodes, count, err := GetRoomTotalBet(page, limits, selector, "-down_bet_time")
// 	if err != nil {
// 		return
// 	}

// 	var result pageData
// 	result.Total = count
// 	result.List = recodes

// 	js, err := json.Marshal(NewResp(SuccCode, "", result))
// 	if err != nil {
// 		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(js)
// }

func getRoomTotalBet(w http.ResponseWriter, r *http.Request) {
	var req CaiYuanReq

	req.GameId = r.FormValue("game_id")
	// req.RoomId = r.FormValue("room_id")
	req.PrizeType = r.FormValue("prize_type")
	req.PeriodsNum = r.FormValue("periods_num")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	selector := bson.M{}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.PrizeType != "" {
		selector["lottery_type"] = req.PrizeType
	}
	var date string
	if req.PeriodsNum != "" {
		selector["periods_num"] = req.PeriodsNum
		PeriodsNumArr := strings.Split(req.PeriodsNum, "-")
		date = PeriodsNumArr[0]
	} else {
		return
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetRoomTotalBet(date[0:6], page, limits, selector, "-down_bet_time")
	if err != nil {
		return
	}

	var result pageData
	result.Total = count
	result.List = recodes

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HandleRoomType(w http.ResponseWriter, r *http.Request) {
	var req RoomType

	req.GameId = r.FormValue("game_id")
	if req.GameId == "" || req.GameId != conf.Server.GameID { //檢測game_id
		log.Debug("game_id 有誤")
		return
	}
	req.RoomId = r.FormValue("room_id")
	if req.RoomId == "" { //檢測room_id
		log.Debug("room_id 有誤")
		return
	}
	req.IsOpen = r.FormValue("room_status")
	if req.IsOpen == "" { //檢測room_id
		log.Debug("roomStatus 有誤")
		return
	}

	log.Debug("RoomId:%v, IsOpen:%v", req.RoomId, req.IsOpen)
	var changeRoomID string // 变动的房间id
	if req.RoomId == "01" {
		changeRoomID = "1"
		if req.IsOpen == "1" {
			for _, v := range hall.roomList {
				if v != nil && v.RoomId == "1" {
					v.IsOpenRoom = true
				}
			}
		}
		if req.IsOpen == "0" {
			for _, v := range hall.roomList {
				if v != nil && v.RoomId == "1" {
					v.IsOpenRoom = false
					v.GameStat = msg.GameStep_LiuJu
				}
			}
		}
	}

	if req.RoomId == "02" {
		changeRoomID = "2"
		if req.IsOpen == "1" {
			for _, v := range hall.roomList {
				if v != nil && v.RoomId == "2" {
					v.IsOpenRoom = true
				}
			}
		}
		if req.IsOpen == "0" {
			for _, v := range hall.roomList {
				if v != nil && v.RoomId == "2" {
					v.IsOpenRoom = false
					v.GameStat = msg.GameStep_LiuJu
				}
			}
		}
	}
	var changeroom = &Room{}
	data := &msg.ChangeRoomType_S2C{}
	for _, v := range hall.roomList {
		if v != nil {
			if v.RoomId == "1" {
				data.Room01 = v.IsOpenRoom
			} else if v.RoomId == "2" {
				data.Room02 = v.IsOpenRoom
			}

			if v.RoomId == changeRoomID {
				changeroom = v
			}
		}
	}

	// 发送给所有玩家房间变更
	hall.UserRecord.Range(func(_, value interface{}) bool {
		u := value.(*Player)
		u.SendMsg(data, "ChangeRoomType_S2C")
		room_id, ok := hall.UserRoom.Load(u.Id)
		if ok {
			if room_id.(string) == changeRoomID && req.IsOpen == "0" && u.IsRobot == false { //此次选择的房间关闭
				kickUserInRoom(u.Id)
				changeroom.unlockUserBetMoney(u)
			}
		}

		return true
	})

	roomidCount := SearchCMD{
		DBName: dbName,
		CName:  RoomStatusDB,
		Query:  bson.M{"room_id": changeroom.RoomId},
	}
	if FindCountByQuery(roomidCount) > 0 {
		Update := SearchCMD{
			DBName: dbName,
			CName:  RoomStatusDB,
			Query:  roomidCount.Query,
			Update: bson.M{"$set": bson.M{
				"Is_Open":  changeroom.IsOpenRoom,
				"time_fmt": time.Now().Format("2006-01-02_15:04:05"),
			}},
		}
		roomUpdate := &RoomStatus{}
		if FindAndUpdateByQuery(Update, roomUpdate) {
			common.Debug_log("房間%v數據庫更新成功 Is_Open:%v ", changeroom.RoomId, changeroom.IsOpenRoom)
		}
	} else {
		roomInsert := &RoomStatus{
			RoomId:  changeroom.RoomId,
			GameId:  req.GameId,
			MinBet:  changeroom.RoomMinBet,
			MaxBet:  changeroom.RoomMaxBet,
			IsOpen:  changeroom.IsOpenRoom,
			TimeFmt: time.Now().Format("2006-01-02_15:04:05"),
		}
		InsertRoomStatus(roomInsert)
	}

	js, err := json.Marshal(NewResp(SuccCode, "", ""))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HandleBaoPay(w http.ResponseWriter, r *http.Request) {
	var req GamePayReq

	req.UserId = r.FormValue("user_id")
	req.Lottery = r.FormValue("lottery")
	req.MinBet = r.FormValue("min_bet")
	req.MaxBet = r.FormValue("max_bet")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Limit = r.FormValue("limit")

	minBet, errmax := strconv.Atoi(req.MinBet)

	maxBet, errmin := strconv.Atoi(req.MaxBet)

	sTime, errst := strconv.Atoi(req.StartTime)

	eTime, erret := strconv.Atoi(req.EndTime)

	log.Debug("赔付数据:%v", req)

	selector := bson.M{}

	if req.UserId == "" {
		log.Debug("UserId為空")
		return
	} else {
		selector["user_id"] = req.UserId
	}

	if req.Lottery != "HNFFC" && req.Lottery != "PTXFFC" {
		log.Debug("Lottery不存在 %v %v %v", req.Lottery, req.Lottery != "PTXFFC", "PTXFFC")
		return
	} else if req.Lottery == "HNFFC" {
		selector["room_id"] = "1"
	} else {
		selector["room_id"] = "2"
	}

	if errmax != nil || errmin != nil {
		log.Debug("maxBet minBet 輸入錯誤")
		return
	} else if minBet > maxBet {
		log.Debug("minBet必須小於maxBet")
		return
	}

	if errst != nil || erret != nil { //此刻往前計算七天
		endTime := time.Now().Unix()
		currentTime := time.Now()
		oldTime := currentTime.AddDate(0, 0, -7)
		startTime := oldTime.Unix()
		selector["down_bet_time"] = bson.M{"$gte": startTime, "$lte": endTime}
		log.Debug("sTime eTime 輸入格式錯誤 默認為此刻往前推七天")
	} else {
		if sTime > eTime {
			log.Debug("end_time必須大於start_time")
			return
		}
		if eTime-sTime > 604800 { //時間段超過七天

			currentTime := time.Unix(int64(eTime), 0)
			oldTime := currentTime.AddDate(0, 0, -7)
			startTime := oldTime.Unix()
			selector["down_bet_time"] = bson.M{"$gte": startTime, "$lte": eTime}
		} else { // 輸入正確
			selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
		}
	}

	limits, _ := strconv.Atoi(req.Limit)
	if limits == 0 {
		limits = 10
	}
	log.Debug("bsonM:%v", selector)
	recodes, err := GetPlayerGameData(selector, limits, "down_bet_time")
	log.Debug("获取数据筆數:%v  \nDATA:%v", len(recodes), recodes)

	data := &GamePayResp{}
	for _, v := range recodes {
		var num int
		if v.DownBetInfo.BigDownBet > 0 {
			num++
		}
		if v.DownBetInfo.SmallDownBet > 0 {
			num++
		}
		if v.DownBetInfo.LeopardDownBet > 0 {
			num += 2
		}
		if num > 1 { //只有單獨下大或小才符合
			continue
		}

		downBet := v.DownBetInfo.BigDownBet + v.DownBetInfo.SmallDownBet + v.DownBetInfo.LeopardDownBet

		if downBet < int32(minBet) || downBet > int32(maxBet) { // 下注超過活動區間
			continue
		}

		data.GameCount++

		if v.SettlementFunds > 0 {
			data.TotalWin += v.TotalWin
		} else {
			data.TotalLose += v.SettlementFunds
		}
	}

	var result pageData
	result.Total = 1
	result.List = data

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HandleHeNeiWin(w http.ResponseWriter, r *http.Request) {
	var req GameWinReq

	req.UserId = r.FormValue("user_id")
	req.LevelAmount = r.FormValue("level_amount")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")

	selector := bson.M{}

	if req.UserId != "" {
		selector["user_id"] = req.UserId
	}

	// 查询河内房间数据
	selector["room_id"] = "1"

	amount, _ := strconv.Atoi(req.LevelAmount)

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime == 0 || eTime == 0 {
		currentTime1 := time.Now()
		startTime := time.Date(currentTime1.Year(), currentTime1.Month(), currentTime1.Day(), 0, 0, 0, 0, currentTime1.Location()).Unix()
		currentTime2 := time.Now()
		endTime := time.Date(currentTime2.Year(), currentTime2.Month(), currentTime2.Day(), 23, 59, 59, 0, currentTime2.Location()).Unix()
		selector["down_bet_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	}

	recodes, err := GetPlayerWinData(selector)

	data := &GamePayResp{}
	var num int
	for _, v := range recodes {
		taxR, _ := mapTaxPercent[v.PackageId] //tax
		resWin := v.TotalLose + (v.TotalWin * (1 - taxR))
		if resWin >= float64(amount) {
			num++
			if num > data.GameCount {
				data.GameCount = num
			}
		} else {
			num = 0
		}
	}

	var result pageData
	result.Total = 1
	result.List = data

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HandleQiQuWin(w http.ResponseWriter, r *http.Request) {
	var req GameWinReq

	req.UserId = r.FormValue("user_id")
	req.LevelAmount = r.FormValue("level_amount")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")

	selector := bson.M{}

	if req.UserId != "" {
		selector["user_id"] = req.UserId
	}

	// 查询河内房间数据
	selector["room_id"] = "2"

	amount, _ := strconv.Atoi(req.LevelAmount)

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime == 0 || eTime == 0 {
		currentTime1 := time.Now()
		startTime := time.Date(currentTime1.Year(), currentTime1.Month(), currentTime1.Day(), 0, 0, 0, 0, currentTime1.Location()).Unix()
		currentTime2 := time.Now()
		endTime := time.Date(currentTime2.Year(), currentTime2.Month(), currentTime2.Day(), 23, 59, 59, 0, currentTime2.Location()).Unix()
		selector["down_bet_time"] = bson.M{"$gte": startTime, "$lte": endTime}
	}

	recodes, err := GetPlayerWinData(selector)

	data := &GamePayResp{}
	var num int
	for _, v := range recodes {

		taxR, _ := mapTaxPercent[v.PackageId] //tax

		resWin := v.TotalLose + (v.TotalWin * (1 - taxR))
		if resWin >= float64(amount) {
			num++
			if num > data.GameCount {
				data.GameCount = num
			}
		} else {
			num = 0
		}
	}

	var result pageData
	result.Total = 1
	result.List = data

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func setUserLimitBet(w http.ResponseWriter, r *http.Request) {
	var req GameLimitBet

	req.UserId = r.PostFormValue("user_id")
	req.GameId = r.PostFormValue("game_id")
	req.MinBet = r.PostFormValue("min_bet")
	req.MaxBet = r.PostFormValue("max_bet")
	req.TimeFmt = time.Now().Format("2006-01-02_15:04:05")

	if req.UserId == "" {
		return
	}
	log.Debug("限制玩家下注:%v", req)
	minBet, _ := strconv.Atoi(req.MinBet)
	maxBet, _ := strconv.Atoi(req.MaxBet)

	uidNum, errU := common.Str2int32(req.UserId)
	if errU != nil {
		return
	}

	hall.UserRecord.Range(func(_, value interface{}) bool {
		u := value.(*Player)
		if u.Id == uidNum {
			log.Debug("玩家id:%v,限制:%v,%v", u.Id, minBet, maxBet)
			u.MinBet = int32(minBet)
			u.MaxBet = int32(maxBet)
		}
		return true
	})

	useridCount := SearchCMD{
		DBName: dbName,
		CName:  UserLimitBetDB,
		Query:  bson.M{"user_id": req.UserId},
	}
	if FindCountByQuery(useridCount) > 0 {
		Update := SearchCMD{
			DBName: dbName,
			CName:  UserLimitBetDB,
			Query:  useridCount.Query,
			Update: bson.M{"$set": bson.M{
				"min_bet":  req.MinBet,
				"max_bet":  req.MaxBet,
				"time_fmt": time.Now().Format("2006-01-02_15:04:05"),
			}},
		}
		userlimitUpdate := &GameLimitBet{}
		if FindAndUpdatelastByQuery(Update, userlimitUpdate) {
			common.Debug_log("玩家%v限紅更新成功 min_bet:%v max_bet:%v", req.UserId, minBet, maxBet)
		}
	} else {
		// 插入玩家限定下注数据
		InsertUserLimitBet(&req)
	}

	js, err := json.Marshal(NewResp(SuccCode, "", req))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getUserLimitBet(w http.ResponseWriter, r *http.Request) {
	var req GameLimitBet

	req.UserId = r.FormValue("user_id")
	req.GameId = r.FormValue("game_id")

	selector := bson.M{}

	if req.UserId != "" {
		selector["user_id"] = req.UserId
	}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	recodes, count, err := GetUserLimitBet(selector)

	var result pageData
	result.Total = count
	result.List = recodes

	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// /api/setRoomLimitBet
// room_id (string) ex:"1","2"
// game_id (string) 游戏ID
// min_bet (int)    投注最小限制
// max_bet (int)    投注最大限制
func setRoomLimitBet(w http.ResponseWriter, r *http.Request) {
	var req GameRoomLimitBet
	var msg string = "修改成功"
	req.Room = r.PostFormValue("room_id")
	req.GameId = r.PostFormValue("game_id")
	req.MinBet = r.PostFormValue("min_bet")
	req.MaxBet = r.PostFormValue("max_bet")

	minBet, errmin := strconv.Atoi(req.MinBet)
	maxBet, errmax := strconv.Atoi(req.MaxBet)

	defer func() {
		js, err := json.Marshal(NewResp(SuccCode, msg, req))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}()

	if errmin != nil || errmax != nil || minBet > maxBet {
		msg = "限制投注min_bet,max_bet 输入错误"
		return
	}

	if req.GameId != conf.Server.GameID {
		msg = "GameId输入错误"
		return
	}

	exist := false
	var changeroom *Room
	for _, v := range hall.roomList {
		common.Debug_log(v.RoomId, req.Room, v.RoomId == req.Room)
		if v != nil && v.RoomId == req.Room {
			v.RoomMaxBet = int32(maxBet)
			v.RoomMinBet = int32(minBet)
			common.Debug_log("房间%v 修改限制下注:%v", v.RoomId, req)
			changeroom = v
			exist = true
		}
	}
	if !exist {
		msg = "无此房间"
		return
	}

	roomidCount := SearchCMD{
		DBName: dbName,
		CName:  RoomStatusDB,
		Query:  bson.M{"room_id": changeroom.RoomId},
	}
	if FindCountByQuery(roomidCount) > 0 {
		Update := SearchCMD{
			DBName: dbName,
			CName:  RoomStatusDB,
			Query:  roomidCount.Query,
			Update: bson.M{"$set": bson.M{
				"min_bet":  changeroom.RoomMinBet,
				"max_bet":  changeroom.RoomMaxBet,
				"time_fmt": time.Now().Format("2006-01-02_15:04:05"),
			}},
		}
		roomUpdate := &RoomStatus{}
		if FindAndUpdateByQuery(Update, roomUpdate) {
			common.Debug_log("房間%v數據庫更新成功 min_bet:%v max_bet:%v", changeroom.RoomId, changeroom.RoomMinBet, changeroom.RoomMaxBet)
		}
	} else {
		roomInsert := &RoomStatus{
			RoomId:  changeroom.RoomId,
			GameId:  req.GameId,
			MinBet:  changeroom.RoomMinBet,
			MaxBet:  changeroom.RoomMaxBet,
			IsOpen:  changeroom.IsOpenRoom,
			TimeFmt: time.Now().Format("2006-01-02_15:04:05"),
		}
		InsertRoomStatus(roomInsert)
	}

}

// /api/setRoomLimitBet
// room_id (string) ex:"1","2"
// game_id (string) 游戏ID
func getRoomLimitBet(w http.ResponseWriter, r *http.Request) {
	var req GameRoomLimitBet
	var msg string = "查詢成功"
	req.Room = r.PostFormValue("room_id")
	req.GameId = r.PostFormValue("game_id")

	defer func() {
		js, err := json.Marshal(NewResp(SuccCode, msg, req))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}()

	if req.GameId != conf.Server.GameID {
		msg = "GameId输入错误"
		return
	}

	exist := false
	for _, v := range hall.roomList {
		common.Debug_log(v.RoomId, req.Room, v.RoomId == req.Room)
		if v != nil && v.RoomId == req.Room {
			req.MaxBet = common.Int32ToStr(v.RoomMaxBet)
			req.MinBet = common.Int32ToStr(v.RoomMinBet)
			common.Debug_log("房间%v 查詢限制下注:%v", v.RoomId, req)
			exist = true
		}
	}
	if !exist {
		msg = "无此房间"
		return
	}

}

func FormatTime(timeUnix int64, layout string) string {
	if timeUnix == 0 {
		return ""
	}
	format := time.Unix(timeUnix, 0).Format(layout)
	return format
}

func NewResp(code int, msg string, data interface{}) ApiResp {
	return ApiResp{Code: code, Msg: msg, Data: data}
}

func kickUser(w http.ResponseWriter, r *http.Request) {

	m1 := make(map[string]interface{})
	m1["msg"] = "succeed"

	uid := r.FormValue("id")
	uidNum, errU := strconv.Atoi(uid)
	if errU != nil {
		m1["msg"] = "非法用户ID"
	} else {
		kickUserInRoom(int32(uidNum))
	}

	b4, err := json.Marshal(m1)
	if err != nil {
		common.Debug_log("%v\n", err)
	}
	_, errW := fmt.Fprintf(w, "%+v", string(b4))
	if errW != nil {
		common.Debug_log("unlockUserMoneyUnexpected 返回错误 %v", errW)
	}

}

// API踢除玩家退款回大廳
func kickUserInRoom(userID int32) {
	common.Debug_log("gameModule kickUserInRoom")
	//检查用户是否已登陆
	// client, ok := AgentFromuserID[userID]
	client, ok := AgentFromuserID_.Load(userID)
	if ok {
		// 如果已经登陆过，需要通知之前登陆的用户被踢出游戏
		kickedBuf := &msg.KickedOutPush{
			ServerTime: time.Now().Unix(),
			Code:       0,
			Reason:     KICKOUT_DISABLE,
		}
		client.(*ClientInfo).agent.WriteMsg(kickedBuf)
		// delete(userIDFromAgent, client.agent)
		userIDFromAgent_.Delete(client.(*ClientInfo).agent)
		sendLogout(userID) // 踢人一起退資金
	}
}

func logoutUser(w http.ResponseWriter, r *http.Request) {

	m1 := make(map[string]interface{})
	m1["msg"] = "succeed"

	uid := r.FormValue("id")
	uidNum, errU := strconv.Atoi(uid)
	if errU != nil {
		m1["msg"] = "非法用户ID"
	} else {
		sendLogout(int32(uidNum))
	}

	b4, err := json.Marshal(m1)
	if err != nil {
		common.Debug_log("%v\n", err)
	}
	//common.Debug_log("S->C = %v", dt.Data)
	_, errW := fmt.Fprintf(w, "%+v", string(b4))
	if errW != nil {
		common.Debug_log("unlockUserMoneyUnexpected 返回错误 %v", errW)
	}

}

func unLockUserMoney(w http.ResponseWriter, r *http.Request) {

	m1 := make(map[string]interface{})
	m1["msg"] = "succeed"

	uid := r.FormValue("id")
	GameId := r.FormValue("game_id")
	unlockmoney := r.FormValue("money")

	uidNum, errU := common.Str2int32(uid)
	if errU != nil {
		m1["msg"] = "非法用户ID"
	}

	if GameId == "" || GameId != conf.Server.GameID { //檢測game_id
		log.Debug("game_id 有誤")
		return
	}

	unlockMoney, errU := strconv.Atoi(unlockmoney)
	if errU != nil {
		m1["msg"] = "非法用户ID"
	}

	AddTurnoverRecord("UserUnLockMoney", common.AmountFlowReq{
		UserID:    int32(uidNum),
		Money:     float64(unlockMoney),
		RoundID:   bson.NewObjectId().Hex(),
		Order:     bson.NewObjectId().Hex(),
		Reason:    "接口手動撤回投注解锁资金",
		TimeStamp: time.Now().Unix(),
	})
	m1["msg"] = uid + " unlockmoney " + unlockmoney + " succeed."

	b4, err := json.Marshal(m1)
	if err != nil {
		common.Debug_log("%v\n", err)
	}
	//common.Debug_log("S->C = %v", dt.Data)
	_, errW := fmt.Fprintf(w, "%+v", string(b4))
	if errW != nil {
		common.Debug_log("unlockUserMoneyUnexpected 返回错误 %v", errW)
	}

}

const (
	Reason_ZhiBo_Reward = "直播消费扣钱"
)

type ClientResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code" `
	Msg    struct {
		CreateTime       int64   `json:"create_time"`
		FinalBalance     float64 `json:"final_balance"`
		FinalLockBalance float64 `json:"final_lock_balance"`
		Order            string  `json:"order" `
	} `json:"msg"`
}

//用户直播赠礼扣款
func userZhiBoReward(w http.ResponseWriter, r *http.Request) {
	resp := &ApiResp{}
	defer func() {
		sendResponse(w, resp)
		common.Debug_log("<- FROM 直播赠礼扣款: %v", resp)
	}()
	uidStr := r.FormValue("user_id")
	amountStr := r.FormValue("amount")
	order := r.FormValue("order_id")
	//参数检查
	uid, errUid := strconv.Atoi(uidStr)
	amount, errAmount := strconv.ParseFloat(amountStr, 64)
	if errUid != nil || errAmount != nil || order == "" {
		resp.Code = -1
		resp.Msg = fmt.Sprintf("参数错误, uid: %s, amount: %s, order: %s", uidStr, amountStr, order)
		return
	}

	//查询用户信息
	player, ok := allUser_.Load(int32(uid))
	if !ok {
		resp.Code = -1
		resp.Msg = fmt.Sprintf("用户%d不存在", int32(uid))
		return
	}
	ClientAgent, ok2 := AgentFromuserID_.Load(int32(uid))
	if !ok2 {
		common.Debug_log("用户不在线,userID=%v\n", int32(uid))
		resp.Code = -1
		resp.Msg = fmt.Sprintf("用户%d不在线上", int32(uid))
		return
	}

	a := ClientAgent.(*ClientInfo).agent
	player.(*msg.PlayerInfo).Account -= amount
	p, ok := a.UserData().(*Player)
	if ok {
		if p.Account < amount {
			resp.Code = -1
			resp.Msg = "餘額不足扣款失败"
			return
		}
	}

	//请求中心服扣款信息
	payReq := PayRequest{
		Auth: AuthData{
			DevKey:  conf.Server.DevKey,
			DevName: conf.Server.DevName,
		},
		Info: PayData{
			GameID:     conf.Server.GameID,
			CreateTime: time.Now().Unix(),
			UserID:     int32(uid),
			RoundID:    order,
			OrderID:    bson.NewObjectId().Hex(),
			PayReason:  fmt.Sprintf("%s(单号:%s)", Reason_ZhiBo_Reward, order),
			Money:      amount,
		},
	}
	s, _ := json.Marshal(payReq)
	//向中心服发送扣款请求
	url := "http://" + conf.Server.CenterServer + ":" + common.IntToStr(conf.Server.CenterServerPort) + "/GameServer/GameUser/loseSettlement"
	payload := strings.NewReader(string(s))
	common.Debug_log("-> TO 直播赠礼扣款: %v %s", url, string(s))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		resp.Code = -1
		resp.Msg = "Error1: " + err.Error()
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err2 := client.Do(req)
	if err2 != nil {
		resp.Code = -1
		resp.Msg = "Error2 " + err2.Error()
		return
	}
	defer res.Body.Close()
	body, err3 := ioutil.ReadAll(res.Body)
	if err3 != nil {
		resp.Code = -1
		resp.Msg = "Error3: " + err3.Error()
		return
	}
	var clientResp ClientResponse
	err4 := json.Unmarshal(body, &clientResp)
	if err4 != nil {
		resp.Code = -1
		resp.Msg = "解析ClientResponse失败: " + err4.Error()
		return
	}
	resp.Data = string(body)
	if clientResp.Code != 200 {
		resp.Code = -1
		resp.Msg = "扣款失败"
	} else {
		resp.Code = 0
		resp.Msg = "扣款成功"

		// 更新玩家余额
		p.Account -= amount

		// 记录流水
		AddTurnoverRecord("ZhiBoGift", common.AmountFlowReq{
			UserID:    int32(uid),
			Money:     amount,
			RoundID:   order,
			Order:     bson.NewObjectId().Hex(),
			Reason:    "直播送礼扣款",
			TimeStamp: time.Now().Unix(),
		})

		// 玩家通知前端更新余额
		GiftPush := &msg.ZhiBoUpdateBalancePush{
			ServerTime: time.Now().Unix(),
			Code:       0,
			Balance:    p.Account - p.LockMoney,
			LockMoney:  p.LockMoney,
			GiftMoney:  amount,
			UserID:     int32(uid),
		}
		common.Debug_log("%v [%v]发送礼物扣款 礼物金额:%v 玩家余额:%v", uid, a, amount, GiftPush.Balance)
		a.WriteMsg(GiftPush)
	}
}

func sendResponse(w http.ResponseWriter, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		common.Debug_log("Resp err: %v", err)
		return
	}
	_, errW := fmt.Fprintf(w, "%+v", string(data))
	if errW != nil {
		common.Debug_log("Resp err: %v", errW)
	}
}

//流水相关数据
type PayData struct {
	UserID     int32   `json:"id,int"`
	CreateTime int64   `json:"create_time"`
	Money      float64 `json:"money"`
	BetMoney   float64 `json:"bet_money"`
	LockMoney  float64 `json:"lock_money"`
	PreMoney   float64 `json:"pre_money"`
	PayReason  string  `json:"pay_reason"`
	OrderID    string  `json:"order"` //自己创建一个唯一ID,方便之后查询
	GameID     string  `json:"game_id"`
	RoundID    string  `json:"round_id"` //唯一ID,用于识别多人是否在同一局游戏
}

//资金变动请求
type PayRequest struct {
	Auth AuthData `json:"auth"`
	Info PayData  `json:"info"`
}

type PayResponse struct {
	Balance            float64 `json:"balance"`
	BankerBalance      float64 `json:"banker_balance"`
	CreateTime         int64   `json:"create_time"`
	DevBrandName       string  `json:"dev_brand_name"`
	DevID              int     `json:"dev_id"`
	FinalBalance       float64 `json:"final_balance"`
	FinalBankerBalance float64 `json:"final_banker_balance"`
	FinalLockBalance   float64 `json:"final_lock_balance"`
	FinalPay           float64 `json:"final_pay"`
	FinalPrepayments   float64 `json:"final_prepayments"`
	GameAccountStatus  int     `json:"game_account_status"`
	GameID             string  `json:"game_id"`
	GameName           string  `json:"game_name"`
	GameNick           string  `json:"game_nick"`
	GameServerIP       string  `json:"game_server_ip"`
	GameUserStatus     int     `json:"game_user_status"`
	GameUserType       int     `json:"game_user_type"`
	ID                 int32   `json:"id"`
	Income             float64 `json:"income"`
	IsBanker           int     `json:"is_banker"`
	LockBalance        float64 `json:"lock_balance"`
	LockMoney          float64 `json:"lock_money"`
	LoginIP            string  `json:"login_ip"`
	Money              float64 `json:"money"`
	Order              string  `json:"order"`
	PackageID          int     `json:"package_id"`
	PlatformTaxPercent float64 `json:"platform_tax_percent"`
	Prepayments        float64 `json:"prepayments"`
	ProxyUserID        int     `json:"proxy_user_id"`
	RoundID            string  `json:"round_id"`
	Status             int     `json:"status"`
	Tax                float64 `json:"tax"`
	UUID               string  `json:"uuid"`
}

//用于验证的数据
type AuthData struct {
	//Token  string `json:"token"`
	DevKey  string `json:"dev_key"`
	DevName string `json:"dev_name"`
}

type OlUsersRes struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data OlUsersData `json:"data"`
}
type OlUsersData struct {
	GameID   string          `json:"game_id"`   // 游戏id
	GameName string          `json:"game_name"` //游戏名称
	GameData []OlUsersDetail `json:"game_data"` //各package_id该品牌的线上人数
}

type OlUsersDetail struct {
	PackageID int     `json:"packageID"` // 游戏id
	UsersList []int32 `json:"userData"`  //在线人数 去重后的uid 也就是查询时间段内的有记录的玩家ID
}

func getOnlineTotal(w http.ResponseWriter, r *http.Request) {

	data := &OlUsersRes{}
	var result = OlUsersData{}
	result.GameID = conf.Server.GameID
	result.GameName = "彩源猜大小"
	result.GameData = []OlUsersDetail{}
	PackageID := r.FormValue("package_id")

	// GameId := r.FormValue("game_id")

	defer func() {
		if len(data.Msg) > 0 {
			data.Code = -1
		} else {
			data.Data = result
		}
		bytes, err := json.Marshal(data)
		if err != nil {
			log.Error("json marshal error:%s", err.Error())
		} else {
			_, err := w.Write(bytes)
			if err != nil {
				log.Error("write getOnlineTotal result error:%s", err.Error())
			}
		}
	}()

	OLUsers.RLock()
	defer OLUsers.RUnlock()
	if PackageID != "" {
		PkgID := common.Str2Int(PackageID)
		if len(OnlineUsers[PkgID]) == 0 {
			return
		}
		userdata := OlUsersDetail{}
		userdata.PackageID = PkgID
		userdata.UsersList = OnlineUsers[PkgID]
		result.GameData = append(result.GameData, userdata)
	} else {
		for k, v := range OnlineUsers {
			if len(v) == 0 {
				continue
			}
			userdata := OlUsersDetail{}
			userdata.PackageID = k
			userdata.UsersList = v
			result.GameData = append(result.GameData, userdata)
		}
	}
	// if len(result.GameData) == 0 {
	// 	result.GameData = nil
	// }

}
