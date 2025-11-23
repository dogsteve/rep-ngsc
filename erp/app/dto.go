package app

import "time"

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserId   int    `json:"userId"`
	ArgId    int    `json:"argId"`
}

type CsvAttendanceLog struct {
	Username    string    `json:"username"`
	Action      string    `json:"action"`
	ActionTime  time.Time `json:"actionTime"`
	ErrorDetail string    `json:"errorDetail"`
	Status      string    `json:"status"`
}
