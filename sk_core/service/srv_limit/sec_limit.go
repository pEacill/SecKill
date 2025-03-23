package srv_limit

type SecLimit struct {
	count   int
	preTime int64
}

func (s *SecLimit) Count(nowTime int64) (curCount int) {
	if s.preTime != nowTime {
		s.count = 1
		s.preTime = nowTime
		curCount = s.count
		return
	}

	s.count++
	curCount = s.count
	return
}

func (s *SecLimit) Check(nowTime int64) int {
	if s.preTime != nowTime {
		return 0
	}
	return s.count
}
