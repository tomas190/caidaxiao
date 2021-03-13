package internal

import (
	"caidaxiao/conf"
	"caidaxiao/msg"
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"net/http"
	"time"
)

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	SuccCode = 0
	ErrCode  = -1
)

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	//http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	//http.HandleFunc("/api/getGameData", getAccessData)
	// 查询子游戏盈余池数据
	//http.HandleFunc("/api/getSurplusOne", getSurplusOne)
	// 修改盈余池数据
	//http.HandleFunc("/api/uptSurplusConf", uptSurplusOne)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)
	// 获取玩家信息
	//http.HandleFunc("/api/getPlayInfo", getPlayInfo)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func reqPlayerLeave(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("id")
	rid := hall.UserRoom[Id]
	v, _ := hall.RoomRecord.Load(rid)
	if v != nil {
		room := v.(*Room)
		user, _ := hall.UserRecord.Load(Id)
		if user != nil {
			p := user.(*Player)
			//if p.IsBanker == true {
			room.IsConBanker = false
			nowTime := time.Now().Unix()
			p.RoundId = fmt.Sprintf("%+v-%+v", time.Now().Unix(), room.RoomId)
			reason := "庄家申请下庄"
			c4c.BankerStatus(p, 0, nowTime, p.RoundId, reason)
			//}
			hall.UserRecord.Delete(p.Id)
			p.PlayerExitRoom()
			c4c.UserLogoutCenter(p.Id, p.Password, p.Token)
			leaveHall := &msg.Logout_S2C{}
			p.SendMsg(leaveHall)

			js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
			if err != nil {
				fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
				return
			}
			w.Write(js)
		}
	}
	//user, _ := hall.UserRecord.Load(Id)
	//if user != nil {
	//	u := user.(*Player)
	//	u.IsAction = false
	//	u.TotalDownBet = 0
	//	if u.IsBanker == true {
	//
	//	}
	//	u.PlayerExitRoom()
	//	js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
	//	if err != nil {
	//		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
	//		return
	//	}
	//	w.Write(js)
	//}
}

func NewResp(code int, msg string, data interface{}) ApiResp {
	return ApiResp{Code: code, Msg: msg, Data: data}
}
