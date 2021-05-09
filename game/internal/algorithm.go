package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/log"
)

func (r *Room) GetType(numSlice []int) {
	log.Debug("numSlice %v", numSlice)
	// 判断豹子
	if numSlice[1] == 9 {
		if numSlice[2] == 9 {
			if numSlice[4] == 9 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 8 {
		if numSlice[2] == 8 {
			if numSlice[4] == 8 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 7 {
		if numSlice[2] == 7 {
			if numSlice[4] == 7 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 6 {
		if numSlice[2] == 6 {
			if numSlice[4] == 6 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 5 {
		if numSlice[2] == 5 {
			if numSlice[4] == 5 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 4 {
		if numSlice[2] == 4 {
			if numSlice[4] == 4 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 3 {
		if numSlice[2] == 3 {
			if numSlice[4] == 3 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 2 {
		if numSlice[2] == 2 {
			if numSlice[4] == 2 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 1 {
		if numSlice[2] == 1 {
			if numSlice[4] == 1 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[1] == 0 {
		if numSlice[2] == 0 {
			if numSlice[4] == 0 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
}

func findRepeatNumber(nums []int) int {
	maps := make(map[int]bool)
	for _, num := range nums {
		if maps[num] {
			return num
		} else {
			maps[num] = true
		}
	}
	return -1
}
