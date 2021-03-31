package internal

import (
	"caidaxiao/msg"
	"errors"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"strconv"
	"sync"
	"time"
)

type GameHall struct {
	UserRecord sync.Map          // 用户记录
	RoomRecord sync.Map          // 房间记录
	roomList   []*Room           // 房间列表
	UserRoom   map[string]string // 用户房间
}

func NewHall() *GameHall {
	return &GameHall{
		UserRecord: sync.Map{},
		RoomRecord: sync.Map{},
		roomList:   make([]*Room, 0),
		UserRoom:   make(map[string]string),
	}
}

func (hall *GameHall) Init() { // 大厅初始化增加一个房间
	// 创建大厅游戏房间
	hall.CreateGameRoom()
}

//LoadHallRobots 为每个房间装载机器人
func (hall *GameHall) LoadHallRobots(num int) {
	for _, room := range hall.roomList {
		if room != nil {
			room.LoadRoomRobots(num)
		}
	}
}

//ReplacePlayerAgent 替换用户链接
func (hall *GameHall) ReplacePlayerAgent(Id string, agent gate.Agent) error {
	log.Debug("用户重连或顶替，正在替换agent %+v", Id)
	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
	if v, ok := hall.UserRecord.Load(Id); ok {
		//ErrorResp(agent, msg.ErrorMsg_UserRemoteLogin, "异地登录")
		user := v.(*Player)
		user.ConnAgent = agent
		user.ConnAgent.SetUserData(v)
		return nil
	} else {
		return errors.New("用户不在记录中~")
	}
}

//agentExist 链接是否已经存在 (是否开销过大？后续可通过新增记录解决)
func (hall *GameHall) agentExist(a gate.Agent) bool {
	var exist bool
	hall.UserRecord.Range(func(key, value interface{}) bool {
		u := value.(*Player)
		if u.ConnAgent == a {
			exist = true
		}
		return true
	})
	return exist
}

func (hall *GameHall) CreateGameRoom() {
	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		r := &Room{}
		r.Init()
		ri := i + 1
		r.RoomId = strconv.Itoa(ri)
		hall.roomList = append(hall.roomList, r)
		hall.RoomRecord.Store(r.RoomId, r)
		log.Debug("CreateRoom 创建新的房间:%v,当前房间数量:%v", r.RoomId, len(hall.roomList))

	}
	// 加载机器人
	gRobotCenter.Start()
}

func (hall *GameHall) PlayerJoinRoom(rid string, p *Player) {
	v, _ := hall.RoomRecord.Load(rid)
	if v != nil {
		r := v.(*Room)
		for i, userId := range r.UserLeave {
			// 把玩家从掉线列表中移除
			if userId == p.Id {
				log.Debug("AllocateUser 清除玩家记录~:%v", len(r.UserLeave))
				r.UserLeave = append(r.UserLeave[:i], r.UserLeave[i+1:]...)
				log.Debug("AllocateUser 清除玩家记录~:%v", userId)
				log.Debug("AllocateUser 清除玩家记录~:%v", len(r.UserLeave))
				break
			}
		}
	}

	// 处理重连
	for _, r := range hall.roomList {
		for _, v := range r.PlayerList {
			if v != nil && v.Id == p.Id {
				roomData := r.RespRoomData()
				enter := &msg.EnterRoom_S2C{}
				enter.RoomData = roomData
				if r.GameStat == msg.GameStep_Banker {
					enter.RoomData.GameTime = BankerTime - r.counter
				} else if r.GameStat == msg.GameStep_Banker2 {
					enter.RoomData.GameTime = Banker2Time - r.counter
				} else if r.GameStat == msg.GameStep_DownBet {
					enter.RoomData.GameTime = DownBetTime - r.counter
				} else if r.GameStat == msg.GameStep_Settle {
					enter.RoomData.GameTime = SettleTime - r.counter
				}
				p.SendMsg(enter)

			}
		}
	}

	r, _ := hall.RoomRecord.Load(rid)
	if r != nil {
		room := r.(*Room)
		room.JoinGameRoom(p)
	}
}
