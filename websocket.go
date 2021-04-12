package main

import (
	"bytes"
	"caidaxiao/msg"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// 定義flag參數，這邊會返回一個相應的指針
var addr = flag.String("addr", "localhost:1355", "http service address")

func main() {

	t := time.NewTicker(50 * time.Millisecond)
	num := 0
	for {
		<-t.C
		if num < 10000 { //建立goroutine數量
			num++
			fmt.Printf("num:%v\n", num)

			go func() {

				/*!!!!!!!!!!!!!!!!!!!!!!!!建立連線 開始!!!!!!!!!!!!!!!!!!!!!!!!*/
				var userID string

				buf := make([]byte, 100)

				// 定義一個os.Signal的通道
				interrupt := make(chan os.Signal, 1)

				// Notify函數讓signal包將輸入信號轉到interrupt
				signal.Notify(interrupt, os.Interrupt)

				// 连接网址
				//u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
				// logger.Debug("connecting to %s", u.String())

				// 连接服务器
				//ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				// 連接服務器 PRE
				ws, _, err := websocket.DefaultDialer.Dial("ws://game.tampk.club/caidaxiao", nil)
				if err != nil {
					log.Fatal("dial:", err)
				}

				// 預先關閉，此行在離開main時會執行
				defer ws.Close()

				// 定義通道
				done := make(chan struct{})
				/*!!!!!!!!!!!!!!!!!!!!!!!!建立連線 結束!!!!!!!!!!!!!!!!!!!!!!!!*/

				/*!!!!!!!!!!!!!!!!!!!!!!!!接收資料 開始!!!!!!!!!!!!!!!!!!!!!!!!*/
				go func() { //接收資料
					// 預先關閉，此行在離開本協程時執行
					defer close(done)
					for {
						// 一直待命讀資料
						_, message, err := ws.ReadMessage()
						if err != nil {
							log.Println("read:", err)
							return
						}
						var pkgID = binary.BigEndian.Uint16(message[0:2])
						//var bodyId = binary.BigEndian.Uint16(string(myMsg.Command_Pong))
						if int16(pkgID) == int16(msg.MessageID_MSG_Pong) { // 心跳回傳
							// logger.Debug("MessageKind_Pong")
							// 將已經編碼的資料解碼成 protobuf.User 格式。
							// var bodyClass myMsg.StoCHeartBeat
							// proto.Unmarshal(message[2:], &bodyClass)
							// logger.Debug("心跳回傳recv: %v %v %v %v", pkgID, int16(myMsg.MessageKind_Pong), message, &bodyClass)
						} else if int16(pkgID) == int16(msg.MessageID_MSG_Login_S2C) {
							bodyClass := &msg.Login_S2C{}
							proto.Unmarshal(message[2:], bodyClass)
							userID = bodyClass.PlayerInfo.Id
							// logger.Debug("登陸回傳recv: %v %v %v %v", pkgID, int16(myMsg.MessageKind_LoginR), message, userID)
						} else {
							// logger.Debug("recv: %v ", message)
						}
					}
				}()

				/*!!!!!!!!!!!!!!!!!!!!!!!!接收資料 結束!!!!!!!!!!!!!!!!!!!!!!!!*/

				/*!!!!!!!!!!!!!!!!!!!!!!!!發送資料 開始!!!!!!!!!!!!!!!!!!!!!!!!*/
				// 心跳
				heartbeat := time.NewTicker(5 * time.Second)

				// spin
				// delayTime := rand.Intn(3) + 2
				delayTime := 10
				replay := time.NewTicker(time.Duration(delayTime) * time.Second)

				// 預先停止，此行在離開main時執行
				defer func() {
					heartbeat.Stop()
					replay.Stop()
				}()

				FirstLogin(ws)
				time.Sleep(1) //確保玩家登入再傳心跳
				JoinRoom(ws)
				//計時器定時發任務
				for {
					select {
					case <-replay.C:
						// 重覆玩
						var pkgID uint16
						pkgID = uint16(msg.MessageID_MSG_PlayerAction_C2S)
						// userIDIndex := rand.Intn(len(allUserID))
						binary.BigEndian.PutUint16(buf[0:2], pkgID)
						num := RandInRange(1, 7+1)
						mssage := &msg.PlayerAction_C2S{
							DownBet:  1,
							DownPot:  msg.PotType(num),
							IsAction: true,
							Id:       userID,
						}

						// 將資料編碼成 Protocol Buffer 格式（請注意是傳入 Pointer）。
						dataBuffer, _ := proto.Marshal(mssage)

						// 將消息ID與DATA整合，一起送出
						pkgData := [][]byte{buf[:2], dataBuffer}
						pkgDatas := bytes.Join(pkgData, []byte{})
						err = ws.WriteMessage(websocket.BinaryMessage, pkgDatas)

						if err != nil {
							log.Println("write:", err)
							return
						}

					case <-heartbeat.C:

						var pkgID uint16
						pkgID = uint16(msg.MessageID_MSG_Ping)
						binary.BigEndian.PutUint16(buf[0:2], pkgID)
						data := msg.Ping{}

						// 將資料編碼成 Protocol Buffer 格式（請注意是傳入 Pointer）。
						dataBuffer, _ := proto.Marshal(&data)

						// 將消息ID與DATA整合，一起送出
						pkgData := [][]byte{buf[:2], dataBuffer}
						pkgDatas := bytes.Join(pkgData, []byte{})
						err = ws.WriteMessage(websocket.BinaryMessage, pkgDatas)

						if err != nil {
							log.Println("write:", err)
							return
						}
						// logger.Debug("write:", pkgDatas)
					case <-interrupt:
						// 強制執行程序時，會進入這邊
						log.Println("interrupt")

						err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							log.Println("write close::", err)
							return
						}
						select {
						case <-done:
							// 結束完成會執行這邊
							log.Println("<-done")
						case <-time.After(10 * time.Second):
							// 超時處理，防止select阻塞著
							log.Println("<-time")
						}
						return
					}
				}
				/*!!!!!!!!!!!!!!!!!!!!!!!!發送資料 結束!!!!!!!!!!!!!!!!!!!!!!!!*/
			}()
		}

	}

}

func FirstLogin(ws *websocket.Conn) {
	var pkgID uint16
	pkgID = uint16(msg.MessageID_MSG_Login_C2S)
	buf := make([]byte, 100)
	binary.BigEndian.PutUint16(buf[0:2], pkgID)
	playerId := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(80000000))
	data := msg.Login_C2S{
		Id:       playerId,
		PassWord: "123456",
	}

	// 將資料編碼成 Protocol Buffer 格式（請注意是傳入 Pointer）。
	dataBuffer, _ := proto.Marshal(&data)

	// 將消息ID與DATA整合，一起送出
	pkgData := [][]byte{buf[:2], dataBuffer}
	pkgDatas := bytes.Join(pkgData, []byte{})
	err := ws.WriteMessage(websocket.BinaryMessage, pkgDatas)

	if err != nil {
		log.Println("write:", err)
		return
	}
}

func JoinRoom(ws *websocket.Conn) {
	var pkgID uint16
	pkgID = uint16(msg.MessageID_MSG_JoinRoom_C2S)
	buf := make([]byte, 100)
	binary.BigEndian.PutUint16(buf[0:2], pkgID)
	data := msg.JoinRoom_C2S{
		RoomId: "1",
	}

	// 將資料編碼成 Protocol Buffer 格式（請注意是傳入 Pointer）。
	dataBuffer, _ := proto.Marshal(&data)

	// 將消息ID與DATA整合，一起送出
	pkgData := [][]byte{buf[:2], dataBuffer}
	pkgDatas := bytes.Join(pkgData, []byte{})
	err := ws.WriteMessage(websocket.BinaryMessage, pkgDatas)

	if err != nil {
		log.Println("write:", err)
		return
	}
}

func RandInRange(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}
