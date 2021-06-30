package internal

import (
	common "caidaxiao/base"
	"caidaxiao/conf"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	logging "github.com/sacOO7/go-logger"

	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/timer"
	"github.com/sacOO7/gowebsocket"
)

const (
	// Time allowed to write a message to the peer.

	pingPeriod = 5 * time.Second

	pongWait = 60 * time.Second

	reconnectWait = 2 * time.Second
)

// 防止并发写Websocket用的锁
var syncWrite sync.Mutex

// 取token用
var intervalTimer *timer.Timer

//與中心服websocket連線   設置config
func (c4c *Conn4Center) S2CS_Conn_init() {
	// common.Debug_log("login *Conn4Center initSBClient")
	c4c.hasInit = false
	c4c.hasLogin = false
	c4c.hasReConnected = false
	c4c.devKey = conf.Server.DevKey
	c4c.devName = conf.Server.DevName
	c4c.gameId = conf.Server.GameID
	c4c.wsAddress = conf.Server.CenterUrl
	c4c.httpAddress = "http://" + conf.Server.CenterServer + ":" + common.IntToStr(conf.Server.CenterServerPort)
	c4c.localHost = strings.Split(conf.Server.WSAddr, ":")[0]
	c4c.localPort, _ = strconv.Atoi(strings.Split(conf.Server.WSAddr, ":")[1])
	// c4c.token = ""
	// c4c.tokenUrl = fmt.Sprintf("%s/Token/getToken?dev_key=%s&dev_name=%s", c4c.httpAddress, c4c.devKey, c4c.devName)
	c4c.S2CS_connect() //與中心服websocket連接(未登入)
}

// 與中心服建立ws連線
func (c4c *Conn4Center) S2CS_connect() {
	// common.Debug_log("loginModule *Conn4Center S2CS_connect")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM) //signal.Notify(interrupt, syscall.SIGTERM)

	socket := gowebsocket.New(c4c.wsAddress)

	//websocket loglevel set
	socket.EnableLogging()
	socket.GetLogger().SetLevel(logging.WARNING)

	// 與中心服連線錯誤
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		common.Debug_log("Received connect error ", err)
	}

	// 與中心服連線成功
	socket.OnConnected = func(socket gowebsocket.Socket) {
		if c4c.hasInit {
			common.Debug_log("Reconnected to center server")
		} else {
			c4c.hasInit = true
			common.Debug_log("connected to center server")
		}
		c4c.socket = socket
		c4c.S2CS_login()
	}

	// 收到中心服訊息
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		// common.Debug_log("<- From CENTER %s", message)
		receiveMessage(message)
	}

	// Received PING from server
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		common.Debug_log("Received ping - " + data)
	}

	// Received PONG from server 中心服回傳 pong
	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		common.Debug_log("Received pong - " + data)
		// 將DeadLine增加時間
		if err := socket.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			common.Debug_log("socket write read dead line err", err)
		}
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		common.Debug_log("Disconnected from center server ")
		// 當中心服與子服務斷線後要處理什麼(還未確定如何處理)
		// c4c.hasLogin = false
		// common.GetInstance().Game.Go("UpdateService", false)
	}

	//建一個goroutine去偵測是否需要斷線重連
	go func() {
		pTicker := time.NewTicker(pingPeriod)
		rTicker := time.NewTicker(reconnectWait)
		defer func() {
			pTicker.Stop()
			rTicker.Stop()
		}()

		for {
			select {
			case <-rTicker.C: // 與中心服Ws斷線重連
				if !socket.IsConnected {
					common.Debug_log("reconnect to center server")
					c4c.hasLogin = false
					c4c.hasReConnected = true
					socket.Connect()
				}
			case <-pTicker.C: //與中心服建立連線後開始心跳
				if !socket.IsConnected {
					break
				}
				syncWrite.Lock()
				if err := socket.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					common.Debug_log("websocket send ping to center err", err.Error())
					socket.Close()
				}
				syncWrite.Unlock()
			case sig := <-interrupt:
				// socket.Close()
				common.Debug_log("close server,signal:%s", sig.String())
				err := syscall.Kill(os.Getpid(), syscall.SIGINT) //殺掉子進程的只有在MACOS以及LINUX可以跑
				if err != nil {
					common.Debug_log("kill process err", err.Error())
				}
				return
			}
		}
	}()
	socket.Connect()
}

// 取token (目前寫死)
func (c4c *Conn4Center) handlerGetToken() {
	// common.Debug_log("loginModule *Conn4Center handlerGetToken")
	if intervalTimer != nil {
		intervalTimer.Stop()
		intervalTimer = nil
	}
	// bridge.LogC("请求token,url:%s", c4c.tokenUrl)
	// tokenResp := &TokenResponse{}
	// response, err := http.Get(c4c.tokenUrl)
	// //程序在使用完回复后必须关闭回复的主体。
	// defer func() {
	// 	if response != nil {
	// 		_ = response.Body.Close()
	// 	}
	delay := time.Second * 2
	// 	if tokenResp != nil && tokenResp.Code == 200 {
	delay = tokenWait
	c4c.token = "3424jkjfkjs9f9d8" //tokenResp.Msg.Token
	// 		bridge.LogC("token data:%+v", tokenResp.Msg)
	if c4c.hasInit {
		c4c.S2CS_login()
	} else {
		c4c.S2CS_connect()
	}
	// 	}
	intervalTimer = skeleton.AfterFunc(delay, c4c.handlerGetToken)
	// }()

	// if err != nil {
	// 	bridge.LogC("token request failed", err)
	// 	return
	// }
	// body, _ := ioutil.ReadAll(response.Body)
	// err1 := json.Unmarshal(body, tokenResp)
	// if err1 == nil {
	// 	fmt.Println("Unmarshal error:", err1)
	// }
}

// 子服務登陸中心服
func (c4c *Conn4Center) S2CS_login() {
	// common.Debug_log("loginModule *Conn4Center S2CS_login")
	if c4c.hasLogin { //是否已登錄
		return
	}
	req := S2CS_Login{
		Host:    c4c.localHost,
		Port:    c4c.localPort,
		GameID:  c4c.gameId,
		DevKey:  c4c.devKey,
		DevName: c4c.devName,
		//Token:  c4c.token,
	}
	//bridge.LogC(req)
	c4c.sendMessage(MSG_SERVER_LOGIN, req)
	common.Debug_log("登陆中心服务器...")
}

// 發送到中心服
func (c4c *Conn4Center) sendMessage(name string, data interface{}) {
	// common.Debug_log("loginModule *Conn4Center sendMessage")
	if !c4c.socket.IsConnected {
		return
	}
	req := &S2CS_Message{
		Event: name,
		Data:  data,
	}
	chars, err := json.Marshal(req)
	if err != nil {
		common.Debug_log("json marshal failed :%s", err)
		return
	}
	message := string(chars)
	syncWrite.Lock()
	defer syncWrite.Unlock()
	common.Debug_log("To CENTER ->: %s", message)
	// Debug_log("发送到中心服：Event=%s ; data=%+v", name, data)
	c4c.socket.SendText(message)
	// 下面是新增紀錄
	// common.GetInstance().Game.Go("SendToCenter", name, data) //counter receiveToCenterMessage

}

// 處理中心服回傳數據
func receiveMessage(msg string) {
	// common.Debug_log("loginModule receiveMessage")
	go func(msg string) {

		common.Debug_log("<- FROM CENTER : %+v", msg)
		rsp := CS2S_Message{}
		err := json.Unmarshal([]byte(msg), &rsp)
		if err != nil {
			common.Debug_log("解析接收数据失败,err=%+v", err)
			return
		}
		// common.Debug_log("从中心服返回：Event=%s ；Data=%+v", rsp.Event, rsp.Data)
		// common.GetInstance().Game.Go("RecordFromCenter", rsp.Event, rsp.Data) //記錄每筆中心服回傳的訊息    之後要做
		switch rsp.Event {
		case MSG_ERROR:
			handlerError(rsp.Data)
		case MSG_SERVER_LOGIN:
			handlerLogin(rsp.Data)
		case MSG_USER_LOGIN:
			CS2C_Login(rsp.Data)
		case MSG_USER_LOGOUT:
			CS2C_Logout(rsp.Data)
		case MSG_USER_LOCK_MONEY:
			CS2C_LockSettlement(rsp.Data)
		case MSG_USER_UNLOCK_MONEY:
			CS2C_UnLockSettlement(rsp.Data)
		case MSG_USER_WIN_MONEY:
			CS2C_WinSettlement(rsp.Data)
		case MSG_USER_LOSE_MONEY:
			CS2C_LoseSettlement(rsp.Data)
		}

	}(msg)
}

// 中心服回傳錯誤訊息
func handlerError(data ResponseData) {
	if data.Code != 0 {
		common.Debug_log("解析登陆中心服务器返回数据失败,err=%+v", data.Code)
		return
	}
	if data.Code == 501 {
		// token相关错误，需要重新请求token
		c4c.handlerGetToken()
	}
}

// 中心服回傳登陸訊息
func handlerLogin(data ResponseData) {
	if c4c.hasReConnected {
		common.GetInstance().Game.Go("UpdateService", true)
		return
	}
	c4c.hasLogin = true
	jsonStr, err1 := json.Marshal(data.Msg["globals"])
	if err1 != nil {
		common.Debug_log("解析登陆中心服务器返回数据失败,err=%+v", err1)
		return
	}

	var info []common.LoginResponse
	err2 := json.Unmarshal([]byte(jsonStr), &info)
	if err2 != nil {
		common.Debug_log("Json序列化登陆中心服务器返回数据失败,err=%+v", err2)
		return
	}
	common.Debug_log("中心服务器登陆成功")
	skeleton.AfterFunc(3*time.Second, func() {
		common.GetInstance().Game.Go("StartServer", info)
	})

}
