package srv_limit

import (
	"fmt"
	"log"
	"sync"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_app/model"
)

type SecLimitMgr struct {
	UserLimitMap map[int]*Limit
	IpLimitMap   map[string]*Limit
	lock         sync.Mutex
}

var SecLimitMgrVars = &SecLimitMgr{
	UserLimitMap: make(map[int]*Limit),
	IpLimitMap:   make(map[string]*Limit),
}

func AntiSpam(req *model.SecRequest) (err error) {
	_, ok := conf.SecKill.IDBlackMap[req.UserId]
	if ok {
		err = fmt.Errorf("Invalid Request.")
		log.Printf("{user: %v} is block by id black.", req.UserId)
		return
	}

	_, ok = conf.SecKill.IPBlackMap[req.ClientAddr]
	if ok {
		err = fmt.Errorf("Invalid Request.")
		log.Printf("{user: %v} is block by ip black.", req.UserId)
		return
	}

	var secIdCount, minIdCount, secIpCount, minIpCount int

	SecLimitMgrVars.lock.Lock()
	{
		limit, ok := SecLimitMgrVars.UserLimitMap[req.UserId]
		if !ok {
			limit = &Limit{
				secLimit: &SecLimit{},
				minLimit: &MinLimit{},
			}
			SecLimitMgrVars.UserLimitMap[req.UserId] = limit
		}

		secIdCount = limit.secLimit.Count(req.AccessTime)
		minIdCount = limit.minLimit.Count(req.AccessTime)

		limit, ok = SecLimitMgrVars.IpLimitMap[req.ClientAddr]
		if !ok {
			limit = &Limit{
				secLimit: &SecLimit{},
				minLimit: &MinLimit{},
			}
			SecLimitMgrVars.IpLimitMap[req.ClientAddr] = limit
		}

		secIpCount = limit.secLimit.Count(req.AccessTime)
		minIpCount = limit.minLimit.Count(req.AccessTime)
	}
	SecLimitMgrVars.lock.Unlock()

	if secIdCount > conf.SecKill.AccessLimitConf.UserSecAccessLimit {
		err = fmt.Errorf("Invalid Request.")
		return
	}

	if minIdCount > conf.SecKill.AccessLimitConf.UserMinAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	if secIpCount > conf.SecKill.AccessLimitConf.IPSecAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	if minIpCount > conf.SecKill.AccessLimitConf.IPMinAccessLimit {
		err = fmt.Errorf("invalid request")
		return
	}

	return
}
