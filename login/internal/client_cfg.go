package internal

// 與中心服傳輸用
type UserBase struct {
	id           int     `json:"_id"`
	ID           int32   `json:"id"`
	UUID         string  `json:"uuid"`
	GameNick     string  `json:"game_nick"`
	ProxyUserID  int     `json:"proxy_user_id"`
	GameGold     float64 `json:"game_gold"` // 玩家擁有的錢
	BankGold     float64 `json:"bank_gold"` // 玩家當莊家的錢(從LockGold來)
	LockGold     float64 `json:"lock_gold"` // 玩家若要當莊家要鎖住的錢
	ChangeTime   int     `json:"change_time"`
	LoginTime    int     `json:"login_time"`
	LoginIP      string  `json:"login_ip"`
	ReginTime    int     `json:"regin_time"`
	ReginIP      string  `json:"regin_ip"`
	PhoneNumber  string  `json:"phone_number"`
	Status       int     `json:"status"`
	GameUserType int     `json:"game_user_type"`
	GameImg      string  `json:"game_img"`
	PackageID    int     `json:"package_id"`
	DeviceID     string  `json:"device_id"`
}

type UserAccount struct {
	Balance       float64 `json:"balance"` //所有余额,包括锁定金额
	GameName      string  `json:"game_name"`
	LockBalance   float64 `json:"lock_balance"` //锁定金额
	Prepayments   float64 `json:"prepayments"`
	BetMoney      float64 `json:"bet_money"`
	WinMoney      float64 `json:"win_money"`
	LoseMoney     float64 `json:"lose_money"`
	BetTimes      int     `json:"bet_times"`
	WinTimes      int     `json:"win_times"`
	Status        int     `json:"status"`
	BankerBalance float64 `json:"banker_balance"`
}

type UserResponse struct {
	GameID  string      `json:"game_id"`
	Base    UserBase    `json:"game_user"`
	Account UserAccount `json:"game_account"`
}

type AuthUserPsw struct {
	UserID   int32  `json:"id,int"`
	GameID   string `json:"game_id"`
	DevKey   string `json:"dev_key"`
	PassWord string `json:"password"`
	DevName  string `json:"dev_name"`
}
type AuthUserToken struct {
	UserID  int32  `json:"id,int"`
	GameID  string `json:"game_id"`
	DevKey  string `json:"dev_key"`
	DevName string `json:"dev_name"`
	Token   string `json:"token"`
}

//////////////////////////////////////////////////////////////

//用于验证的数据
type AuthData struct {
	//Token  string `json:"token"`
	DevKey  string `json:"dev_key"`
	DevName string `json:"dev_name"`
}

//流水相关数据
type PayData struct {
	UserID     int32   `json:"id,int"`
	CreateTime int64   `json:"create_time"`
	Money      float64 `json:"money"`
	BetMoney   float64 `json:"bet_money"`
	LockMoney  float64 `json:"lock_money"`
	PreMoney   float64 `json:"pre_money"`
	PayReason  string  `json:"pay_reason"`
	OrderID    string  `json:"order"` //自己创建一个唯一ID,方便之后查询
	GameID     string  `json:"game_id"`
	RoundID    string  `json:"round_id"` //唯一ID,用于识别多人是否在同一局游戏
}

//资金变动请求
type PayRequest struct {
	Auth AuthData `json:"auth"`
	Info PayData  `json:"info"`
}

type PayResponse struct {
	Balance            float64 `json:"balance"`
	BankerBalance      float64 `json:"banker_balance"`
	CreateTime         int64   `json:"create_time"`
	DevBrandName       string  `json:"dev_brand_name"`
	DevID              int     `json:"dev_id"`
	FinalBalance       float64 `json:"final_balance"`
	FinalBankerBalance float64 `json:"final_banker_balance"`
	FinalLockBalance   float64 `json:"final_lock_balance"`
	FinalPay           float64 `json:"final_pay"`
	FinalPrepayments   float64 `json:"final_prepayments"`
	GameAccountStatus  int     `json:"game_account_status"`
	GameID             string  `json:"game_id"`
	GameName           string  `json:"game_name"`
	GameNick           string  `json:"game_nick"`
	GameServerIP       string  `json:"game_server_ip"`
	GameUserStatus     int     `json:"game_user_status"`
	GameUserType       int     `json:"game_user_type"`
	ID                 int32   `json:"id"`
	Income             float64 `json:"income"`
	IsBanker           int     `json:"is_banker"`
	LockBalance        float64 `json:"lock_balance"`
	LockMoney          float64 `json:"lock_money"`
	LoginIP            string  `json:"login_ip"`
	Money              float64 `json:"money"`
	Order              string  `json:"order"`
	PackageID          int     `json:"package_id"`
	PlatformTaxPercent float64 `json:"platform_tax_percent"`
	Prepayments        float64 `json:"prepayments"`
	ProxyUserID        int     `json:"proxy_user_id"`
	RoundID            string  `json:"round_id"`
	Status             int     `json:"status"`
	Tax                float64 `json:"tax"`
	UUID               string  `json:"uuid"`
}

type NoticeMessage struct {
	ServerKind int    `json:"type"`
	Message    string `json:"message"`
	Title      string `json:"title"`
}

type NoticeRequest struct {
	// AuthUser
	// Message NoticeMessage `json:"msg"`
	UserID   int32  `json:"id,int"`
	GameID   string `json:"game_id"`
	DevKey   string `json:"dev_key"`
	PassWord string `json:"password"`
	DevName  string `json:"dev_name"`
	Type     int    `json:"type"`
	Message  string `json:"message"`
	Topic    string `json:"topic"`
}
