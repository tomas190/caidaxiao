package internal

// 註冊loginModule的ChanRPC，gameModule可以透過bridge中chanrpcManage的方式來調用
func initLoginChanRPC() {

	// 用戶登入登出
	skeleton.RegisterChanRPC("UserLogin", C2CS_Login)
	skeleton.RegisterChanRPC("UserLogout", C2CS_Logout)

	// 用戶贏錢，用戶輸錢
	skeleton.RegisterChanRPC("UserWinMoney", C2CS_WinSettlement)
	skeleton.RegisterChanRPC("UserLoseMoney", C2CS_LoseSettlement)

	// 用戶鎖定金錢，解鎖金錢
	skeleton.RegisterChanRPC("UserLockMoney", C2CS_LockSettlement)
	skeleton.RegisterChanRPC("UserUnLockMoney", C2CS_UnLockSettlement)

	// 向大廳推送通知(有玩家在骰寶贏得..錢)
	skeleton.RegisterChanRPC("NoticeBroadcast", C2CS_NoticeBroadcast)
}
