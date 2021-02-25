package internal

import (
	"caidaxiao/conf"
	"caidaxiao/msg"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	session *mgo.Session
)

const (
	dbName          = "caidaxiao-Game"
	playerInfo      = "playerInfo"
	roomSettle      = "roomSettle"
	settleWinMoney  = "settleWinMoney"
	settleLoseMoney = "settleLoseMoney"
	accessDB        = "accessData"
	surPlusDB       = "surPlusDB"
	surPool         = "surplus-pool"
	playerGameData  = "playerGameData"
	robotData       = "robotData"
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
}

func connect(dbName, cName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	c := s.DB(dbName).C(cName)
	return s, c
}

func (p *Player) FindPlayerInfo() {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	player := &msg.PlayerInfo{}
	player.Id = p.Id
	player.NickName = p.NickName
	player.HeadImg = p.HeadImg
	player.Account = p.Account

	err := c.Find(bson.M{"id": player.Id}).One(player)
	if err != nil {
		err2 := InsertPlayerInfo(player)
		if err2 != nil {
			log.Error("<----- 插入用户信息数据失败 ~ ----->:%v", err)
			return
		}
		log.Debug("<----- 插入用户信息数据成功 ~ ----->")
	}
}

func InsertPlayerInfo(player *msg.PlayerInfo) error {
	s, c := connect(dbName, playerInfo)
	defer s.Close()

	err := c.Insert(player)
	return err
}

//LoadPlayerCount 获取玩家数量
func LoadPlayerCount() int32 {
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
