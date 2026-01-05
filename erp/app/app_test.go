package app

import (
	"encoding/json"
	"go-ngsc-erp/erp/attendance"
	"go-ngsc-erp/erp/login"
	"testing"
)

func TestLoginAndAttendance(t *testing.T) {
	// 1. Dữ liệu JSON đầu vào
	rawJson := `[
		{
			"username": "duydv@ngs.com.vn",
			"password": "Dangmai@123",
			"userId": 6100,
			"argId": 10246
		},
		{
			"username": "minhnq1@ngs.com.vn",
			"password": "Zen284366990!",
			"userId": 6151,
			"argId": 10335
		},
		{
			"username": "cuongtv@ngs.com.vn",
			"password": "554433aA@",
			"userId": 6070,
			"argId": 10183
		}
	]`

	// 2. Parse JSON vào một slice tạm thời
	var users []UserCredentials
	err := json.Unmarshal([]byte(rawJson), &users)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// 3. Đưa dữ liệu từ slice vào USER_STORE
	for _, u := range users {
		USER_STORE.Store(u.Username, u)
	}

	// 4. Lặp qua USER_STORE và thực hiện các hành động
	USER_STORE.Range(func(key, value interface{}) bool {
		credentials := value.(UserCredentials)

		t.Logf("Processing user: %s", credentials.Username)

		// Step 1: Login
		t.Log("  -> Attempting Login...")
		err := login.DoLogin(credentials.Username, credentials.Password)
		if err != nil {
			t.Errorf("  [FAILED] Login error for %s: %v", credentials.Username, err)
			return true // Tiếp tục sang user tiếp theo
		}
		t.Log("  [SUCCESS] Login successful")

		// Step 2: Attendance
		t.Log("  -> Attempting Attendance...")
		err = attendance.DoAttendance(credentials.Username, credentials.UserId, credentials.ArgId)
		if err != nil {
			t.Errorf("  [FAILED] Attendance error for %s: %v", credentials.Username, err)
			return true
		}
		t.Log("  [SUCCESS] Attendance successful")

		return true
	})
}
