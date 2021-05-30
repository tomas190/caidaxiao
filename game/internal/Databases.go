package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"caidaxiao/msg"
	"time"

	"github.com/name5566/leaf/db/mongodb"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	session   *mgo.Session
	dbContext *mongodb.DialContext
)

const (
	dbName           = "caidaxiao-Game"
	playerInfo       = "playerInfo"
	settleWinMoney   = "settleWinMoney"
	settleLoseMoney  = "settleLoseMoney"
	accessDB         = "accessData"
	surPlusDB        = "surPlusDB"
	surPool          = "surplus-pool"
	robotData        = "robotData"
	PlayerDownBetDB  = "PlayerDownBetDB"
	PlayerGameDataDB = "PlayerGameDataDB"
	RoomTotalBetDB   = "RoomTotalBetDB"
	UserLimitBetDB   = "UserLimitBetDB"
)

// 连接数据库集合的函数 传入集合 默认连接IM数据库
func InitMongoDB() {
	// 此处连接正式线上数据库  下面是模拟的直接连接
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.Server.MongoDBAddr},
		Timeout:  60 * time.Second,
		Database: conf.Server.MongoDBAuth,
		Username: conf.Server.MongoDBUser,
		Password: conf.Server.MongoDBPwd,
	}

	var err error
	session, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal("Connect DataBase 数据库连接ERROR: %v ", err)
	}
	log.Debug("Connect DataBase 数据库连接SUCCESS~")

	//打开数据库
	session.SetMode(mgo.Monotonic, true)

	createUniqueIndex("PlayerDownBetDB", []string{"game_id"})
	createUniqueIndex("PlayerGameDataDB", []string{"user_id"})

}

func connect(dbName, cName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	c := s.DB(dbName).C(cName)
	return s, c
}

// func (p *Player) FirstPlayerInfo() bool {
// 	s, c := connect(dbName, playerInfo)
// 	defer s.Close()
// 	player := &msg.PlayerInfo{}
// 	player.Id = p.Id
// 	player.NickName = p.NickName
// 	player.HeadImg = p.HeadImg
// 	player.Account = p.Account

// 	err := c.Find(bson.M{"id": player.Id}).One(player)
// 	if err != nil {
// 		return true
// 	}
// 	return false
// }

func (p *Player) InsertPlayerInfo() {

	s, c := connect(dbName, playerInfo)
	defer s.Close()

	player := &msg.PlayerInfo{}
	player.Id = common.Int32ToStr(p.Id)
	player.NickName = p.NickName
	player.HeadImg = p.HeadImg
	player.Account = p.Account

	err := c.Insert(player)
	if err != nil {
		log.Debug("<----- 插入用户信息数据成功 ~ ----->")
	}

}

// 更新用戶資料。
func (p *Player) updateInfo(data common.UserInfo, playerInfo interface{}) {
	// common.Debug_log("gameModule *BaseUser updateInfo")
	playerInfo.(*msg.PlayerInfo).Id = common.Int32ToStr(data.UserID)
	playerInfo.(*msg.PlayerInfo).NickName = data.UserName
	playerInfo.(*msg.PlayerInfo).HeadImg = data.UserHead
	playerInfo.(*msg.PlayerInfo).Account = data.Balance
}

//LoadPlayerCount 获取玩家数量
func GetPlayerCount() int32 {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	n, err := c.Find(nil).Count()
	if err != nil {
		log.Debug("not Found Player Count, Maybe don't have Player")
		return 0
	}
	return int32(n)
}

type ChipDownBet struct {
	Chip1    int32
	Chip5    int32
	Chip10   int32
	Chip50   int32
	Chip100  int32
	Chip500  int32
	Chip1000 int32
}

type RobotDATA struct {
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

//InsertRobotData 插入机器人数据
func InsertRobotData(rb *RobotDATA) {
	s, c := connect(dbName, robotData)
	defer s.Close()

	err := c.Insert(rb)
	if err != nil {
		log.Debug("插入机器人数据失败:%v", err)
		return
	}
	log.Debug("插入机器人数据成功~")
}

//GetRobotData 获取机器人数据
func GetRobotData() ([]RobotDATA, error) {
	s, c := connect(dbName, robotData)
	defer s.Close()

	var rd []RobotDATA

	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).Unix()
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location()).Unix()

	selector := bson.M{}
	selector["room_time"] = bson.M{"$gte": startTime, "$lte": endTime}

	err := c.Find(selector).All(&rd)

	if err != nil {
		log.Debug("获取机器人数据失败:%v", err)
		return nil, err
	}
	log.Debug("获取机器人数据成功:%v,长度为:%v", rd, len(rd))
	return rd, nil
}

//InsertWinMoney 插入赢钱数据
func InsertWinMoney(base interface{}) {
	s, c := connect(dbName, settleWinMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 赢钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 赢钱结算数据插入成功 ~ ----->")
}

//InsertLoseMoney 插入输钱数据
func InsertLoseMoney(base interface{}) {
	s, c := connect(dbName, settleLoseMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 输钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 输钱结算数据插入成功 ~ ----->")
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id              int32             `json:"id" bson:"id"`                             // 玩家Id
	GameId          string            `json:"game_id" bson:"game_id"`                   // gameId
	RoundId         string            `json:"round_id" bson:"round_id"`                 // 随机Id
	RoomId          string            `json:"room_id" bson:"room_id"`                   // 所在房间
	DownBetInfo     *msg.DownBetMoney `json:"down_bet_info" bson:"down_bet_info"`       // 玩家各注池下注的金额
	DownBetTime     int64             `json:"down_bet_time" bson:"down_bet_time"`       // 下注时间
	StartTime       int64             `json:"start_time" bson:"start_time"`             // 开始时间
	EndTime         int64             `json:"end_time" bson:"end_time"`                 // 结束时间
	Lottery         []int             `json:"lottery" bson:"lottery"`                   // 开奖号码
	CardResult      *msg.PotWinList   `json:"card_result" bson:"card_result"`           // 当局开牌结果
	SettlementFunds float64           `json:"settlement_funds" bson:"settlement_funds"` // 当局输赢结果(税后)
	SpareCash       float64           `json:"spare_cash" bson:"spare_cash"`             // 剩余金额
	TaxRate         float64           `json:"tax_rate" bson:"tax_rate"`                 // 税率
	PeriodsNum      string            `json:"periods_num" bson:"periods_num"`           // 获奖期数
}

//InsertAccessData 插入运营数据接入
func InsertAccessData(data *PlayerDownBetRecode) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	//log.Debug("AccessData 数据: %v", data)
	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%v", err)
		return
	}
	//log.Debug("<----- 运营接入数据插入成功 ~ ----->")
}

//GetDownRecodeList 获取运营数据接入
func GetDownRecodeList(page, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	s, c := connect(dbName, accessDB)
	defer s.Close()
	var wts []PlayerDownBetRecode
	log.Debug("%v", selector)
	cmd := SearchCMD{
		DBName: dbName,
		CName:  accessDB,
		Query:  selector,
	}
	n := FindCountByQuery(cmd)
	// n, err := c.Find(selector).Count()
	// if err != nil {
	// return nil, 0, err
	// }
	log.Debug("获取 %v 条数据,limit:%v", n, limit)
	skip := page * limit
	err := c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

// 玩家游戏数据
type PlayerGameData struct {
	UserId          int32             `json:"user_id" bson:"user_id"`                   // 玩家Id
	RoomId          string            `json:"room_id" bson:"room_id"`                   // 所在房间
	DownBetInfo     *msg.DownBetMoney `json:"down_bet_info" bson:"down_bet_info"`       // 玩家各注池下注的金额
	DownBetTime     int64             `json:"down_bet_time" bson:"down_bet_time"`       // 下注时间
	StartTime       int64             `json:"start_time" bson:"start_time"`             // 开始时间
	EndTime         int64             `json:"end_time" bson:"end_time"`                 // 结束时间
	SettlementFunds float64           `json:"settlement_funds" bson:"settlement_funds"` // 当局输赢结果(税后)
	TotalWin        float64           `json:"total_win" bson:"total_win"`               // 玩家当局赢
	TotalLose       float64           `json:"total_lose" bson:"total_lose"`             // 玩家当局输
	PackageId       int               `json:"package_id" bson:"package_id"`             // PackageId
}

//InsertPlayerGame 玩家游戏数据
func InsertPlayerGame(data *PlayerGameData) {
	s, c := connect(dbName, PlayerGameDataDB)
	defer s.Close()

	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 玩家游戏数据插入失败 ~ ----->:%v", err)
		return
	}
}

//GetPlayerGameData 获取玩家游戏数据
func GetPlayerGameData(selector bson.M, limit int, sortBy string) ([]PlayerGameData, error) {
	s, c := connect(dbName, PlayerGameDataDB)
	defer s.Close()

	var wts []PlayerGameData

	log.Debug("获取玩家数据条件:%v", selector)
	err := c.Find(selector).Sort(sortBy).Limit(limit).All(&wts)
	if err != nil {
		log.Debug("获取玩家游戏数据:%v", err)
		return nil, err
	}
	return wts, nil
}

//GetPlayerWinData 获取玩家游戏数据
func GetPlayerWinData(selector bson.M) ([]PlayerGameData, error) {
	s, c := connect(dbName, PlayerGameDataDB)
	defer s.Close()

	var wts []PlayerGameData

	log.Debug("获取玩家数据条件:%v", selector)
	err := c.Find(selector).All(&wts)
	if err != nil {
		log.Debug("获取玩家游戏数据:%v", err)
		return nil, err
	}
	return wts, nil
}

//盈余池数据存入数据库
// type SurplusPoolDB struct {
// 	UpdateTime     time.Time
// 	TimeNow        string  //记录时间（分为时间戳/字符串显示）
// 	Rid            string  //房间ID
// 	TotalWinMoney  float64 //玩家当局总赢
// 	TotalLoseMoney float64 //玩家当局总输
// 	PoolMoney      float64 //盈余池
// 	HistoryWin     float64 //玩家历史总赢
// 	HistoryLose    float64 //玩家历史总输
// 	PlayerNum      int32   //历史玩家人数
// }

//InsertSurplusPool 插入盈余池数据
// func InsertSurplusPool(sur *SurplusPoolDB) {
// 	s, c := connect(dbName, surPlusDB)
// 	defer s.Close()

// 	//log.Debug("surplusPoolDB 数据: %v", sur.PoolMoney)

// 	err := c.Insert(sur)
// 	if err != nil {
// 		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
// 		return
// 	}
// 	//log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")

// 	SurPool := &SurPoolOld{}
// 	SurPool.GameId = conf.Server.GameID
// 	SurPool.SurplusPool = sur.PoolMoney
// 	SurPool.PlayerTotalLoseWin = sur.HistoryLose - sur.HistoryWin
// 	SurPool.PlayerTotalLose = sur.HistoryLose
// 	SurPool.PlayerTotalWin = sur.HistoryWin
// 	SurPool.TotalPlayer = sur.PlayerNum
// 	SurPool.FinalPercentage = 0.5
// 	SurPool.PercentageToTotalWin = 1
// 	SurPool.CoefficientToTotalPlayer = sur.PlayerNum * 0
// 	SurPool.PlayerLoseRateAfterSurplusPool = 0.7
// 	SurPool.DataCorrection = 0
// 	SurPool.PlayerWinRate = 0.6

// 	FindSurPool(SurPool)
// }

//FindSurplusPool
// func FindSurplusPool() *SurplusPoolDB {
// 	s, c := connect(dbName, surPlusDB)
// 	defer s.Close()

// 	sur := &SurplusPoolDB{}
// 	err := c.Find(nil).Sort("-updatetime").One(sur)
// 	if err != nil {
// 		log.Error("<----- 查找SurplusPool数据失败 ~ ----->:%v", err)
// 		return nil
// 	}
// 	return sur
// }

// type SurPoolOld struct {
// 	GameId                         string  `json:"game_id" bson:"game_id"`
// 	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
// 	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
// 	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
// 	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
// 	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
// 	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
// 	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
// 	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
// 	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
// 	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
// 	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
// }

// func FindSurPool(SurP *SurPoolOld) {
// 	s, c := connect(dbName, surPool)
// 	defer s.Close()

// 	sur := &SurPool{}
// 	err := c.Find(nil).One(sur)
// 	if err != nil {
// 		InsertSurPool(SurP)
// 	} else {
// 		SurP.SurplusPool = (SurP.PlayerTotalLose - (SurP.PlayerTotalWin * sur.PercentageToTotalWin) - float64(SurP.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
// 		SurP.FinalPercentage = sur.FinalPercentage
// 		SurP.PercentageToTotalWin = sur.PercentageToTotalWin
// 		SurP.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
// 		SurP.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
// 		SurP.DataCorrection = sur.DataCorrection
// 		SurP.PlayerWinRate = sur.PlayerWinRate
// 		UpdateSurPool(SurP)
// 	}
// }

//InsertSurPool 插入盈余池数据
// func InsertSurPool(sur *SurPoolOld) {
// 	s, c := connect(dbName, surPool)
// 	defer s.Close()

// 	//log.Debug("SurPool 数据: %v", sur)

// 	err := c.Insert(sur)
// 	if err != nil {
// 		log.Error("<----- 数据库插入SurPool数据失败 ~ ----->:%v", err)
// 		return
// 	}
// 	//log.Debug("<----- 数据库插入SurPool数据成功 ~ ----->")
// }

//UpdateSurPool 更新盈余池数据
// func UpdateSurPool(sur *SurPool) {
// 	s, c := connect(dbName, surPool)
// 	defer s.Close()

// 	err := c.Update(bson.M{}, sur)
// 	if err != nil {
// 		log.Error("<----- 更新 SurPool数据失败 ~ ----->:%v", err)
// 		return
// 	}
// 	//log.Debug("<----- 更新SurPool数据成功 ~ ----->")
// }

//GetDownRecodeList 获取盈余池数据
// func GetSurPoolData(selector bson.M) (SurPool, error) {
// 	s, c := connect(dbName, surPool)
// 	defer s.Close()

// 	var wts SurPool

// 	err := c.Find(selector).One(&wts)
// 	if err != nil {
// 		return wts, err
// 	}
// 	return wts, nil
// }

type PlayerDownBet struct {
	Id          int32             `json:"id" bson:"id"`                       // 玩家Id
	GameId      string            `json:"game_id" bson:"game_id"`             // gameId
	RoomId      string            `json:"room_id" bson:"room_id"`             // 所在房间
	PeriodsNum  string            `json:"periods_num" bson:"periods_num"`     // 奖期
	PeriodsTime string            `json:"periods_time" bson:"periods_time"`   // 奖期时间
	LotteryType string            `json:"lottery_type" bson:"lottery_type"`   // 彩种
	DownBetInfo *msg.DownBetMoney `json:"down_bet_info" bson:"down_bet_info"` // 玩家各注池下注的金额
	DownBetTime string            `json:"down_bet_time" bson:"down_bet_time"` // 下注时间
}

func InsertPlayerDownBet(sur *PlayerDownBet) {
	s, c := connect(dbName, PlayerDownBetDB)
	defer s.Close()

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入PlayerDownBet数据失败 ~ ----->:%v", err)
		return
	}
}

//GetPlayerDownBet 获取玩家投注数据
func GetPlayerDownBet(page, limit int, selector bson.M, sortBy string) ([]PlayerDownBet, int, error) {
	s, c := connect(dbName, PlayerDownBetDB)
	defer s.Close()

	var wts []PlayerDownBet

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	log.Debug("获取 %v 条数据,limit:%v", n, limit)
	skip := (page - 1) * limit
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

type RoomTotalBet struct {
	GameId        string            `json:"game_id" bson:"game_id"`             // gameId
	RoomId        string            `json:"room_id" bson:"room_id"`             // 所在房间
	PeriodsNum    string            `json:"periods_num" bson:"periods_num"`     // 奖期
	PeriodsTime   string            `json:"periods_time" bson:"periods_time"`   // 奖期时间
	LotteryType   string            `json:"lottery_type" bson:"lottery_type"`   // 彩种
	PotTotalMoney *msg.DownBetMoney `json:"down_bet_info" bson:"down_bet_info"` // 注池玩家下注总金额
}

func InsertRoomTotalBet(sur *RoomTotalBet) {
	s, c := connect(dbName, RoomTotalBetDB)
	defer s.Close()

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入RoomTotalBet数据失败 ~ ----->:%v", err)
		return
	}
}

//GetPlayerDownBet 获取玩家投注数据
func GetRoomTotalBet(page, limit int, selector bson.M, sortBy string) ([]RoomTotalBet, int, error) {
	s, c := connect(dbName, RoomTotalBetDB)
	defer s.Close()

	var wts []RoomTotalBet

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	log.Debug("获取 %v 条数据,limit:%v", n, limit)
	skip := (page - 1) * limit
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

//InsertUserLimitBet 设定玩家下注限红
func InsertUserLimitBet(sur *GameLimitBet) {
	s, c := connect(dbName, UserLimitBetDB)
	defer s.Close()

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入UserLimitBet数据失败 ~ ----->:%v", err)
		return
	}
}

//LoadUserLimitBet 获取玩家限制下注数据
func LoadUserLimitBet(player *Player) GameLimitBet {
	s, c := connect(dbName, UserLimitBetDB)
	defer s.Close()

	game := &GameLimitBet{}
	err := c.Find(bson.M{"user_id": player.Id}).Sort("-time_fmt").One(game)
	if err != nil {
		log.Error("<----- 数据库读取LoadUserLimitBet数据失败 ~ ----->:%v", err)
	}
	return *game
}

//GetUserLimitBet 获取玩家下注限红数据
func GetUserLimitBet(selector bson.M) ([]GameLimitBet, int, error) {
	s, c := connect(dbName, UserLimitBetDB)
	defer s.Close()

	var wts []GameLimitBet

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}

	err = c.Find(selector).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

func createUniqueIndex(cName string, keys []string) {
	// DBSession := dbContext.Ref()
	// defer dbContext.UnRef(DBSession)
	col := session.DB(dbName).C(cName)
	// 設定統計表唯一索引
	index := mgo.Index{
		Key:        keys,  // 索引鍵
		Background: true,  // 不长时间占用写锁
		Unique:     false, // 唯一索引
		DropDups:   true,  // 存在資料後建立, 則自動刪除重複資料
	}

	err := col.EnsureIndex(index)
	if err != nil {
		log.Debug("mgo建立index錯誤:%v", err)
	} else {
		log.Debug("mgo建立index:%v %v", cName, index)
	}
}

func BulkUpdateAll(cmd SearchCMD, pairs []interface{}) {
	// session := dbContext.Ref()
	// defer dbContext.UnRef(session)

	bulk := session.DB(cmd.DBName).C(cmd.CName).Bulk()
	bulk.Update(pairs...)
	_, err := bulk.Run()
	if err != nil {
		log.Debug("%v批量更新 %v", cmd, err.Error())

	}
}
