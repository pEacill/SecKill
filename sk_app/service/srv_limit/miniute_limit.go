package srv_limit

type MinLimit struct {
	count   int
	curTime int64
}

func (m *MinLimit) Count(nowTime int64) (curCount int) {
	if nowTime-m.curTime > 60 {
		m.count = 1
		m.curTime = nowTime
		curCount = m.count
		return
	}

	m.count++
	curCount = m.count
	return
}

func (m *MinLimit) Check(nowTime int64) int {
	if nowTime-m.curTime > 60 {
		return 0
	}
	return m.count
}
