package login

import (
	"fmt"
	"go-ngsc-erp/erp"
	"net/http"
	"os"
	"strings"
	"time"

	"go-ngsc-erp/internal/elog"

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
	elog.Info("Added login session", elog.F("user", username))
}

func DoLogin(username, password string) error {
	currentTime := time.Now()
	elog.Info("Start login process", elog.Fields{"user": username, "ts": currentTime.Format("15:04:05")})
	restyClient := resty.New()
	defer func(restyClient *resty.Client) {
		err := restyClient.Close()
		if err != nil {
			elog.Error("error closing resty client", elog.F("err", err))
		}
	}(restyClient)

	loginUrl := erp.ROOT_NGSC_URL + erp.LOGIN_PREFIX_URL
	elog.Debug("login url", elog.F("url", loginUrl))

	getResp, err := restyClient.R().Get(loginUrl)
	if err != nil {
		elog.Error("error fetching login page", elog.Fields{"err": err, "user": username})
		return err
	}
	if getResp.StatusCode() != 200 {
		elog.Warn("login page returned non-200", elog.Fields{"code": getResp.StatusCode(), "body": getResp.String()})
		return fmt.Errorf("code is not 200: httpCode %d", getResp.StatusCode())
	}
	htmlBody := getResp.String()
	csrfToken, err := erp.FindByRegex(`csrf_token: *"([^\"]+)"`, htmlBody)
	if err != nil {
		elog.Error("csrf token not found", elog.F("err", err))
		return err
	}
	csrfToken = strings.Replace(strings.Replace(csrfToken, "\"", "", -1), "csrf_token: ", "", -1)
	elog.Debug("csrf token parsed", elog.F("token_len", len(csrfToken)))

	sessionIdCookie, err := erp.FindFromCookie("session_id", getResp.Cookies())
	if err != nil {
		elog.Error("session cookie not found", elog.F("err", err))
		return err
	}
	elog.Debug("initial session id found", elog.F("cookie", sessionIdCookie.Value))
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
		elog.Error("error posting login form", elog.F("err", err))
		return err
	}

	loginPostStt := postResp.StatusCode()
	if (loginPostStt != 200 && loginPostStt != 303 && loginPostStt != 302) || strings.Contains(postResp.String(), "Login") {
		elog.Warn("Login not valid", elog.Fields{"code": loginPostStt, "body": postResp.String(), "user": username})
		return fmt.Errorf("Login not valid: httpCode %d", loginPostStt)
	}

	sessionIdCookie, err = erp.FindFromCookie("session_id", postResp.Cookies())
	if err != nil {
		elog.Error("session cookie after login not found", elog.F("err", err))
		return err
	}
	sessionId = sessionIdCookie.Value
	expireTime := sessionIdCookie.Expires
	elog.Info("new session", elog.Fields{"session_id": sessionId, "expire": expireTime.Format(time.RFC3339), "user": username})

	addLoginSession(username, sessionId, expireTime)
	elog.Info("Finish login process", elog.F("user", username))

	// if running under short-lived CLI tests we may want to flush
	if os.Getenv("CI") == "true" {
		// noop for now
	}

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
