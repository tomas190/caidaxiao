package internal

import (
	// "github.com/name5566/leaf/log"

	"time"

	"github.com/sacOO7/gowebsocket"
)

const (
	//中心服指令定義
	MSG_SERVER_LOGIN      = "/GameServer/Login/login"
	MSG_USER_LOGIN        = "/GameServer/GameUser/login"
	MSG_USER_LOGOUT       = "/GameServer/GameUser/loginout"
	MSG_USER_LOCK_MONEY   = "/GameServer/GameUser/lockSettlement"
	MSG_USER_UNLOCK_MONEY = "/GameServer/GameUser/unlockSettlement"
	MSG_USER_WIN_MONEY    = "/GameServer/GameUser/winSettlement"
	MSG_USER_LOSE_MONEY   = "/GameServer/GameUser/loseSettlement"
	MSG_NOTICE            = "/GameServer/Notice/notice"
	MSG_ERROR             = "error"
	//token 等待時間
	tokenWait = 7000 * time.Second
)

//與中心服連線
type Conn4Center struct {
	devKey         string
	devName        string
	gameId         string
	token          string
	wsAddress      string
	httpAddress    string
	tokenUrl       string
	localHost      string
	localPort      int
	hasInit        bool //是否已初始化
	hasLogin       bool //是否已登錄
	hasReConnected bool
	socket         gowebsocket.Socket
}

// S -> 子服務 ,CS -> 中心服 CenterServer
//子服務登陸中心服request
type S2CS_Login struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	GameID  string `json:"game_id"`
	DevKey  string `json:"dev_key"`
	DevName string `json:"dev_name"`
}

//子服務登陸中心服Response
type CS2S_Login struct {
	TaxPercent int32 `json:"platform_tax_percent"`
}

// 子服務對中心服基本消息结构 Event:中心服指令定義
type S2CS_Message struct {
	Event string      `json:"event"` // 事件
	Data  interface{} `json:"data"`  // 数据
}

// 子服務對中心服基本消息结构 Event:中心服指令定義
type CS2S_Message struct {
	Event string       `json:"event"`
	Data  ResponseData `json:"data"`
}

//中心服回傳的資料
type ResponseData struct {
	Status string                 `json:"status"`
	Code   int32                  `json:"code"`
	Error  string                 `json:"error"`
	Msg    map[string]interface{} `json:"msg"`
}

//token詳細資料 (現在沒有使用，之後可能需要動態取得使用)
type TokenInfo struct {
	LeftTime     int32  `json:"left_times"`
	DevBrandName string `json:"dev_brand_name"`
	Status       int32  `json:"status"`
	DevID        int32  `json:"dev_id"`
	Token        string `json:"token"`
	IP           string `json:"ip"`
	DevKey       string `json:"dev_key"`
	DevName      string `json:"dev_name"`
	CreateTime   int32  `json:"create_time"`
	Expire       int32  `json:"expire"`
	History      string `json:"history"`
	EndTime      int32  `json:"end_time"`
}

// token回復 (現在沒有使用，之後可能需要動態取得使用)
type TokenResponse struct {
	Status int32      `json:"status"`
	Code   int32      `json:"code"`
	Msg    *TokenInfo `json:"msg"`
}
