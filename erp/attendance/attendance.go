package attendance

import (
	"fmt"
	"go-ngsc-erp/erp"
	"go-ngsc-erp/erp/login"
	"math/rand"
	"time"

	"resty.dev/v3"
)

const (
	TIMEZONE_DEFAULT = "Asia/Saigon"
	FIXED_LATITUDE   = 21.051364 // Ví dụ: Hà Nội
	FIXED_LONGITUDE  = 105.799611
	EN_LOCATION_ID   = "2"
)

func BuildAttendanceJSON(userArgID int, userID int) DataJSON {
	// Khởi tạo seed cho hàm rand dựa trên thời gian hiện tại
	// CHÚ Ý: Trong môi trường production, nên sử dụng crypto/rand để có tính bảo mật cao hơn
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. Sinh ID ngẫu nhiên từ 1 đến 100
	requestID := rand.Intn(100) + 1

	// 2. Định nghĩa Context
	context := Context{
		Lang:              "vi_VN",
		TZ:                TIMEZONE_DEFAULT,
		UID:               userID,
		AllowedCompanyIDs: []int{1}, // Giả định company ID luôn là 1
		Latitude:          FIXED_LATITUDE,
		Longitude:         FIXED_LONGITUDE,
		EnLocationID:      EN_LOCATION_ID,
	}

	// 3. Định nghĩa Kwargs và Params
	params := Params{
		Args: []interface{}{
			[]int{userArgID}, // Tham số 1: Mảng chứa ID cần thao tác (ví dụ: employee ID)
			"hr_attendance.hr_attendance_action_my_attendances", // Tham số 2: Tên hành động
		},
		Model:  "hr.employee",
		Method: "attendance_manual",
		Kwargs: Kwargs{
			Context: context,
		},
	}

	// 4. Định nghĩa DataJSON cấp cao nhất
	dataJSON := DataJSON{
		ID:      requestID,
		JSONRPC: "2.0",
		Method:  "call",
		Params:  params,
	}

	return dataJSON
}

func DoAttendance(username string, userId, userArgId int) error {
	loginSession, ok := login.LOGIN_SESSION[username]
	if !ok {
		fmt.Println("login.LOGIN_SESSION ERROR")
		return fmt.Errorf("need login first " + username)
	}
	if loginSession.ExpireTime.Compare(time.Now()) < 0 {
		fmt.Println("login.LOGIN_SESSION ERROR")
		return fmt.Errorf("need login first " + username)
	}

	fmt.Printf("login.LOGIN_SESSION OK %v\n", loginSession)
	dataJSON := BuildAttendanceJSON(userArgId, userId)

	restyClient := resty.New()
	defer func(restyClient *resty.Client) {
		err := restyClient.Close()
		if err != nil {
			fmt.Printf("%f", err)
		}
	}(restyClient)

	attendanceUrl := erp.ROOT_NGSC_URL + erp.ATTENDANCE_PREFIX_URL
	postResp, err := restyClient.R().
		SetBody(dataJSON).
		SetCookies(login.CreateLoginCookies(loginSession.SessionId)).
		SetHeaders(map[string]string{
			"Accept":       "*/*",
			"Content-Type": "application/json",
			"Origin":       "https://erp-ngsc.com.vn",
			"Referer":      erp.ROOT_NGSC_URL,
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
		}).
		Post(attendanceUrl)

	if err != nil {
		return err
	}

	attendanceStt := postResp.StatusCode()
	if attendanceStt != 200 {
		fmt.Printf("Code is not valid: httpCode %d %s \n", attendanceStt, postResp.String())
		return fmt.Errorf("code is not 200: httpCode %d", attendanceStt)
	}

	fmt.Println("Attendance success")

	return nil
}
