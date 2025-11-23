package erp

import (
	"fmt"
	"net/http"
	"regexp"
)

const ROOT_NGSC_URL = "https://erp-ngsc.com.vn/web"
const LOGIN_PREFIX_URL = "/login"
const ATTENDANCE_PREFIX_URL = "/dataset/call_kw/hr.employee/attendance_manual"
const REQUEST_COOKIE_HEADER = "cids=1; frontend_lang=vi_VN; tz=Asia/Saigon; session_id=%s"

func FindByRegex(regexPattern, sourceVal string) (string, error) {
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(sourceVal)
	if len(matches) > 0 {
		return matches[0], nil
	} else {
		return "", fmt.Errorf("cannot find regex")
	}
}

func FindFromCookie(keyname string, cookies []*http.Cookie) (*http.Cookie, error) {
	for _, cookie := range cookies {
		if cookie != nil {
			if cookie.Name == keyname {
				return cookie, nil
			}
		}
	}
	return nil, fmt.Errorf("cookie not found")
}
