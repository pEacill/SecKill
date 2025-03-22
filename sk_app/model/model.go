package model

type SecRequest struct {
	ProductId     int             `json:"product_id"`
	Source        string          `json:"source"`
	AuthCode      string          `json:"auth_code"`
	SecTime       int64           `json:"sec_time"`
	Nance         string          `json:"nance"`
	UserId        int             `json:"user_id"`
	UserAuthSign  string          `json:"user_auth_sign"`
	AccessTime    int64           `json:"access_time"`
	ClientAddr    string          `json:"client_addr"`
	ClientRefence string          `json:"client_refence"`
	CloseNotify   <-chan bool     `json:"-"`
	ResultChan    chan *SecResult `json:"-"`
}

type SecResult struct {
	ProductId int    `json:"product_id"`
	UserId    int    `json:"user_id"`
	Token     string `json:"token"`      //Token
	TokenTime int64  `json:"token_time"` //Token gen time
	Code      int    `json:"code"`       //state code
}
