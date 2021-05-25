package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	RoomId     string `form:"room_id" json:"room_id"`
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
	// 获取彩源玩家投注数据
	http.HandleFunc("/api/getPlayerDownBet", getPlayerDownBet)
	// 获取彩源房间投注统计
	http.HandleFunc("/api/getRoomTotalBet", getRoomTotalBet)
	// 接口操作关闭或开启房源
	http.HandleFunc("/api/HandleRoomType", HandleRoomType)
	// 分分彩包赔活动（河内分分彩）
	http.HandleFunc("/api/HandleHeNeiPay", HandleHeNeiPay)
	// 分分彩连赢活动（河内分分彩）
	http.HandleFunc("/api/HandleHeNeiWin", HandleHeNeiWin)
	// 分分彩连赢活动（奇趣分分彩）
	http.HandleFunc("/api/HandleQiQuWin", HandleQiQuWin)
	// 设定玩家下注限紅
	http.HandleFunc("/api/setUserLimitBet", setUserLimitBet)
	// 获取玩家下注限紅
	http.HandleFunc("/api/getUserLimitBet", getUserLimitBet)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq

	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.RoomId = r.FormValue("room_id")
	req.RoundId = r.FormValue("round_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

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
	}

	if sTime != 0 && eTime == 0 {
		selector["start_time"] = bson.M{"$gte": sTime}
	}

	if eTime != 0 && sTime == 0 {
		selector["end_time"] = bson.M{"$lte": eTime}
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetDownRecodeList(page, limits, selector, "-down_bet_time")
	if err != nil {
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
		gd.RoomId = pr.RoomId
		gd.RoundId = pr.RoundId
		gd.Lottery = pr.Lottery
		gd.BetInfo = *pr.DownBetInfo
		gd.Card = *pr.CardResult
		gd.SettlementFunds = pr.SettlementFunds
		gd.SpareCash = pr.SpareCash
		gd.TaxRate = pr.TaxRate
		gd.CreatedAt = pr.DownBetTime
		gd.PeriodsNum = pr.PeriodsNum
		gameData = append(gameData, gd)
	}

	var result pageData
	result.Total = count
	result.List = gameData

	//fmt.Fprintf(w, "%+v", ApiResp{Code: SuccCode, Msg: "", Data: result})
	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
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
	Id := r.FormValue("id")
	log.Debug("reqPlayerLeave 踢出玩家:%v", Id)
	rid, _ := hall.UserRoom.Load(Id)
	v, _ := hall.RoomRecord.Load(rid)
	if v != nil {
		room := v.(*Room)
		user, _ := hall.UserRecord.Load(Id)
		if user != nil {
			p := user.(*Player)
			room.IsConBanker = false
			hall.UserRecord.Delete(p.Id)
			p.PlayerExitRoom()
			c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			leaveHall := &msg.Logout_S2C{}
			p.SendMsg(leaveHall, "Logout_S2C")

			js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
			if err != nil {
				fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
				return
			}
			w.Write(js)
		}
	}
}

func getRobotData(w http.ResponseWriter, r *http.Request) {
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
	req.RoomId = r.FormValue("room_id")
	req.PeriodsNum = r.FormValue("periods_num")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	selector := bson.M{}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoomId != "" {
		selector["room_id"] = req.RoomId
	}

	if req.PeriodsNum != "" {
		selector["periods_num"] = req.PeriodsNum
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetPlayerDownBet(page, limits, selector, "-down_bet_time")
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

func getRoomTotalBet(w http.ResponseWriter, r *http.Request) {
	var req CaiYuanReq

	req.GameId = r.FormValue("game_id")
	req.RoomId = r.FormValue("room_id")
	req.PeriodsNum = r.FormValue("periods_num")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	selector := bson.M{}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoomId != "" {
		selector["room_id"] = req.RoomId
	}

	if req.PeriodsNum != "" {
		selector["periods_num"] = req.PeriodsNum
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetRoomTotalBet(page, limits, selector, "-down_bet_time")
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
	req.RoomId = r.FormValue("room_id")
	req.IsOpen = r.FormValue("room_status")

	log.Debug("RoomId:%v, IsOpen:%v", req.RoomId, req.IsOpen)

	if req.RoomId == "01" {
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
				}
			}
		}
	}

	if req.RoomId == "02" {
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
				}
			}
		}
	}

	data := &msg.ChangeRoomType_S2C{}
	for _, v := range hall.roomList {
		if v != nil {
			if v.RoomId == "1" {
				data.Room01 = v.IsOpenRoom
			}
			if v.RoomId == "2" {
				data.Room02 = v.IsOpenRoom
			}
		}
	}

	// 发送给所有玩家
	hall.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*Player)
		u.SendMsg(data, "ChangeRoomType_S2C")
		return true
	})

	js, err := json.Marshal(NewResp(SuccCode, "", ""))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HandleHeNeiPay(w http.ResponseWriter, r *http.Request) {
	var req GamePayReq

	req.UserId = r.FormValue("user_id")
	req.MinBet = r.FormValue("min_bet")
	req.MaxBet = r.FormValue("max_bet")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Limit = r.FormValue("limit")

	log.Debug("河内赔付数据:%v", req)

	selector := bson.M{}

	if req.UserId != "" {
		selector["user_id"] = req.UserId
	}

	minBet, _ := strconv.Atoi(req.MinBet)

	maxBet, _ := strconv.Atoi(req.MaxBet)

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime == 0 || eTime == 0 {
		startTime := time.Now().Unix()
		currentTime := time.Now()
		oldTime := currentTime.AddDate(0, 0, -7)
		endTime := oldTime.Unix()
		selector["down_bet_time"] = bson.M{"$gte": endTime, "$lte": startTime}
	}

	limits, _ := strconv.Atoi(req.Limit)
	if limits == 0 {
		limits = 10
	}

	recodes, err := GetPlayerGameData(selector, limits, "-down_bet_time")
	log.Debug("获取数据:%v", recodes)

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
			num++
		}
		if num > 1 {
			continue
		}
		downBet := v.DownBetInfo.BigDownBet + v.DownBetInfo.SmallDownBet + v.DownBetInfo.LeopardDownBet
		if downBet < int32(minBet) || downBet > int32(maxBet) {
			continue
		}
		data.GameCount++
		if v.SettlementFunds > 0 {
			data.TotalWin += v.SettlementFunds
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
		pac := packageTax[v.PackageId]
		taxR := pac / 100
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
		pac := packageTax[v.PackageId]
		taxR := pac / 100
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

	log.Debug("限制玩家下注:%v", req)
	minBet, _ := strconv.Atoi(req.MinBet)
	maxBet, _ := strconv.Atoi(req.MaxBet)

	hall.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*Player)
		if u.Id == req.UserId {
			log.Debug("玩家id:%v,限制:%v,%v", u.Id, minBet, maxBet)
			u.MinBet = int32(minBet)
			u.MaxBet = int32(maxBet)
		}
		return true
	})

	if req.UserId != "" {
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
