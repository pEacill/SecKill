package srv_limit

type TimeLimit interface {
	Count(nowTime int64) (curCount int)
	Check(nowTime int64) int
}

type Limit struct {
	secLimit TimeLimit
	minLimit TimeLimit
}
