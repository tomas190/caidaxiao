package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/gate"
)

type Player struct {
	// 玩家代理链接
	ConnAgent gate.Agent

	Id       string
	NickName string
	HeadImg  string
	Account  float64
	Password string
	Token    string
	RoundId  string

	Status          msg.PlayerStatus     // 玩家状态
	BankerMoney     float64              // 庄家金额
	BankerCount     int32                // 庄家次数
	BankerStatus    msg.BankerStatus     // 庄家状态
	DownBetMoney    msg.DownBetMoney     // 玩家各注池下注金额
	ResultMoney     float64              // 结算金额
	WinResultMoney  float64              // 本局赢钱金额
	LoseResultMoney float64              // 本局输钱金额
	TotalDownBet    int32                // 房间下注总金额
	WinTotalCount   int32                // 玩家房间获胜Win总次数
	TwentyData      []int32              // 20局Win数据,1Lose,2Win
	DownBetHistory  []msg.DownBetHistory // 下注记录 70条
	IsBanker        bool                 // 是否庄家
	IsDownBanker    bool                 // 是否下庄
	IsAction        bool                 // 玩家是否行动
	IsRobot         bool                 // 是否机器人
	IsOnline        bool                 // 玩家是否在线

}

func (p *Player) Init() {
	p.RoundId = ""
	p.BankerMoney = 0
	p.Status = msg.PlayerStatus_XX_Status
	p.BankerStatus = msg.BankerStatus_BankerNot
	p.DownBetMoney = msg.DownBetMoney{}
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.TotalDownBet = 0
	p.WinTotalCount = 0
	p.TwentyData = nil
	p.DownBetHistory = make([]msg.DownBetHistory, 0)
	p.IsDownBanker = false
	p.IsBanker = false
	p.IsAction = false
	p.IsRobot = false
	p.IsOnline = true
}

//SendMsg 玩家向客户端发送消息
func (p *Player) SendMsg(msg interface{}) {
	if p.ConnAgent != nil {
		p.ConnAgent.WriteMsg(msg)
	}
}

func (p *Player) RespPlayerData() *msg.PlayerData {
	pd := &msg.PlayerData{}
	pd.PlayerInfo = new(msg.PlayerInfo)
	pd.PlayerInfo.Id = p.Id
	pd.PlayerInfo.NickName = p.NickName
	pd.PlayerInfo.HeadImg = p.HeadImg
	pd.PlayerInfo.Account = p.Account
	pd.BankerMoney = p.BankerMoney
	pd.BankerCount = p.BankerCount
	pd.DownBetMoney = new(msg.DownBetMoney)
	pd.DownBetMoney.BigDownBet = p.DownBetMoney.BigDownBet
	pd.DownBetMoney.SmallDownBet = p.DownBetMoney.SmallDownBet
	pd.DownBetMoney.SingleDownBet = p.DownBetMoney.SingleDownBet
	pd.DownBetMoney.DoubleDownBet = p.DownBetMoney.DoubleDownBet
	pd.DownBetMoney.PairDownBet = p.DownBetMoney.PairDownBet
	pd.DownBetMoney.StraightDownBet = p.DownBetMoney.StraightDownBet
	pd.DownBetMoney.LeopardDownBet = p.DownBetMoney.LeopardDownBet
	pd.TotalDownBet = p.TotalDownBet
	pd.WinTotalCount = p.WinTotalCount
	pd.ResultMoney = p.ResultMoney
	pd.IsAction = p.IsAction
	pd.IsBanker = p.IsBanker
	pd.IsRobot = p.IsRobot
	return pd
}
