package app

import (
	"fmt"
	"go-ngsc-erp/erp/attendance"
	"go-ngsc-erp/erp/login"
	"log"
	"math/rand"
	"sync"
	"time"

	"go-ngsc-erp/internal/elog"

	"github.com/robfig/cron/v3"
)

const customTimeFormat = "2006-01-02T15:04"
const CsvPath = "./attendance.csv"
const TimeLayout = time.RFC3339

var DailyMorningCron = "0 0 8 * * 1-5"

//var DailyMorningCron = "0 * * * * *"

var DailyEveningCron = "0 45 17 * * 1-5"

//var DailyEveningCron = "0 * * * * *"

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
		elog.Error("Error when do login", elog.Fields{"user": credentials.Username, "err": err})
		csvLog.ErrorDetail = "LOGIN ERROR: " + err.Error()
		csvLog.Status = "ATTENDANCE FAILED"
		CsvWriterChan <- csvLog
		return
	}
	time.Sleep(5 * time.Second) // Thời gian chờ giữa login và attendance
	err = attendance.DoAttendance(credentials.Username, credentials.UserId, credentials.ArgId)
	if err != nil {
		elog.Error("Error when do attendance", elog.Fields{"user": credentials.Username, "err": err})
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
		elog.Error("Error when create csv writer", elog.F("err", err))
		return
	}
	for logItem := range CsvWriterChan {
		writeErr := csvWriter.WriteRow([]string{
			logItem.Username,
			logItem.Action,
			logItem.ActionTime.Format(time.RFC3339),
			logItem.ErrorDetail,
			logItem.Status,
		})
		if writeErr != nil {
			elog.Warn("Error when write log", elog.F("err", writeErr))
		}
	}
	close(CsvWriterChan)
}

type OneTimeJob struct {
	Cron        *cron.Cron // Tham chiếu đến scheduler để gọi Remove
	ID          cron.EntryID
	Username    string
	Credentials UserCredentials
	ActionType  string
}

func printNextRunTime(cronString string) {
	// 1. Phân tích chuỗi cron string thành một Schedule.
	// Chúng ta sử dụng cron.ParseStandard() để phân tích cú pháp 5 trường (phút giờ ngày tháng thứ).
	// Nếu chuỗi của bạn có giây (6 trường), bạn cần dùng:
	// parser := cron.NewParser(cron.StandardSecondsSpec)
	// schedule, err := parser.Parse(cronString)
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronString)

	if err != nil {
		// Trả về lỗi nếu chuỗi cron không hợp lệ (ví dụ: quá ít hoặc quá nhiều trường).
		elog.Warn("invalid cron string", elog.Fields{"cron": cronString, "err": err})
		return
	}

	// 2. Lấy thời điểm hiện tại.
	// Phương thức Next sẽ tính thời điểm chạy tiếp theo SAU thời điểm này.
	now := time.Now()

	// 3. Tính toán thời gian chạy tiếp theo.
	nextRunTime := schedule.Next(now)

	// 4. Trả về thời gian chạy tiếp theo và không có lỗi.
	elog.Info("next run time", elog.Fields{"cron": cronString, "next": nextRunTime.Format(time.RFC3339)})
}

func (j *OneTimeJob) Run() {
	defer func() {
		elog.Info("removing job entry", elog.Fields{"entry_id": j.ID, "user": j.Username})
		j.Cron.Remove(j.ID)
	}()

	elog.Info("start job", elog.Fields{"action": j.ActionType, "user": j.Username})
	DoAction(j.ActionType, j.Credentials)
}

func RunJob() {

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		elog.Fatal("could not load timezone", elog.F("err", err))
		log.Fatal(err)
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	c := cron.New(cron.WithLocation(loc), cron.WithParser(parser))

	printNextRunTime(DailyMorningCron)
	printNextRunTime(DailyEveningCron)

	_, err = c.AddFunc(DailyMorningCron, func() {
		currentTime := time.Now()
		elog.Info("start morning routine", elog.F("ts", currentTime.Format("15:04:05")))
		USER_STORE.Range(func(key, value interface{}) bool {
			addTime := time.Duration(generateRandomInt(1, 20)) * time.Minute
			newTime := currentTime.Add(addTime)

			newCronn := createSpecificCronStringFromTime(newTime)
			printNextRunTime(newCronn)

			userCredential := value.(UserCredentials)

			oneTimeJob := &OneTimeJob{
				Cron:        c,
				Username:    userCredential.Username,
				Credentials: userCredential,
				ActionType:  "CHECKIN",
			}

			entryID, err := c.AddJob(newCronn, oneTimeJob)
			if err != nil {
				elog.Error("Error adding CHECKIN Job", elog.Fields{"user": userCredential.Username, "err": err})
			} else {
				oneTimeJob.ID = entryID
				elog.Info("scheduled checkin", elog.Fields{"user": userCredential.Username, "cron": newCronn, "entry_id": entryID})
			}
			return true
		})
		printNextRunTime(DailyMorningCron)
	})
	if err != nil {
		elog.Error("Error adding Morning Routine Job", elog.F("err", err))
	}

	_, err = c.AddFunc(DailyEveningCron, func() {
		currentTime := time.Now()
		elog.Info("start evening routine", elog.F("ts", currentTime.Format("15:04:05")))
		USER_STORE.Range(func(key, value interface{}) bool {
			addTime := time.Duration(generateRandomInt(1, 40)) * time.Minute
			newTime := currentTime.Add(addTime)
			newCronn := createSpecificCronStringFromTime(newTime)
			printNextRunTime(newCronn)

			userCredential := value.(UserCredentials)

			oneTimeJob := &OneTimeJob{
				Cron:        c,
				Username:    userCredential.Username,
				Credentials: userCredential,
				ActionType:  "CHECKOUT",
			}

			entryID, err := c.AddJob(newCronn, oneTimeJob)
			if err != nil {
				elog.Error("Error adding CHECKOUT Job", elog.Fields{"user": userCredential.Username, "err": err})
			} else {
				oneTimeJob.ID = entryID
				elog.Info("scheduled checkout", elog.Fields{"user": userCredential.Username, "cron": newCronn, "entry_id": entryID})
			}
			return true
		})
		printNextRunTime(DailyEveningCron)
	})
	if err != nil {
		elog.Error("Error adding Evening Routine Job", elog.F("err", err))
	}

	c.Start()
}

func createSpecificCronStringFromTime(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d %d *",
		t.Second(),
		t.Minute(), // 0-59
		t.Hour(),   // 0-23
		t.Day(),    // 1-31
		t.Month(),  // 1-12
	)
}

func generateRandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}
