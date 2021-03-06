package internal

// import (
// 	"bytes"
// 	"caidaxiao/conf"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"reflect"
// 	"runtime"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/name5566/leaf/log"
// 	"gopkg.in/mgo.v2/bson"
// )

// // 防止并发写Websocket用的锁
// var syncWrite sync.Mutex

// //CGTokenRsp 接受Token结构体
// type CGTokenRsp struct {
// 	Token string
// }

// //CGCenterRsp 中心返回消息结构体
// type CGCenterRsp struct {
// 	Status string
// 	Code   int
// 	Msg    *CGTokenRsp
// }

// //Conn4Center 连接到Center(中心服务器)的网络协议处理器
// type Conn4Center struct {
// 	GameId    string
// 	centerUrl string
// 	token     string
// 	DevKey    string
// 	conn      *websocket.Conn

// 	//除于登录成功状态
// 	LoginStat bool

// 	closebreathchan  chan bool
// 	closereceivechan chan bool

// 	//待处理的用户登录请求
// 	waitUser map[int32]*UserCallback
// }

// //Init 初始化
// func (c4c *Conn4Center) Init() {
// 	c4c.GameId = conf.Server.GameID
// 	c4c.DevKey = conf.Server.DevKey
// 	c4c.LoginStat = false
// 	c4c.closebreathchan = make(chan bool, 1)
// 	c4c.closereceivechan = make(chan bool, 1)
// 	c4c.waitUser = make(map[int32]*UserCallback)
// 	//go changeToken()
// }

// //onDestroy 销毁用户
// func (c4c *Conn4Center) onDestroy() {
// 	log.Debug("Conn4Center onDestroy ~")
// 	//c4c.UserLogoutCenter("991738698","123456") //测试用户 和 密码
// }

// //ReqCenterToken 向中心服务器请求token
// func (c4c *Conn4Center) ReqCenterToken() {
// 	// 拼接center Url
// 	url4Center := fmt.Sprintf("%s?dev_key=%s&dev_name=%s", conf.Server.TokenServer, c4c.DevKey, conf.Server.DevName)

// 	//log.Debug("<--- TokenServer Url --->: %v ", conf.Server.TokenServer)
// 	log.Debug("<--- Center access Url --->: %v ", url4Center)

// 	resp, err1 := http.Get(url4Center)
// 	if err1 != nil {
// 		panic(err1.Error())
// 	}
// 	log.Debug("<--- resp --->: %v ", resp)

// 	defer resp.Body.Close()

// 	if err1 == nil && resp.StatusCode == 200 {
// 		body, err2 := ioutil.ReadAll(resp.Body)
// 		if err2 != nil {
// 			panic(err2.Error())
// 		}
// 		//log.Debug("<----- resp.StatusCode ----->: %v", resp.StatusCode)
// 		log.Debug("<--- body --->: %v ,<--- err2 --->: %v", string(body), err2)

// 		var t CGCenterRsp
// 		err3 := json.Unmarshal(body, &t)
// 		log.Debug("<--- err3 --->: %v <--- Results --->: %v", err3, t)

// 		if t.Status == "SUCCESS" && t.Code == 200 {
// 			c4c.token = conf.Server.DevName
// 			c4c.CreatConnect()
// 		} else {
// 			log.Fatal("<--- Request Token Fail~ --->")
// 		}
// 	}
// }

// //CreatConnect 和Center建立链接
// func (c4c *Conn4Center) CreatConnect() {
// 	c4c.centerUrl = conf.Server.CenterUrl
// 	//c4c.centerUrl = "ws://172.16.1.41:9502/" //Pre
// 	//c4c.centerUrl = "ws://172.16.100.2:9502/" //上线
// 	//c4c.centerUrl = "ws" + strings.TrimPrefix(conf.Server.CenterServer, "http") //域名生成使用

// 	log.Debug("--- dial: --- : %v", c4c.centerUrl)
// 	conn, rsp, err := websocket.DefaultDialer.Dial(c4c.centerUrl, nil)
// 	c4c.conn = conn
// 	log.Debug("<--- Dial rsp --->: %v", rsp)

// 	if err != nil {
// 		log.Debug("CreatConnect:%v", err.Error())
// 		// log.Fatal("CreatConnect:%v", err.Error())
// 	} else {
// 		c4c.Run()
// 	}
// }

// func (c4c *Conn4Center) ReConnect() {
// 	go func() {
// 		for {
// 			if c4c.LoginStat == true {
// 				log.Debug("c4c.LoginStat == true~~~~~~~~~~~~~~~~~~~~~~~~~")
// 				return
// 			}
// 			c4c.closebreathchan <- true
// 			c4c.closereceivechan <- true
// 			c4c.CreatConnect()
// 			time.Sleep(time.Second * 5)
// 		}
// 	}()
// }

// //Run 开始运行,监听中心服务器的返回
// func (c4c *Conn4Center) Run() {

// 	go func() {
// 		ticker := time.NewTicker(time.Second * 5)
// 		for { //循环
// 			select {
// 			case <-ticker.C:
// 				c4c.onBreath()
// 				break
// 			case <-c4c.closebreathchan:
// 				ticker.Stop()
// 				return
// 			}
// 		}
// 	}()

// 	go func() {
// 		for {
// 			select {
// 			case <-c4c.closereceivechan:
// 				return
// 			default:
// 				typeId, message, err := c4c.conn.ReadMessage()
// 				if err != nil {
// 					log.Debug("Here is error by ReadMessage~ %v", err)
// 					log.Error(err.Error())
// 				}
// 				if typeId == -1 {
// 					log.Debug("中心服异常消息~")
// 					c4c.LoginStat = false
// 					c4c.ReConnect()
// 					return
// 				} else {
// 					c4c.onReceive(typeId, message)
// 				}
// 				break
// 			}
// 		}
// 	}()

// 	c4c.ServerLoginCenter()
// }

// //onBreath 中心服心跳
// func (c4c *Conn4Center) onBreath() {
// 	syncWrite.Lock()
// 	err := c4c.conn.WriteMessage(websocket.TextMessage, []byte(""))
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	syncWrite.Unlock()
// }

// //onReceive 接收消息
// func (c4c *Conn4Center) onReceive(messType int, messBody []byte) {
// 	if messType == websocket.TextMessage {
// 		baseData := &BaseMessage{}

// 		decoder := json.NewDecoder(strings.NewReader(string(messBody)))
// 		decoder.UseNumber()

// 		err := decoder.Decode(&baseData)
// 		if err != nil {
// 			log.Error(err.Error())
// 		}
// 		log.Debug("Receive From Center:%v", baseData)
// 		switch baseData.Event {
// 		case msgServerLogin:
// 			c4c.onServerLogin(baseData.Data)
// 			break
// 		case msgUserLogin:
// 			c4c.onUserLogin(baseData.Data)
// 			break
// 		case msgUserLogout:
// 			c4c.onUserLogout(baseData.Data)
// 			break
// 		case msgUserWinScore:
// 			c4c.onUserWinScore(baseData.Data)
// 			break
// 		case msgUserLoseScore:
// 			c4c.onUserLoseScore(baseData.Data)
// 			break
// 		case msgLockSettlement:
// 			c4c.onLockSettlement(baseData.Data)
// 			break
// 		case msgUnlockSettlement:
// 			c4c.onUnlockSettlement(baseData.Data)
// 			break
// 		case msgWinMoreThanNotice:
// 			c4c.onWinMoreThanNotice(baseData.Data)
// 			break
// 		case msgBankerStatus:
// 			c4c.onBankerStatus(baseData.Data)
// 			break
// 		case msgBankerWinScore:
// 			c4c.onBankerWinScore(baseData.Data)
// 			break
// 		case msgBankerLoseScore:
// 			c4c.onBankerLoseScore(baseData.Data)
// 			break
// 		default:
// 			log.Error("Receive a message but don't identify~")
// 		}
// 	}
// }

// //onServerLogin 服务器登录
// func (c4c *Conn4Center) onServerLogin(msgBody interface{}) {
// 	log.Debug("<-------- onServerLogin -------->: %v", msgBody)
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onServerLogin Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- serverLogin SUCCESS~!!! -------->")
// 		c4c.LoginStat = true

// 		msginfo := data["msg"].(map[string]interface{})
// 		fmt.Println("globals:", msginfo["globals"], reflect.TypeOf(msginfo["globals"]))

// 		globals := msginfo["globals"].([]interface{})
// 		fmt.Println("allList", globals)
// 		for k, v := range globals {
// 			fmt.Println(k, v)
// 			info := v.(map[string]interface{})
// 			//fmt.Println("package_id", info["package_id"])

// 			var nPackage uint16
// 			var nTax float64

// 			jsonPackageId, err := info["package_id"].(json.Number).Int64()
// 			if err != nil {
// 				log.Debug("jsonPackageId:%v", err.Error())
// 			} else {
// 				//fmt.Println("nPackage", uint16(jsonPackageId))
// 				nPackage = uint16(jsonPackageId)
// 			}
// 			jsonTax, err := info["platform_tax_percent"].(json.Number).Float64()

// 			if err != nil {
// 				log.Debug("jsonTax:%v", err.Error())
// 			} else {
// 				//fmt.Println("tax", uint8(jsonTax))
// 				nTax = jsonTax
// 			}

// 			SetPackageTaxM(nPackage, nTax)

// 			//log.Debug("packageId:%v,tax:%v", nPackage, nTax)
// 		}
// 	}

// }

// //onUserLogin 收到中心服的用户登录回应
// func (c4c *Conn4Center) onUserLogin(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onUserLogout Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if code != 200 {
// 		log.Debug("同步中心服登录失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- UserLogin SUCCESS~ -------->")

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		var strId int32
// 		var userData *UserCallback
// 		if ok {
// 			gameUser, uok := userInfo["game_user"].(map[string]interface{})
// 			if uok {
// 				nick := gameUser["game_nick"]
// 				headImg := gameUser["game_img"]
// 				userId := gameUser["id"]
// 				packageId := gameUser["package_id"]

// 				intID, err := userId.(json.Number).Int64()
// 				if err != nil {
// 					log.Fatal("onUserLogin intID:%v", err.Error())
// 				}
// 				strId = int32(intID)

// 				pckId, err2 := packageId.(json.Number).Int64()
// 				if err2 != nil {
// 					log.Fatal("onUserLogin pckId:%v", err2.Error())
// 				}

// 				//找到等待登录玩家
// 				userData, ok = c4c.waitUser[strId]
// 				if ok {
// 					userData.Data.HeadImg = headImg.(string)
// 					userData.Data.NickName = nick.(string)
// 					userData.Data.PackageId = uint16(pckId)
// 				}
// 			}
// 			gameAccount, okA := userInfo["game_account"].(map[string]interface{})

// 			if okA {
// 				balance := gameAccount["balance"]
// 				floatBalance, err := balance.(json.Number).Float64()
// 				if err != nil {
// 					log.Error(err.Error())
// 				}

// 				userData.Data.Account = floatBalance

// 				//调用玩家绑定回调函数
// 				if userData.Callback != nil {
// 					userData.Callback(&userData.Data)
// 				}
// 			}
// 		}
// 	}
// }

// func (c4c *Conn4Center) onUserLogout(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onUserLogout Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if code != 200 {
// 		log.Debug("同步中心服登出失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- UserLogout SUCCESS~ -------->")

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		var strId int32
// 		var userData *UserCallback
// 		if ok {
// 			gameUser, uok := userInfo["game_user"].(map[string]interface{})
// 			if uok {
// 				nick := gameUser["game_nick"]
// 				headImg := gameUser["game_img"]
// 				userId := gameUser["id"]

// 				intID, err := userId.(json.Number).Int64()
// 				if err != nil {
// 					log.Fatal("onUserLogout:%v", err.Error())
// 				}
// 				strId = int32(intID)
// 				//找到等待登录玩家
// 				userData, ok = c4c.waitUser[strId]
// 				if ok {
// 					userData.Data.HeadImg = headImg.(string)
// 					userData.Data.NickName = nick.(string)
// 				}
// 			}
// 		}
// 	}
// }

// func (c4c *Conn4Center) onUserWinScore(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onUserWinScore Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if code != 200 {
// 		log.Debug("同步中心服赢钱失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- UserWinScore SUCCESS~ -------->")

// 		//将Win数据插入数据
// 		InsertWinMoney(msgBody) //todo

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		if ok {
// 			jsonScore := userInfo["final_pay"]
// 			score, err := jsonScore.(json.Number).Float64()

// 			log.Debug("同步中心服赢钱成功:%v", score)

// 			if err != nil {
// 				log.Error(err.Error())
// 				return
// 			}
// 		}
// 	}
// }

// func (c4c *Conn4Center) onUserLoseScore(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onUserLoseScore Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	if code != 200 {
// 		log.Error("同步中心服输钱失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- UserLoseScore SUCCESS~ -------->")

// 		//将Lose数据插入数据
// 		InsertLoseMoney(msgBody) //todo

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		if ok {
// 			jsonScore := userInfo["final_pay"]
// 			score, err := jsonScore.(json.Number).Float64()

// 			log.Debug("同步中心服输钱成功:%v", score)

// 			if err != nil {
// 				log.Error(err.Error())
// 				return
// 			}
// 		}
// 	}
// }

// //onWinMoreThanNotice 加锁金额
// func (c4c *Conn4Center) onLockSettlement(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if ok {
// 		code, err := data["code"].(json.Number).Int64()
// 		if err != nil {
// 			log.Fatal("onLockSettlement:%v", err.Error())
// 		}

// 		fmt.Println(code, reflect.TypeOf(code))
// 		if data["status"] == "SUCCESS" && code == 200 {
// 			log.Debug("<-------- onLockSettlement SUCCESS~!!! -------->")
// 		}
// 	}
// }

// //onWinMoreThanNotice 解锁金额
// func (c4c *Conn4Center) onUnlockSettlement(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if ok {
// 		code, err := data["code"].(json.Number).Int64()
// 		if err != nil {
// 			log.Fatal("onUnlockSettlement:%v", err.Error())
// 		}

// 		fmt.Println(code, reflect.TypeOf(code))
// 		if data["status"] == "SUCCESS" && code == 200 {
// 			log.Debug("<-------- onUnlockSettlement SUCCESS~!!! -------->")
// 		}
// 	}
// }

// func (c4c *Conn4Center) onBankerStatus(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onBankerStatus Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if code != 200 {
// 		log.Debug("同步中心服庄家状态失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- onBankerStatus SUCCESS~ -------->")
// 	}
// }

// func (c4c *Conn4Center) onBankerWinScore(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onBankerWinScore Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}

// 	if code != 200 {
// 		log.Debug("同步中心服庄家赢钱失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- onBankerWinScore SUCCESS~ -------->")

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		if ok {
// 			jsonScore := userInfo["final_pay"]
// 			score, err := jsonScore.(json.Number).Float64()

// 			log.Debug("同步中心服庄家赢钱成功:%v", score)

// 			if err != nil {
// 				log.Error(err.Error())
// 				return
// 			}
// 		}
// 	}
// }

// func (c4c *Conn4Center) onBankerLoseScore(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if !ok {
// 		log.Debug("onBankerLoseScore Error")
// 	}

// 	code, err := data["code"].(json.Number).Int64()
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	if code != 200 {
// 		log.Error("同步中心服庄家输钱失败:%v", data)
// 		return
// 	}

// 	if data["status"] == "SUCCESS" && code == 200 {
// 		log.Debug("<-------- onBankerLoseScore SUCCESS~ -------->")

// 		userInfo, ok := data["msg"].(map[string]interface{})
// 		if ok {
// 			jsonScore := userInfo["final_pay"]
// 			score, err := jsonScore.(json.Number).Float64()

// 			log.Debug("同步中心服庄家输钱成功:%v", score)

// 			if err != nil {
// 				log.Error(err.Error())
// 				return
// 			}
// 		}
// 	}
// }

// //onWinMoreThanNotice 服务器登录
// func (c4c *Conn4Center) onWinMoreThanNotice(msgBody interface{}) {
// 	data, ok := msgBody.(map[string]interface{})
// 	if ok {
// 		code, err := data["code"].(json.Number).Int64()
// 		if err != nil {
// 			log.Fatal("onWinMoreThanNotice:%v", err.Error())
// 		}

// 		fmt.Println(code, reflect.TypeOf(code))
// 		if data["status"] == "SUCCESS" && code == 200 {
// 			log.Debug("<-------- onWinMoreThanNotice SUCCESS~!!! -------->")
// 		}
// 	}
// }

// //ServerLoginCenter 服务器登录Center
// func (c4c *Conn4Center) ServerLoginCenter() {
// 	baseData := &BaseMessage{}
// 	baseData.Event = msgServerLogin
// 	baseData.Data = ServerLogin{
// 		Host:    conf.Server.CenterServer,
// 		Port:    conf.Server.CenterServerPort,
// 		GameId:  c4c.GameId,
// 		DevName: conf.Server.DevName,
// 		DevKey:  c4c.DevKey,
// 	}
// 	// 发送消息到中心服
// 	c4c.SendMsg2Center(baseData)
// }

// //UserLoginCenter 用户登录
// func (c4c *Conn4Center) UserLoginCenter(userId int32, password string, token string, callback func(data *Player)) {
// 	if !c4c.LoginStat {
// 		log.Debug("<-------- caidaxiao not ready~!!! -------->")
// 		return
// 	}
// 	baseData := &BaseMessage{}
// 	baseData.Event = msgUserLogin
// 	if password != "" {
// 		baseData.Data = &UserReq{
// 			ID:       userId,
// 			PassWord: password,
// 			GameId:   c4c.GameId,
// 			DevName:  conf.Server.DevName,
// 			DevKey:   c4c.DevKey}
// 	} else {
// 		baseData.Data = &UserReq{
// 			ID:      userId,
// 			Token:   token,
// 			GameId:  c4c.GameId,
// 			DevName: conf.Server.DevName,
// 			DevKey:  c4c.DevKey}
// 	}

// 	c4c.SendMsg2Center(baseData)

// 	//加入待处理map，等待处理
// 	c4c.waitUser[userId] = &UserCallback{}
// 	c4c.waitUser[userId].Data.Id = userId
// 	c4c.waitUser[userId].Callback = callback
// }

// //UserLogoutCenter 用户登出
// func (c4c *Conn4Center) UserLogoutCenter(userId int32, password string, token string) {
// 	base := &BaseMessage{}
// 	base.Event = msgUserLogout
// 	if password != "" {
// 		base.Data = &UserReq{
// 			ID:       userId,
// 			PassWord: password,
// 			GameId:   c4c.GameId,
// 			DevName:  conf.Server.DevName,
// 			DevKey:   c4c.DevKey}
// 	} else {
// 		base.Data = &UserReq{
// 			ID:      userId,
// 			Token:   token,
// 			GameId:  c4c.GameId,
// 			DevName: conf.Server.DevName,
// 			DevKey:  c4c.DevKey}
// 	}

// 	// 发送消息到中心服
// 	c4c.SendMsg2Center(base)
// }

// //SendMsg2Center 发送消息到中心服
// func (c4c *Conn4Center) SendMsg2Center(data interface{}) {
// 	syncWrite.Lock()
// 	defer syncWrite.Unlock()
// 	// Json序列化
// 	codeData, err1 := json.Marshal(data)
// 	if err1 != nil {
// 		log.Error(err1.Error())
// 	}
// 	log.Debug("Msg to Send Center:%v", string(codeData))

// 	err2 := c4c.conn.WriteMessage(websocket.TextMessage, []byte(codeData))
// 	if err2 != nil {
// 		log.Fatal("SendMsg2Center:%v", err2.Error())
// 	}
// }

// //UserSyncWinScore 同步赢分
// func (c4c *Conn4Center) UserSyncWinScore(p *Player, timeUnix int64, roundId, reason string, betMoney float64) {
// 	baseData := &BaseMessage{}
// 	baseData.Event = msgUserWinScore
// 	userWin := &UserChangeScore{}
// 	userWin.Auth.DevName = conf.Server.DevName
// 	userWin.Auth.DevKey = c4c.DevKey
// 	userWin.Info.CreateTime = timeUnix
// 	userWin.Info.GameId = c4c.GameId
// 	userWin.Info.ID = int(p.Id)
// 	userWin.Info.LockMoney = 0
// 	userWin.Info.Money = p.WinResultMoney
// 	userWin.Info.BetMoney = betMoney
// 	userWin.Info.Order = bson.NewObjectId().Hex()

// 	userWin.Info.PayReason = reason
// 	userWin.Info.PreMoney = 0
// 	userWin.Info.RoundId = roundId
// 	baseData.Data = userWin
// 	c4c.SendMsg2Center(baseData)
// }

// //UserSyncWinScore 同步输分
// func (c4c *Conn4Center) UserSyncLoseScore(p *Player, timeUnix int64, roundId, reason string, betMoney float64) {
// 	baseData := &BaseMessage{}
// 	baseData.Event = msgUserLoseScore
// 	userLose := &UserChangeScore{}
// 	userLose.Auth.DevName = conf.Server.DevName
// 	userLose.Auth.DevKey = c4c.DevKey
// 	userLose.Info.CreateTime = timeUnix
// 	userLose.Info.GameId = c4c.GameId
// 	userLose.Info.ID = int(p.Id)
// 	userLose.Info.LockMoney = 0
// 	userLose.Info.Money = p.LoseResultMoney
// 	userLose.Info.BetMoney = betMoney
// 	userLose.Info.Order = bson.NewObjectId().Hex()
// 	userLose.Info.PayReason = reason
// 	userLose.Info.PreMoney = 0
// 	userLose.Info.RoundId = roundId
// 	baseData.Data = userLose
// 	c4c.SendMsg2Center(baseData)
// }

// //锁钱
// func (c4c *Conn4Center) LockSettlement(p *Player) {
// 	timeStr := time.Now().Format("2006-01-02_15:04:05")
// 	loseOrder := string(p.Id) + "_" + timeStr + "_LockMoney"

// 	baseData := &BaseMessage{}
// 	baseData.Event = msgLockSettlement
// 	lockMoney := &UserChangeScore{}
// 	lockMoney.Auth.DevName = conf.Server.DevName
// 	lockMoney.Auth.DevKey = c4c.DevKey
// 	lockMoney.Info.CreateTime = time.Now().Unix()
// 	lockMoney.Info.GameId = c4c.GameId
// 	lockMoney.Info.ID = int(p.Id)
// 	lockMoney.Info.LockMoney = 0
// 	lockMoney.Info.Money = 0
// 	lockMoney.Info.Order = loseOrder
// 	lockMoney.Info.PayReason = "lockMoney"
// 	lockMoney.Info.PreMoney = 0
// 	lockMoney.Info.RoundId = p.RoundId
// 	baseData.Data = lockMoney
// 	c4c.SendMsg2Center(baseData)
// }

// //解锁
// func (c4c *Conn4Center) UnlockSettlement(p *Player) {
// 	timeStr := time.Now().Format("2006-01-02_15:04:05")
// 	loseOrder := string(p.Id) + "_" + timeStr + "_UnlockMoney"

// 	baseData := &BaseMessage{}
// 	baseData.Event = msgUnlockSettlement
// 	lockMoney := &UserChangeScore{}
// 	lockMoney.Auth.DevName = conf.Server.DevName
// 	lockMoney.Auth.DevKey = c4c.DevKey
// 	lockMoney.Info.CreateTime = time.Now().Unix()
// 	lockMoney.Info.GameId = c4c.GameId
// 	lockMoney.Info.ID = int(p.Id)
// 	lockMoney.Info.LockMoney = p.Account
// 	lockMoney.Info.Money = 0
// 	lockMoney.Info.Order = loseOrder
// 	lockMoney.Info.PayReason = "UnlockMoney"
// 	lockMoney.Info.PreMoney = 0
// 	lockMoney.Info.RoundId = p.RoundId
// 	baseData.Data = lockMoney
// 	c4c.SendMsg2Center(baseData)
// }

// func (c4c *Conn4Center) NoticeWinMoreThan(playerId, playerName string, winGold float64) {
// 	log.Debug("<-------- NoticeWinMoreThan  -------->")
// 	msg := fmt.Sprintf("<size=20><color=yellow>恭喜!</color><color=orange>%v</color><color=yellow>在</color></><color=orange><size=25>德州扑克</color></><color=yellow><size=20>中一把赢了</color></><color=yellow><size=30>%.2f</color></><color=yellow><size=25>金币！</color></>", playerName, winGold)

// 	base := &BaseMessage{}
// 	base.Event = msgWinMoreThanNotice
// 	id, _ := strconv.Atoi(playerId)
// 	base.Data = &Notice{
// 		DevName: conf.Server.DevName,
// 		DevKey:  conf.Server.DevKey,
// 		ID:      id,
// 		GameId:  c4c.GameId,
// 		Type:    2000,
// 		Message: msg,
// 		Topic:   "系统提示",
// 	}
// 	c4c.SendMsg2Center(base)
// }

// //BankerStatus 庄家同步状态
// // func (c4c *Conn4Center) BankerStatus(p *Player, status int, timeUnix int64, roundId, reason string) {
// // 	baseData := &BaseMessage{}
// // 	baseData.Event = msgBankerStatus
// // 	id, _ := strconv.Atoi(p.Id)
// // 	banker := &BankerReqInfo{}
// // 	banker.Auth.DevName = conf.Server.DevName
// // 	banker.Auth.DevKey = c4c.DevKey

// // 	banker.Info.Id = id
// // 	banker.Info.Status = status
// // 	banker.Info.CreateTime = timeUnix
// // 	banker.Info.PayReason = reason
// // 	banker.Info.Order = bson.NewObjectId().Hex()
// // 	banker.Info.GameId = c4c.GameId
// // 	banker.Info.RoundId = roundId
// // 	banker.Info.LockMoney = 0
// // 	banker.Info.Money = p.Account

// // 	baseData.Data = banker
// // 	c4c.SendMsg2Center(baseData)
// // }

// //BankerWinScore 庄家同步赢分
// // func (c4c *Conn4Center) BankerWinScore(p *Player, timeUnix int64, roundId, reason string) {
// // 	baseData := &BaseMessage{}
// // 	baseData.Event = msgBankerWinScore
// // 	id, _ := strconv.Atoi(p.Id)
// // 	banker := &UserChangeScore{}
// // 	banker.Auth.DevName = conf.Server.DevName
// // 	banker.Auth.DevKey = c4c.DevKey

// // 	banker.Info.ID = id
// // 	banker.Info.CreateTime = timeUnix
// // 	banker.Info.PayReason = reason
// // 	banker.Info.GameId = c4c.GameId
// // 	banker.Info.RoundId = roundId
// // 	banker.Info.PreMoney = 0
// // 	banker.Info.LockMoney = 0
// // 	banker.Info.Money = p.WinResultMoney
// // 	banker.Info.Order = bson.NewObjectId().Hex()

// // 	baseData.Data = banker
// // 	c4c.SendMsg2Center(baseData)
// // }

// //BankerLoseScore 庄家同步输分
// // func (c4c *Conn4Center) BankerLoseScore(p *Player, timeUnix int64, roundId, reason string) {
// // 	baseData := &BaseMessage{}
// // 	baseData.Event = msgBankerLoseScore
// // 	id, _ := strconv.Atoi(p.Id)
// // 	banker := &UserChangeScore{}
// // 	banker.Auth.DevName = conf.Server.DevName
// // 	banker.Auth.DevKey = c4c.DevKey

// // 	banker.Info.ID = id
// // 	banker.Info.CreateTime = timeUnix
// // 	banker.Info.PayReason = reason
// // 	banker.Info.GameId = c4c.GameId
// // 	banker.Info.RoundId = roundId
// // 	banker.Info.PreMoney = 0
// // 	banker.Info.LockMoney = 0
// // 	banker.Info.Money = p.LoseResultMoney
// // 	banker.Info.Order = bson.NewObjectId().Hex()

// // 	baseData.Data = banker
// // 	c4c.SendMsg2Center(baseData)
// // }

// //Init 初始化
// func (cc *mylog) Init() {

// }
// func (cc *mylog) log(v ...interface{}) {
// 	senddata := logmsg{
// 		Type:     "LOG",
// 		From:     "RedBlack-War",
// 		GameName: "红黑大战",
// 		Id:       conf.Server.GameID,
// 		Host:     "",
// 		Time:     time.Now().Unix(),
// 	}

// 	_, file, line, ok := runtime.Caller(2)
// 	if ok {
// 		senddata.File = file
// 		senddata.Line = line
// 	}
// 	Msg := fmt.Sprintln(v...)
// 	senddata.Msg = Msg
// 	cc.sendMsg(senddata)
// }

// func (cc *mylog) debug(v ...interface{}) {
// 	senddata := logmsg{
// 		Type:     "DEG",
// 		From:     "RedBlack-War",
// 		GameName: "红黑大战",
// 		Id:       conf.Server.GameID,
// 		Host:     "",
// 		Time:     time.Now().Unix(),
// 	}

// 	_, file, line, ok := runtime.Caller(2)
// 	if ok {
// 		senddata.File = file
// 		senddata.Line = line
// 	}
// 	Msg := fmt.Sprintln(v...)
// 	senddata.Msg = Msg
// 	cc.sendMsg(senddata)
// }

// func (cc *mylog) error(v ...interface{}) {
// 	senddata := logmsg{
// 		Type:     "ERR",
// 		From:     "RedBlack-War",
// 		GameName: "RedBlack-War",
// 		Id:       conf.Server.GameID,
// 		Host:     "",
// 		Time:     time.Now().Unix(),
// 	}

// 	_, file, line, ok := runtime.Caller(2)
// 	if ok {
// 		senddata.File = file
// 		senddata.Line = line
// 	}
// 	Msg := fmt.Sprintln(v...)
// 	senddata.Msg = Msg
// 	cc.sendMsg(senddata)
// }

// func (cc *mylog) sendMsg(senddata logmsg) {
// 	bodyJson, err1 := json.Marshal(senddata)
// 	if err1 != nil {
// 		log.Error(err1.Error())
// 	}
// 	req, err2 := http.NewRequest(http.MethodPost, conf.Server.LogAddr, bytes.NewBuffer(bodyJson))
// 	if err2 != nil {
// 		log.Error(err1.Error())
// 	}
// 	if req != nil {
// 		req.Header.Add("content-type", "application/json")
// 		err3 := req.Body.Close()
// 		if err3 != nil {
// 			log.Error(err1.Error())
// 		}
// 	}
// }
