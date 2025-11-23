package login

import "time"

type Session struct {
	Username   string `json:"username"`
	SessionId  string `json:"sessionId"`
	LoginTime  time.Time
	ExpireTime time.Time
}

type LoginRequest struct {
	CsrfToken string `json:"csrf_token" form:"csrf_token"`
	Login     string `json:"login" form:"login"`
	Password  string `json:"password" form:"password"`
	Redirect  string `json:"redirect" form:"redirect"`
}
