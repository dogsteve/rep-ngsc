package login

import (
	"fmt"
	"go-ngsc-erp/erp"
	"net/http"
	"strings"
	"time"

	"resty.dev/v3"
)

var LOGIN_SESSION = make(map[string]Session)

func addLoginSession(username string, sessionId string, expireTime time.Time) {
	loginSession := Session{
		Username:   username,
		SessionId:  sessionId,
		LoginTime:  time.Now(),
		ExpireTime: expireTime,
	}
	LOGIN_SESSION[username] = loginSession
}

func DoLogin(username, password string) error {
	restyClient := resty.New()
	defer func(restyClient *resty.Client) {
		err := restyClient.Close()
		if err != nil {
			fmt.Printf("%f", err)
		}
	}(restyClient)

	loginUrl := erp.ROOT_NGSC_URL + erp.LOGIN_PREFIX_URL
	fmt.Printf("login url %s \n", loginUrl)

	getResp, err := restyClient.R().Get(loginUrl)
	if err != nil {
		return err
	}
	if getResp.StatusCode() != 200 {
		fmt.Printf("Code is not 200: httpCode %d %s \n", getResp.StatusCode(), getResp.String())
		return fmt.Errorf("code is not 200: httpCode %d", getResp.StatusCode())
	}
	htmlBody := getResp.String()
	csrfToken, err := erp.FindByRegex(`csrf_token: *"([^"]+)"`, htmlBody)
	if err != nil {
		return err
	}
	csrfToken = strings.Replace(strings.Replace(csrfToken, "\"", "", -1), "csrf_token: ", "", -1)
	fmt.Println("csrf token " + csrfToken)

	sessionIdCookie, err := erp.FindFromCookie("session_id", getResp.Cookies())
	if err != nil {
		return err
	}
	fmt.Println("sessionIdCookie " + sessionIdCookie.Value)
	sessionId := sessionIdCookie.Value

	postResp, err := restyClient.R().
		SetCookies(CreateLoginCookies(sessionId)).
		SetFormData(map[string]string{
			"csrf_token": csrfToken,
			"login":      username,
			"password":   password,
		}).
		SetContentType("application/x-www-form-urlencoded").
		Post(loginUrl)
	if err != nil {
		return err
	}

	loginPostStt := postResp.StatusCode()
	if loginPostStt != 200 && loginPostStt != 303 && loginPostStt != 302 {
		fmt.Printf("Code is not valid: httpCode %d %s \n", loginPostStt, postResp.String())
		return fmt.Errorf("code is not 200: httpCode %d", loginPostStt)
	}

	sessionIdCookie, err = erp.FindFromCookie("session_id", postResp.Cookies())
	if err != nil {
		return err
	}
	sessionId = sessionIdCookie.Value
	expireTime := sessionIdCookie.Expires
	fmt.Printf("new sessionId: %s expire time %v \n", sessionId, expireTime)

	addLoginSession(username, sessionId, expireTime)

	return nil
}

func CreateLoginCookies(sessionID string) []*http.Cookie {
	cookies := []*http.Cookie{
		{Name: "cids", Value: "1"},
		{Name: "session_id", Value: sessionID},
		{Name: "frontend_lang", Value: "vi_VN"},
		{Name: "tz", Value: "Asia/Saigon"},
	}
	return cookies
}
