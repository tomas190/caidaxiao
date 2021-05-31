package internal

import (
	. "caidaxiao/base"
	"time"

	"github.com/name5566/leaf/gate"
)

var (
	HBcheck    *time.Ticker // HeartBeat Check timmer
	heartWait  = 10         // 每十秒檢查一次
	heartIndex = 0          // 計時到十秒檢查
)

// 開始計時器檢查用戶心跳
func HeartBeatLoop() {
	Debug_log("HeartBeatLoop Start!")
	HBcheck = time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-HBcheck.C:
				intervalHeartBeat() //心跳  用戶
			}
		}
	}()
}

// Client<心跳包>傳送的結構體的對應方法
func HeartBeatHandler(args []interface{}) {

	// 消息的发送者
	a := args[1].(gate.Agent)

	userID, ok := userIDFromAgent_.Load(a)

	if !ok {
		return
	}

	client, ok := AgentFromuserID_.Load(userID.(int32))
	if ok {
		client.(*ClientInfo).expire = time.Now().Unix() + 10
	} else {
		Debug_log("[Error]Server can't send Pong to %v    agent:%v", userID, a)
	}
}

// 检测客户端连接是否超时
func intervalHeartBeat() {

	if heartIndex < heartWait { //每隔十秒檢查一次
		heartIndex++
		return
	}

	heartIndex = 0
	timestamp := time.Now().Unix()

	AgentFromuserID_.Range(func(k, client interface{}) bool {
		if client.(*ClientInfo).expire < timestamp {
			unusualLogout(client.(*ClientInfo).agent, "心跳超时")
		}
		return true
	})
}

// 關閉心跳
func CloseGameServer() {
	Debug_log("CloseGameServer")
	if HBcheck != nil { //關閉心跳
		HBcheck.Stop()
	}
}