package internal

import (
	"caidaxiao/msg"
	"github.com/name5566/leaf/log"
)

func (r *Room) GetType(numSlice []int) {
	log.Debug("numSlice %v", numSlice)
	// 判断豹子
	if numSlice[0] == 9 {
		if numSlice[1] == 9 {
			if numSlice[2] == 9 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 8 {
		if numSlice[1] == 8 {
			if numSlice[2] == 8 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 7 {
		if numSlice[1] == 7 {
			if numSlice[2] == 7 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 6 {
		if numSlice[1] == 6 {
			if numSlice[2] == 6 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 5 {
		if numSlice[1] == 5 {
			if numSlice[2] == 5 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 4 {
		if numSlice[1] == 4 {
			if numSlice[2] == 4 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 3 {
		if numSlice[1] == 3 {
			if numSlice[2] == 3 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 2 {
		if numSlice[1] == 2 {
			if numSlice[2] == 2 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 1 {
		if numSlice[1] == 1 {
			if numSlice[2] == 1 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	if numSlice[0] == 0 {
		if numSlice[1] == 0 {
			if numSlice[2] == 0 {
				r.LotteryResult.CardType = msg.CardsType_Leopard
				return
			}
		}
	}
	//// 判断顺子
	//if numSlice[0] == 1 {
	//	if numSlice[1] == 2 {
	//		if numSlice[2] == 3 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 2 {
	//	if numSlice[1] == 3 {
	//		if numSlice[2] == 4 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 3 {
	//	if numSlice[1] == 4 {
	//		if numSlice[2] == 5 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 4 {
	//	if numSlice[1] == 5 {
	//		if numSlice[2] == 6 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 5 {
	//	if numSlice[1] == 6 {
	//		if numSlice[2] == 7 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 6 {
	//	if numSlice[1] == 7 {
	//		if numSlice[2] == 8 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//if numSlice[0] == 7 {
	//	if numSlice[1] == 8 {
	//		if numSlice[2] == 9 {
	//			r.LotteryResult.CardType = msg.CardsType_Straight
	//			return
	//		}
	//	}
	//}
	//// 判断对子
	//if findRepeatNumber(numSlice) != -1 {
	//	r.LotteryResult.CardType = msg.CardsType_Pair
	//	return
	//}
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
