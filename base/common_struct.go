package base

const VersionCode = "1.0.13" //服务版本

type UserInfo struct {
	UserID      int32
	UserName    string
	UserHead    string
	Balance     float64
	LockBalance float64
	PackageID   int // 不同平台id(税收)
}

//资金周转
type AmountFlowReq struct {
	UserID     int32
	TimeStamp  int64
	Money      float64
	BetMoney   float64
	LockMoney  float64
	Order      string
	RoundID    string
	PackageID  int    // 不同平台id(税收)
	Reason     string //资金变动原因
	UserName   string //用户名
	RoomNumber string //房间号码
}

// 登入返回税收
type LoginResponse struct {
	PackageID  int     `json:"package_id"`
	TaxPercent float64 `json:"platform_tax_percent"`
}

// 更新餘額結構體
type AmountFlowRes struct {
	UserID      int32
	TimeStamp   int64
	Money       float64 //资金变动额
	Balance     float64 //余额
	LockBalance float64 //锁定余额
	Tax         float64 //扣税金额
	DiffMoney   float64 //资金变动差值
	Order       string  //
	RoundID     string  //红包ID
	Reason      string  //资金变动原因
}

// LogCenter 日志中心数据结构
type LogCenter struct {
	Type     string `json:"type"`      //"LOG"|"ERR"|"DEG",
	From     string `json:"from"`      //"game-server",
	GameName string `json:"game_name"` // "lunpan"
	ID       int32  `json:"id"`        //用户ID
	Host     string `json:"host"`      //服务IP地址,
	File     string `json:"file"`      //文件名
	Line     int    `json:"line"`      //行号,
	Msg      string `json:"msg"`
	Time     string `json:"time"` // 时间(YYYY-MM-DD HH:II:SS),
}

type TurnoverInfo struct {
	Kind         string
	Turnoverdata AmountFlowReq
}
