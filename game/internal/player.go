package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/gate"
)

type DownBetHistory struct {
	TimeStr       string           // 开奖时间
	Lottery       []int            // 开奖数据
	LotteryResult msg.PotWinList   // 开奖结果
	DownBetMoney  msg.DownBetMoney // 注池下注金额
}

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

	Status          msg.PlayerStatus  // 玩家状态
	BankerStatus    msg.BankerStatus  // 庄家状态
	DownBetMoney    msg.DownBetMoney  // 玩家各注池下注金额
	ResultMoney     float64           // 结算金额
	WinResultMoney  float64           // 本局赢钱金额
	LoseResultMoney float64           // 本局输钱金额
	TotalDownBet    int32             // 房间下注总金额
	WinTotalCount   int32             // 玩家房间获胜Win总次数
	TwentyData      []int32           // 20局Win数据,1Lose,2Win
	DownBetHistory  []*DownBetHistory // 下注记录 10条
	IsButton        bool              // 是否庄家
	IsAction        bool              // 玩家是否行动
	IsRobot         bool              // 是否机器人
	IsOnline        bool              // 玩家是否在线
}

func (p *Player) Init() {
	p.RoundId = ""
	p.Status = msg.PlayerStatus_XX_Status
	p.BankerStatus = msg.BankerStatus_BankerNot
	p.DownBetMoney = msg.DownBetMoney{}
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.TotalDownBet = 0
	p.WinTotalCount = 0
	p.TwentyData = nil
	p.DownBetHistory = nil
	p.IsButton = false
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
