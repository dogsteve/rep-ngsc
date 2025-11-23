package app

import (
	"fmt"
	"go-ngsc-erp/erp/attendance"
	"go-ngsc-erp/erp/login"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const customTimeFormat = "2006-01-02T15:04"
const CsvPath = "./attendance.csv"
const TimeLayout = time.RFC3339

var DailyMorningCron = "0 8 * * *"
var DailyEveningCron = "45 17 * * *"

var USER_STORE = sync.Map{}

var CsvWriterChan = make(chan CsvAttendanceLog)

func DoAction(action string, credentials UserCredentials) {
	csvLog := CsvAttendanceLog{
		Username:    credentials.Username,
		Action:      action,
		ActionTime:  time.Now(),
		ErrorDetail: "",
		Status:      "NOT_PROCESSED",
	}
	err := login.DoLogin(credentials.Username, credentials.Password)
	if err != nil {
		csvLog.ErrorDetail = "LOGIN ERROR: " + err.Error()
		csvLog.Status = "ATTENDANCE FAILED"
		CsvWriterChan <- csvLog
		return
	}

	err = attendance.DoAttendance(credentials.Username, credentials.UserId, credentials.ArgId)
	if err != nil {
		csvLog.ErrorDetail = "ATTENDANCE ERROR: " + err.Error()
		csvLog.Status = "ATTENDANCE FAILED"
		CsvWriterChan <- csvLog
		return
	}

	csvLog.Status = "ATTENDANCE SUCCESS"
	CsvWriterChan <- csvLog
}

func WaitForWritingLog() {
	csvWriter, err := NewSyncCSVWriter(CsvPath, []string{"Username", "Action", "ActionTime", "ErrorDetail", "Status"})
	if err != nil {
		fmt.Println("Error when create csv writer " + err.Error())
		return
	}
	for log := range CsvWriterChan {
		writeErr := csvWriter.WriteRow([]string{
			log.Username,
			log.Action,
			log.ActionTime.Format(time.RFC3339),
			log.ErrorDetail,
			log.Status,
		})
		if writeErr != nil {
			fmt.Println("Error when write log " + writeErr.Error())
		}
	}
	close(CsvWriterChan)
}

func RunJob() {
	c := cron.New()
	_, err := c.AddFunc(DailyMorningCron, func() {
		currentTime := time.Now()
		USER_STORE.Range(func(key, value interface{}) bool {
			newTime := currentTime.Add(time.Duration(generateRandomInt(0, 40)) * time.Minute)
			newCronn := createSpecificCronStringFromTime(newTime)
			fmt.Println(newCronn)
			_, err := c.AddFunc(newCronn, func() {
				userCredential := value.(UserCredentials)
				DoAction("CHECKIN", userCredential)
			})
			if err != nil {
				fmt.Println(err)
			}
			return true
		})
	})
	if err != nil {
		fmt.Printf("Error adding Job A: %v\n", err)
	}
	_, err = c.AddFunc(DailyEveningCron, func() {
		currentTime := time.Now()
		USER_STORE.Range(func(key, value interface{}) bool {
			newTime := currentTime.Add(time.Duration(generateRandomInt(10, 60)) * time.Minute)
			newCronn := createSpecificCronStringFromTime(newTime)
			fmt.Println(newCronn)
			_, err := c.AddFunc(newCronn, func() {
				userCredential := value.(UserCredentials)
				DoAction("CHECKOUT", userCredential)
			})
			if err != nil {
				fmt.Println(err)
			}
			return true
		})
	})
	if err != nil {
		fmt.Printf("Error adding Job A: %v\n", err)
	}
	c.Run()
}

func createSpecificCronStringFromTime(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d *",
		t.Minute(), // 0-59
		t.Hour(),   // 0-23
		t.Day(),    // 1-31
		t.Month(),  // 1-12
	)
}

func generateRandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}
