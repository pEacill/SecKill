package service

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_app/config"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/pEacill/SecKill/sk_app/service/srv_err"
	"github.com/pEacill/SecKill/sk_app/service/srv_limit"
)

type Service interface {
	HealthCheck() bool
	SecInfo(productId int) (data map[string]interface{})
	SecKill(req *model.SecRequest) (map[string]interface{}, int, error)
	SecInfoList() ([]map[string]interface{}, int, error)
}

type SkAppService struct{}

func (s SkAppService) HealthCheck() bool {
	return true
}

func (s SkAppService) SecInfo(productId int) (data map[string]interface{}) {
	config.SkAppContext.RWSecProductLock.RLock()
	defer config.SkAppContext.RWSecProductLock.RUnlock()

	v, ok := conf.SecKill.SecProductInfoMap[productId]
	if !ok {
		return nil
	}

	data = make(map[string]interface{})
	data["product_id"] = productId
	data["start_time"] = v.StartTime
	data["end_time"] = v.EndTime
	data["status"] = v.Status

	return data
}

func (s SkAppService) SecInfoList() ([]map[string]interface{}, int, error) {
	config.SkAppContext.RWSecProductLock.RLock()
	defer config.SkAppContext.RWSecProductLock.RUnlock()

	var data []map[string]interface{}
	for _, v := range conf.SecKill.SecProductInfoMap {
		item, _, err := s.SecInfoById(v.ProductId)
		if err != nil {
			log.Printf("Get SecKill info error: %v", err)
			continue
		}
		data = append(data, item)
	}
	return data, 0, nil
}

func (s SkAppService) SecInfoById(productId int) (map[string]interface{}, int, error) {
	config.SkAppContext.RWSecProductLock.RLock()
	defer config.SkAppContext.RWSecProductLock.RUnlock()

	var code int
	v, ok := conf.SecKill.SecProductInfoMap[productId]
	if !ok {
		return nil, srv_err.ErrNotFoundProductId, fmt.Errorf("Not found Product id: {product id: %v}", productId)
	}

	start := false
	end := false
	status := "success"
	var err error
	nowTime := time.Now().Unix()

	if nowTime-v.StartTime < 0 {
		start, end = false, false
		status = "SecKill not start."
		code = srv_err.ErrActiveNotStart
		err = fmt.Errorf(status)
	}

	if nowTime-v.StartTime > 0 {
		start = true
	}

	if nowTime-v.EndTime > 0 {
		end = true
		status = "SecKill is already end."
		code = srv_err.ErrActiveAlreadyEnd
		err = fmt.Errorf(status)
	}

	if v.Status == config.ProductStatusSaleOut || v.Status == config.ProductStatusForceSaleOut {
		start, end = false, false
		status = "Product is sale out."
		code = srv_err.ErrActiveSaleOut
		err = fmt.Errorf(status)
	}

	curRate := rand.Float64()

	if curRate > v.BuyRate*1.5 {
		start, end = false, false
		status = "retry"
		code = srv_err.ErrRetry
		err = fmt.Errorf(status)
	}

	data := map[string]interface{}{
		"product_id": productId,
		"start":      start,
		"end":        end,
		"status":     status,
	}

	return data, code, err
}

func (s SkAppService) SecKill(req *model.SecRequest) (map[string]interface{}, int, error) {
	config.SkAppContext.RWSecProductLock.RLock()
	defer config.SkAppContext.RWSecProductLock.RUnlock()

	var code int
	err := srv_limit.AntiSpam(req)
	if err != nil {
		code = srv_err.ErrUserServiceBusy
		log.Printf("{User: %v} AntiSpam failed, {req: %v}", req.UserId, req)
		return nil, code, err
	}

	data, code, err := s.SecInfoById(req.ProductId)
	if err != nil {
		log.Printf("{User: %v} SecInfoById failed, {req: %v}", req.UserId, req)
		return nil, code, err
	}

	userKey := fmt.Sprintf("%d_%d", req.UserId, req.ProductId)
	ResultChan := make(chan *model.SecResult, 1)
	config.SkAppContext.UserConnMapLock.Lock()
	config.SkAppContext.UserConnMap[userKey] = ResultChan
	config.SkAppContext.UserConnMapLock.Unlock()

	config.SkAppContext.SecReqChan <- req
	ticker := time.NewTicker(time.Millisecond * time.Duration(conf.SecKill.AppWaitResultTimeout))

	defer func() {
		ticker.Stop()
		config.SkAppContext.UserConnMapLock.Lock()
		delete(config.SkAppContext.UserConnMap, userKey)
		config.SkAppContext.UserConnMapLock.Unlock()
	}()

	select {
	case <-ticker.C:
		code = srv_err.ErrProcessTimeout
		err = fmt.Errorf("Request timeout.")
		return nil, code, err
	case <-req.CloseNotify:
		code = srv_err.ErrClientClosed
		err = fmt.Errorf("Client already closed.")
		return nil, code, err
	case result := <-ResultChan:
		code = result.Code
		if code != srv_err.SeckillSucc {
			return data, code, srv_err.GetErrMsg(code)
		}

		log.Printf("{User: %v} SecKill success.", req.UserId)
		data["product_id"] = result.ProductId
		data["token"] = result.Token
		data["user_id"] = result.UserId
		return data, code, nil
	}
}

func NewSecRequest() *model.SecRequest {
	secRequest := &model.SecRequest{
		ResultChan: make(chan *model.SecResult, 1),
	}
	return secRequest
}

type ServiceMiddleware func(Service) Service
