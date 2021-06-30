package internal

import (
	"caidaxiao/msg"
)

func (r *Room) GetType(numSlice []int) {
	// log.Debug("numSlice %v", numSlice)
	// 判断豹子
	if numSlice[1] == numSlice[2] && numSlice[1] == numSlice[4] {
		r.LotteryResult.Result.CardType = msg.CardsType_Leopard
		r.LotteryResult.ResultFX.CardType = msg.CardsType_Leopard
		return
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
