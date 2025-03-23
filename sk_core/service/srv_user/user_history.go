package srv_user

import "sync"

type UserBuyHistory struct {
	History map[int]int
	Lock    sync.RWMutex
}

func (u *UserBuyHistory) GetProductBuyCount(productId int) int {
	u.Lock.RLock()
	defer u.Lock.RUnlock()

	count, _ := u.History[productId]
	return count
}

func (u *UserBuyHistory) Add(productId, count int) {
	u.Lock.Lock()
	defer u.Lock.Unlock()

	cur, ok := u.History[productId]
	if !ok {
		cur = count
	} else {
		cur += count
	}

	u.History[productId] = cur
}
