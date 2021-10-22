package internal

import (
	common "caidaxiao/base"
	"encoding/json"
	"fmt"
	"math"
)

// 傳送至中心服的登入登出請求
func makeAuthRequest(userID int32) AuthUserPsw {
	return AuthUserPsw{
		UserID: userID,
		GameID: c4c.gameId,
		// Token:  c4c.token,
		DevKey:  c4c.devKey,
		DevName: c4c.devName,
	}
}

// 中心服回傳登入登出資料解析
func makeAuthResponse(data ResponseData) *UserResponse {
	jsonStr, err1 := json.Marshal(data.Msg)
	if err1 != nil {
		fmt.Printf("解析用户数据失败,err=%+v", err1)
		return nil
	}

	info := &UserResponse{}
	err2 := json.Unmarshal([]byte(jsonStr), info)
	if err2 != nil {
		fmt.Printf("解析用户数据失败,err=%+v", err2)
		return nil
	}
	return info
}

// 傳送至中心服的支付要求
func makePayRequest(data common.AmountFlowReq) *PayRequest {
	record := &PayRequest{
		Auth: AuthData{
			//Token:  c4c.token,
			DevKey:  c4c.devKey,
			DevName: c4c.devName,
		},
		Info: PayData{
			GameID:     c4c.gameId,
			CreateTime: data.TimeStamp,
			UserID:     data.UserID,
			RoundID:    data.RoundID,
			OrderID:    data.Order,
			PayReason:  data.Reason,
		},
	}
	return record
}

// 中心服回傳支付響應
func makePayResponse(v interface{}) *PayResponse {
	jsonStr, err1 := json.Marshal(v)
	if err1 != nil {
		common.Debug_log("解析资金周转返回数据失败,err=%+v", err1)
		return nil
	}

	info := &PayResponse{}
	err2 := json.Unmarshal([]byte(jsonStr), info)
	if err2 != nil {
		common.Debug_log("Json序列化资金周转返回数据失败,err=%+v", err2)
		return nil
	}
	return info
}

////////////////////////////////////////////////////////////////////////////////
// client -> centerserver
// 用戶相關邏輯對應的處理(ChanRPC)
////////////////////////////////////////////////////////////////////////////////

// 用戶登入 client -> centerserver
func C2CS_Login(args []interface{}) {
	// common.Debug_log("loginModule C2CS_Login")
	userID := args[0].(int32)
	pwd := args[1].(string)
	token := args[2].(string)
	common.Debug_log("C2CS_Login UserID:%v(%T) UserPW:%v(%T) UserToken:%v(%T)", userID, userID, pwd, pwd, token, token)
	if len(pwd) == 6 {
		auth := AuthUserPsw{
			UserID:  userID,
			GameID:  c4c.gameId,
			DevKey:  c4c.devKey,
			DevName: c4c.devName,
		}
		auth.PassWord = pwd
		c4c.sendMessage(MSG_USER_LOGIN, auth)
	} else {
		auth := AuthUserToken{
			UserID:  userID,
			GameID:  c4c.gameId,
			DevKey:  c4c.devKey,
			DevName: c4c.devName,
		}
		auth.Token = token
		c4c.sendMessage(MSG_USER_LOGIN, auth)
	}
}

// 用戶登出 client -> centerserver
func C2CS_Logout(args []interface{}) {
	// common.Debug_log("loginModule C2CS_Logout")
	userID := args[0].(int32)
	req := makeAuthRequest(userID)
	c4c.sendMessage(MSG_USER_LOGOUT, req)
}

// 用戶贏錢 client -> centerserver
func C2CS_WinSettlement(args []interface{}) {
	// common.Debug_log("loginModule C2CS_WinSettlement")
	data := args[0].(common.AmountFlowReq)
	req := makePayRequest(data)
	req.Info.BetMoney = data.BetMoney
	req.Info.Money = data.Money
	req.Info.LockMoney = data.LockMoney
	c4c.sendMessage(MSG_USER_WIN_MONEY, req)
}

// 用戶輸錢 client -> centerserver
func C2CS_LoseSettlement(args []interface{}) {
	// common.Debug_log("loginModule C2CS_LoseSettlement")
	data := args[0].(common.AmountFlowReq)
	req := makePayRequest(data)
	req.Info.BetMoney = data.BetMoney
	req.Info.Money = data.Money
	req.Info.LockMoney = data.LockMoney
	c4c.sendMessage(MSG_USER_LOSE_MONEY, req)
}

// 用戶資金鎖定 client -> centerserver
func C2CS_LockSettlement(args []interface{}) {
	// common.Debug_log("loginModule C2CS_LockSettlement")
	data := args[0].(common.AmountFlowReq)
	req := makePayRequest(data)
	req.Info.LockMoney = data.Money
	c4c.sendMessage(MSG_USER_LOCK_MONEY, req)
}

// 用戶資金解鎖 client -> centerserver
func C2CS_UnLockSettlement(args []interface{}) {
	// common.Debug_log("loginModule C2CS_UnLockSettlement")
	data := args[0].(common.AmountFlowReq)
	req := makePayRequest(data)
	req.Info.LockMoney = data.Money
	c4c.sendMessage(MSG_USER_UNLOCK_MONEY, req)

}

// 用戶向中心服推送消息 client -> centerserver
func C2CS_NoticeBroadcast(args []interface{}) {
	// common.Debug_log("loginModule C2CS_NoticeBroadcast")
	data := args[0].(common.AmountFlowReq)
	msg := fmt.Sprintf("<size=20><color=yellow>恭喜!</color><color=orange>%s</color><color=yellow>在</color></><color=orange><size=25>彩源猜大小</color></><color=yellow><size=20>中一把赢了</color></><color=yellow><size=30>%.2f</color></><color=yellow><size=25>金币！</color></>", data.UserName, data.Money)
	record := &NoticeRequest{
		UserID:  data.UserID,
		GameID:  c4c.gameId,
		DevKey:  c4c.devKey,
		DevName: c4c.devName,
		Type:    2000,
		Topic:   "系统提示",
		Message: msg,
	}
	c4c.sendMessage(MSG_NOTICE, record)
}

//////////////////////////////////////////////////////////////////////////
// centerserver -> client
// 接收中心服回傳後的處理
//////////////////////////////////////////////////////////////////////////

// 用戶登入 client -> centerserver
func CS2C_Login(data ResponseData) {
	// common.Debug_log("loginModule CS2C_Login")
	if data.Code != 200 {
		common.Debug_log("Login Error Code : ", data.Code)
		return
	}
	info := makeAuthResponse(data)
	if info == nil {
		return
	}
	common.GetInstance().Game.Go("UserLogin", common.UserInfo{ //進入遊戲服務
		UserID:      info.Base.ID,
		UserName:    info.Base.GameNick,
		UserHead:    info.Base.GameImg,
		Balance:     info.Account.Balance,
		LockBalance: info.Account.LockBalance,
		PackageID:   info.Base.PackageID,
	})
	common.Debug_log("登陆成功:%v", info.Base.ID)
}

// 用戶登出 client -> centerserver
func CS2C_Logout(data ResponseData) {
	common.Debug_log("loginModule CS2C_Logout")
	if data.Code != 200 {
		return
	}
	info := makeAuthResponse(data)
	if info == nil {
		return
	}
	common.GetInstance().Game.Go("UserLogout", common.UserInfo{
		UserID:      info.Base.ID,
		UserName:    info.Base.GameNick,
		UserHead:    info.Base.GameImg,
		Balance:     info.Account.Balance,
		LockBalance: info.Account.LockBalance,
		PackageID:   info.Base.PackageID,
	})
	common.Debug_log("登出成功:%v", info.Base.ID)
}

// 用戶贏錢 client -> centerserver
func CS2C_WinSettlement(data ResponseData) {
	// common.Debug_log("loginModule CS2C_WinSettlement")
	if data.Code != 200 {
		return
	}
	info := makePayResponse(data.Msg)
	if info == nil {
		return
	}

	// fmt.Printf("请求增加用户资金的返回数据=%+v", info)
	offsetMoney := info.Money - info.Tax
	common.GetInstance().Game.Go("WinMoney", common.AmountFlowRes{
		UserID:      info.ID,
		Order:       info.Order,
		RoundID:     info.RoundID,
		Balance:     info.FinalBalance,
		LockBalance: info.FinalLockBalance,
		Money:       info.Money,
		TimeStamp:   info.CreateTime,
		Tax:         info.Tax,
		DiffMoney:   offsetMoney,
	})
	common.Debug_log("用户赢钱%f", offsetMoney)
	// if offsetMoney >= 100 { // 贏錢推播
	// 	C2CS_NoticeBroadcast(info.ID, info.GameNick, offsetMoney)
	// }
}

// 用戶輸錢 client -> centerserver
func CS2C_LoseSettlement(data ResponseData) {
	// common.Debug_log("loginModule CS2C_LoseSettlement")
	if data.Code != 200 {
		return
	}
	info := makePayResponse(data.Msg)
	if info == nil {
		return
	}

	// fmt.Printf("请求扣除用户资金的返回数据=%+v", info)
	common.GetInstance().Game.Go("LoseMoney", common.AmountFlowRes{
		UserID:      info.ID,
		Order:       info.Order,
		RoundID:     info.RoundID,
		Balance:     info.FinalBalance,
		LockBalance: info.FinalLockBalance,
		Money:       info.Money,
		TimeStamp:   info.CreateTime,
		DiffMoney:   info.LockMoney - math.Abs(info.Money),
	})
	common.Debug_log("用户输钱" + common.FloatToStr(info.Money))
}

// 用戶錢鎖定 client -> centerserver
func CS2C_LockSettlement(data ResponseData) {
	// common.Debug_log("loginModule CS2C_LockSettlement")
	if data.Code != 200 {
		return
	}
	info := makePayResponse(data.Msg)
	if info == nil {
		return
	}
	// fmt.Printf("解析出的锁定资金返回数据=%+v", info)
	common.GetInstance().Game.Go("LockMoney", common.AmountFlowRes{
		UserID:      info.ID,
		Order:       info.Order,
		RoundID:     info.RoundID,
		Balance:     info.FinalBalance,
		LockBalance: info.FinalLockBalance,
		Money:       info.FinalLockBalance - info.LockBalance,
		DiffMoney:   info.LockBalance - info.FinalLockBalance,
		TimeStamp:   info.CreateTime,
	})
}

// 用戶錢解鎖 client -> centerserver
func CS2C_UnLockSettlement(data ResponseData) {
	// common.Debug_log("loginModule CS2C_UnLockSettlement")
	if data.Code != 200 {
		return
	}
	info := makePayResponse(data.Msg)
	if info == nil {
		return
	}
	common.GetInstance().Game.Go("UnLockMoney", common.AmountFlowRes{
		UserID:      info.ID,
		Order:       info.Order,
		RoundID:     info.RoundID,
		Balance:     info.FinalBalance,
		LockBalance: info.FinalLockBalance,
		Money:       info.Money,
		DiffMoney:   info.LockBalance - info.FinalLockBalance,
		TimeStamp:   info.CreateTime,
	})
}
