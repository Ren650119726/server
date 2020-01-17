package room

import (
	"testing"
)

func Test(t *testing.T) {
	schedule_bomb([]int64{}, []int64{158, 147, 133, 188, 63, 100, 75, 71, 43, 22}, 2, 6)
}

// 雷数调控
func schedule_bomb(bomb, normal []int64, need_count, bomb_num int8) (bombdeal, normaldeal []int64) {
	surplus_bomb := int8(len(bomb)) - need_count
	if surplus_bomb > 0 { // 消除雷
		total := 0
		n := int(surplus_bomb)
		for i := 0; i < n; i++ {
			total += 1
			bomb[i] -= 1
		}

		normal = append(normal, bomb[:n]...)
		bomb = bomb[n:]

		for k, v := range normal {
			if int8((v+int64(total))%10) != bomb_num {
				normal[k] += int64(total)
				break
			}
		}
	} else { // 加雷
		total := int64(0)
		n := int(-surplus_bomb)
		for i := 0; i < n; i++ {
			t := (normal[i] % 10) - int64(bomb_num)
			normal[i] -= t
			total += t
		}

		bomb = append(bomb, normal[:n]...)
		normal = normal[n:]

		for k, v := range normal {
			if int8((v+int64(total))%10) != bomb_num {
				normal[k] += int64(total)
				break
			}
		}
	}
	return bomb, normal
}
