package srv_redis

import (
	"crypto/md5"
	"fmt"
	"log"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_core/config"
	"github.com/pEacill/SecKill/sk_core/model"
	"github.com/pEacill/SecKill/sk_core/service/srv_error"
	"github.com/pEacill/SecKill/sk_core/service/srv_user"
)

func HandleUser() {
	log.Println("Handle user running.")

	for req := range config.SecLayerCtx.Read2HandleChan {
		log.Printf("Begin process request: {Req: %v}", req)
		res, err := HandleSeckill(req)
		if err != nil {
			log.Printf("Process {Req: %v} failed, with error: %v", req, err)
			res = &model.SecResult{
				Code: srv_error.ErrServiceBusy,
			}
		}

		fmt.Println("Processing ...", res)
		timer := time.NewTicker(time.Millisecond * time.Duration(conf.SecKill.SendToWriteChanTimeout))
		select {
		case config.SecLayerCtx.Handle2WriteChan <- res:
		case <-timer.C:
			log.Printf("Send {Result: %v} to response chan timeout.", res)
			break
		}
	}
	return
}

func HandleSeckill(req *model.SecRequest) (res *model.SecResult, err error) {
	config.SecLayerCtx.RWSecProductLock.RLock()
	defer config.SecLayerCtx.RWSecProductLock.RUnlock()

	res = &model.SecResult{}
	res.ProductId, res.UserId = req.ProductId, req.UserId

	product, ok := conf.SecKill.SecProductInfoMap[req.ProductId]
	if !ok {
		log.Printf("{Product: %v} not found.", req.ProductId)
		res.Code = srv_error.ErrNotFoundProduct
		return
	}

	if product.Status == srv_error.ProductStatusSoldout {
		res.Code = srv_error.ErrSoldout
		return
	}

	nowTime := time.Now().Unix()
	config.SecLayerCtx.HistoryMapLock.Lock()
	userHistory, ok := config.SecLayerCtx.HistoryMap[req.UserId]
	if !ok {
		userHistory = &srv_user.UserBuyHistory{
			History: make(map[int]int, 16),
		}
		config.SecLayerCtx.HistoryMap[req.UserId] = userHistory
	}
	historyCount := userHistory.GetProductBuyCount(req.ProductId)
	config.SecLayerCtx.HistoryMapLock.Unlock()

	if historyCount >= product.OnePersonBuyLimit {
		res.Code = srv_error.ErrAlreadyBuy
		return
	}

	curSoldCount := config.SecLayerCtx.ProductCountMgr.Count(req.ProductId)
	if curSoldCount >= product.Total {
		res.Code = srv_error.ErrSoldout
		product.Status = srv_error.ProductStatusSoldout
		return
	}

	curRate := 0.1
	fmt.Println(curRate, product.BuyRate)
	if curRate > product.BuyRate {
		res.Code = srv_error.ErrRetry
		return
	}

	userHistory.Add(req.ProductId, 1)
	config.SecLayerCtx.ProductCountMgr.Add(req.ProductId, 1)

	res.Code = srv_error.ErrSecKillSucc
	tokenData := fmt.Sprintf("userId=%d&productId=%d&timestamp=%d&security=%s", req.UserId, req.ProductId, nowTime, conf.SecKill.TokenPassWd)
	res.Token = fmt.Sprintf("%x", md5.Sum([]byte(tokenData)))
	res.TokenTime = nowTime
	return
}
