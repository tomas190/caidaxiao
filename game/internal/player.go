package internal

import (
	common "caidaxiao/base"
	"caidaxiao/msg"
)

type Player struct {
	// 玩家代理链接
	// ConnAgent gate.Agent

	Id       int32
	NickName string
	HeadImg  string
	Account  float64
	// Password  string
	// Token     string
	RoundId   string
	PackageId int

	Status          msg.PlayerStatus      // 玩家状态
	BankerMoney     float64               // 庄家金额
	BankerCount     int32                 // 庄家次数
	BankerStatus    msg.BankerStatus      // 庄家状态
	DownBetMoney    *msg.DownBetMoney     // 玩家各注池下注金额
	ResultMoney     float64               // 结算金额
	WinResultMoney  float64               // 本局赢钱金额
	LoseResultMoney float64               // 本局输钱金额
	TotalDownBet    int32                 // 房间下注总金额
	WinTotalCount   int32                 // 玩家房间获胜Win总次数
	TwentyData      []int32               // 20局Win数据,1Lose,2Win
	DownBetHistory  []*msg.DownBetHistory // 下注记录 70条
	IsBanker        bool                  // 是否庄家
	IsDownBanker    bool                  // 是否下庄
	IsAction        bool                  // 玩家是否行动
	IsRobot         bool                  // 是否机器人
	IsOnline        bool                  // 玩家是否在线

	MinBet int32 // 限定下注最小金额
	MaxBet int32 // 限定下注最大金额
}

func (p *Player) Init() {
	p.RoundId = ""
	p.BankerMoney = 0
	p.Status = msg.PlayerStatus_XX_Status
	p.BankerStatus = msg.BankerStatus_BankerNot
	p.DownBetMoney = &msg.DownBetMoney{}
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0
	p.TotalDownBet = 0
	p.WinTotalCount = 0
	p.TwentyData = nil
	p.DownBetHistory = make([]*msg.DownBetHistory, 0)
	p.IsDownBanker = false
	p.IsBanker = false
	p.IsAction = false
	p.IsRobot = false
	p.IsOnline = true
	p.MinBet = 0
	p.MaxBet = 0
}

//SendMsg 玩家向客户端发送消息
func (p *Player) SendMsg(msg interface{}, event string) {

	if p.IsRobot != true {
		// if  event != "SendActTime_S2C" && event != "PlayerAction_S2C" && event != "PotChangeMoney_S2C" && event != "JoinRoom_S2C" && event != "ActionTime_S2C" && event != "ResultData_S2C" { //過濾:1.反回遊戲時間(表演) 2.玩家、機器人下注(表演) 3.更新注池金额 4.加入房間 5.結算資料
		// 	log.Debug("Send To Client playerID: %v Event:%v Message : %v", p.Id, event, msg)
		// }
		client, ok := AgentFromuserID_.Load(p.Id)
		if !ok {
			return
		}
		client.(*ClientInfo).agent.WriteMsg(msg) //通知用戶 client.agent 前端踢掉
	}
}

func (p *Player) RespPlayerData() *msg.PlayerData {
	pd := &msg.PlayerData{}
	pd.PlayerInfo = new(msg.PlayerInfo)
	pd.PlayerInfo.Id = common.Int32ToStr(p.Id)
	pd.PlayerInfo.NickName = p.NickName
	pd.PlayerInfo.HeadImg = p.HeadImg
	pd.PlayerInfo.Account = p.Account
	pd.BankerMoney = p.BankerMoney
	pd.BankerCount = p.BankerCount
	pd.DownBetMoney = p.DownBetMoney
	// pd.DownBetMoney = new(msg.DownBetMoney)
	// pd.DownBetMoney.BigDownBet = p.DownBetMoney.BigDownBet
	// pd.DownBetMoney.SmallDownBet = p.DownBetMoney.SmallDownBet
	// pd.DownBetMoney.SingleDownBet = p.DownBetMoney.SingleDownBet
	// pd.DownBetMoney.DoubleDownBet = p.DownBetMoney.DoubleDownBet
	// pd.DownBetMoney.PairDownBet = p.DownBetMoney.PairDownBet
	// pd.DownBetMoney.StraightDownBet = p.DownBetMoney.StraightDownBet
	// pd.DownBetMoney.LeopardDownBet = p.DownBetMoney.LeopardDownBet
	pd.DownBetHistory = p.DownBetHistory
	// for _, v := range p.DownBetHistory {
	// 	his := &msg.DownBetHistory{}
	// 	his.TimeFmt = v.TimeFmt
	// 	his.ResNum = v.ResNum
	// 	his.Result = v.Result
	// 	his.BigSmall = v.BigSmall
	// 	his.SinDouble = v.SinDouble
	// 	his.CardType = v.CardType
	// 	his.DownBetMoney = v.DownBetMoney
	// 	pd.DownBetHistory = append(pd.DownBetHistory, his)
	// }
	pd.TotalDownBet = p.TotalDownBet
	pd.WinTotalCount = p.WinTotalCount
	pd.ResultMoney = p.ResultMoney
	pd.IsAction = p.IsAction
	pd.IsBanker = p.IsBanker
	pd.IsRobot = p.IsRobot
	return pd
}
