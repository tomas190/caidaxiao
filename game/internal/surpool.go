package internal

import (
	common "caidaxiao/base"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

// var(
// dbContext  *mongodb.DialContext
// 	dbContext = context
// )

type SurPool struct {
	ID        bson.ObjectId `bson:"_id"`
	TotalLost float64       `bson:"player_total_lose"` // 全部历史总输 正值
	TotalWin  float64       `bson:"player_total_win"`  // 全部历史总赢 正值

	SumUser              float64 `bson:"total_player"`                        // 历史实际的玩家总数字段
	UserLostMinusWin     float64 `bson:"player_total_lose_win"`               // 玩家总输-总赢
	KillPercent          float64 `bson:"percentage_to_total_win"`             // 杀数，目前100%
	PoolBalance          float64 `bson:"surplus_pool"`                        //奖金池余额 可正可负
	MoneyPrizeOneUser    float64 `bson:"coefficient_to_total_player"`         //玩家赠送金额
	FinalPercentage      float64 `bson:"final_percentage"`                    //0.5
	LoseRateAfterSurplus float64 `bson:"player_lose_rate_after_surplus_pool"` //0.964
	DataCorrection       float64 `bson:"data_correction"`                     // 誤差修正值(盈餘池異常時修正用)

	CountAfterWin       float64 `bson:"random_count_after_win"`       // 玩家贏錢重新開獎次數
	PercentageAfterWin  float64 `bson:"random_percentage_after_win"`  // 玩家贏錢重新開獎機率
	CountAfterLose      float64 `bson:"random_count_after_lose"`      // 玩家輸錢重新開獎次數
	PercertageAfterLose float64 `bson:"random_percentage_after_lose"` // 玩家輸錢重新開獎機率

	TaxPercent float64 `bson:"-"` // 平台扣税比例
	UpdateTime string  `bson:"-"`

	AgentNum int32 `bson:"-"` // 目前線上玩家連線數
}

type SearchCMD struct {
	DBName    string        //数据库名称
	CName     string        //数据表名称
	SortField string        //排序条件
	LenLimit  int           //数量限制
	ItemID    bson.ObjectId //数据ID
	Query     interface{}   //查询条件
	Update    interface{}   //更新内容
	Skip      int           //数量起始偏移值
}

var (
	ServerSurPool = &SurPool{}
	mutexPool     = new(sync.RWMutex)
)

func LoadServerSurpool() {
	// s, c := connect(dbName, surPool)
	// defer s.Close()

	// sur := make([]*SurPool, 0)
	// err := c.Find(nil)
	// if err != nil {
	// 	log.Debug("<----- 查找SurplusPool数据失败 ~ ----->:%v", err)

	// }

	cmd := SearchCMD{
		DBName: dbName,
		CName:  surPool, // "SERVER",
	}
	sur := make([]*SurPool, 0)
	ok := FindAllItems(cmd, &sur)
	if !ok {
		log.Debug("查找失败 %v", ok)
		return
	}

	if len(sur) == 1 {
		ServerSurPool = sur[0]
		ServerSurPool.SumUser = float64(GetPlayerCount())
		log.Debug("Release : 服务器配置加载成功 MongoDBName: %v", dbName)
		log.Debug("Release : 盈餘池數據: %v", ServerSurPool)

	} else { // 沒有配置過或者資料超過1筆

		if len(sur) > 1 {
			cmd := SearchCMD{
				DBName: dbName,
				CName:  surPool, // "SERVER",
				Query:  bson.M{"final_percentage": 0.5},
			}
			RemoveItemsByQuery(cmd)
		}

		ServerSurPool = makeServerConfig()
		cmd := SearchCMD{
			DBName: dbName,
			CName:  surPool, // "SERVER",
		}
		ok := AddOneItemRecord(cmd, ServerSurPool)
		if !ok {
			log.Debug("Error : 写入服务器配置数据出错")

			return
		}
		SaveServerConfig()
		log.Debug("Release : 初始化服务器配置并写入成功")
	}
	ServerSurPool.AgentNum = 0 // 重啟 agent 歸0

}

// 初始化serverConfig
func makeServerConfig() *SurPool {
	log.Debug("gameModule makeServerConfig")

	// 首次佈署
	return &SurPool{
		ID:                   bson.NewObjectId(),
		PoolBalance:          0,
		KillPercent:          1,
		MoneyPrizeOneUser:    0,
		FinalPercentage:      0.5,
		LoseRateAfterSurplus: 0.7,
		DataCorrection:       0,
		CountAfterWin:        0,
		PercentageAfterWin:   0,
		CountAfterLose:       0,
		PercertageAfterLose:  0,
	}

}

func SaveServerConfig() { //更新盈餘池表
	// common.Debug_log("gameModule SaveServerConfig")
	ServerSurPool.UpdateTime = common.TimeFormatDate(time.Now().Unix())
	cmd := SearchCMD{
		DBName: dbName,
		CName:  surPool,
		ItemID: ServerSurPool.ID,
		Update: bson.M{"$set": ServerSurPool},
	}
	ok := UpdateItemByID(cmd)
	if !ok {
		common.Debug_log("Error : 更新服务器配置数据出错 ID:%v", cmd.ItemID)
	}
}

// moneyOffset 可正可负
func (item *SurPool) updatePoolBalance() {
	// common.Debug_log("gameModule  updatePoolBalance")
	mutexPool.Lock()
	defer mutexPool.Unlock()
	// if moneyOffset < 0 {
	// 	item.TotalLost -= moneyOffset // 保持正值
	// } else if moneyOffset > 0 {
	// 	item.TotalWin += moneyOffset
	// }
	// if moneyOffset != 0 {
	itemPoolBalance := (item.TotalLost - item.TotalWin*item.KillPercent - item.SumUser*item.MoneyPrizeOneUser + item.DataCorrection) * item.FinalPercentage // - item.SumUser * 6
	// itemPoolBalance := (item.TotalLost - item.TotalWin*float64(item.KillPercent)) * item.FinalPercentage // - item.SumUser * 6
	item.PoolBalance, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", itemPoolBalance), 64)
	item.UserLostMinusWin = item.TotalLost - item.TotalWin
	SaveServerConfig()
	// common.Debug_log("item.PoolBalance:%v item.TotalLost:%v item.TotalWin:%v", item.PoolBalance, item.TotalLost, item.TotalWin)

	// }
}
