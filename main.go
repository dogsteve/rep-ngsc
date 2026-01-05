package main

import (
	"encoding/json"
	"go-ngsc-erp/erp/app"
	"go-ngsc-erp/internal/elog"
	"go-ngsc-erp/server"
	"os"
)

func main() {
	// Initialize structured logger. Use LOG_LEVEL env var, default to "info".
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	_ = elog.Init(logLevel, "go-ngsc-erp")

	rawJson := `[
			{
				"username": "duydv@ngs.com.vn",
				"password": "Dangmai@123",
				"userId": 6100,
				"argId": 10246
			},
			{
				"username": "luyendv@ngs.com.vn",
				"password": "Haiphong1234@",
				"userId": 5231,
				"argId": 6303
			},
			{
				"username": "ngoc.truong@ngs.com.vn",
				"password": "TruongThanhTung10@",
				"userId": 4644,
				"argId": 5319
			},
			{
				"username": "hangltk@ngs.com.vn",
				"password": "1q2w3e$R98",
				"userId": 4592,
				"argId": 5235
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
	var users []app.UserCredentials
	err := json.Unmarshal([]byte(rawJson), &users)
	if err != nil {
		elog.Fatal("Failed to parse JSON: ", elog.F("err", err))
	}

	// 3. Đưa dữ liệu từ slice vào USER_STORE
	for _, u := range users {
		app.USER_STORE.Store(u.Username, u)
		elog.Info("Added user"+u.Username, elog.F("user", u))
	}

	go app.WaitForWritingLog()
	app.RunJob()
	server.StartServer()
}
