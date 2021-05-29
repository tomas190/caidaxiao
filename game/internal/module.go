package internal

import (
	common "caidaxiao/base"
	"math/rand"
	"time"

	"github.com/name5566/leaf/module"
)

var (
	skeleton = common.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	hall = NewHall()

	// c4c = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	packageTax = make(map[uint16]float64)

	InitMongoDB() // 初始连接数据库  //todo

	hall.Init() // 大厅初始化

	LoadServerSurpool() // 載入盈餘池

	LoadUserList() // 載入玩家列表
	// 中心服初始化并创建链接
	// c4c.Init()
	// c4c.CreatConnect()
	rand.Seed(time.Now().UnixNano())

	common.GetInstance().SetGameChanRpc(ChanRPC) //單例模式
	HeartBeatLoop()                              //檢設用戶心跳開始
}

func (m *Module) OnDestroy() {
	SaveServerConfig() //盈餘池資訊儲存
	SaveAllUserInfo()  //儲存所有人員資料
	LogoutAllUsers()   //登出所有用戶
	CloseGameServer()  //關閉GameServer(清理心跳.所有用戶登出)
}
