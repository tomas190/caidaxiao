package internal

import (
	"strconv"
	"sync"
	"time"

	"github.com/name5566/leaf/log"
)

type GameHall struct {
	UserRecord sync.Map // 用户记录(玩家在線列表) key:userId,value:*Player
	RoomRecord sync.Map // 房间记录(房間開啟列表) key:roomid,value:*Room
	roomList   []*Room  // 房间列表
	UserRoom   sync.Map // 用户房间 key:userId,value:RoomId
}

func NewHall() *GameHall {
	return &GameHall{
		UserRecord: sync.Map{},
		RoomRecord: sync.Map{},
		roomList:   make([]*Room, 0),
		UserRoom:   sync.Map{},
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
// func (hall *GameHall) ReplacePlayerAgent(Id string, agent gate.Agent) error {
// 	log.Debug("用户重连或顶替，正在替换agent %+v", Id)
// 	// tip 这里会拷贝一份数据，需要替换的是记录中的，而非拷贝数据中的，还要注意替换连接之后要把数据绑定到新连接上
// 	if v, ok := hall.UserRecord.Load(Id); ok {
// 		//ErrorResp(agent, msg.ErrorMsg_UserRemoteLogin, "异地登录")
// 		user := v.(*Player)
// 		user.ConnAgent = agent
// 		user.ConnAgent.SetUserData(v)
// 		return nil
// 	} else {
// 		return errors.New("用户不在记录中~")
// 	}
// }

//agentExist 链接是否已经存在 (是否开销过大？后续可通过新增记录解决)
// func (hall *GameHall) agentExist(a gate.Agent) bool {
// 	var exist bool
// 	hall.UserRecord.Range(func(key, value interface{}) bool {
// 		u := value.(*Player)
// 		if u.ConnAgent == a {
// 			exist = true
// 		}
// 		return true
// 	})
// 	return exist
// }

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
		r.GetRoomType()
	}
	// 加载机器人
	gRobotCenter.Start()
}

func (hall *GameHall) PlayerJoinRoom(rid string, p *Player) {

	r, _ := hall.RoomRecord.Load(rid)
	if r != nil {
		room := r.(*Room)
		if room.IsOpenRoom == true {
			// 把玩家从掉线列表中移除
			for i, userId := range room.UserLeave {
				if userId == p.Id {
					room.UserLeave = append(room.UserLeave[:i], room.UserLeave[i+1:]...)
					break
				}
			}
			room.JoinGameRoom(p)
		}
	}
}
