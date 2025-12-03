package app

import (
	"fmt"
	"go-ngsc-erp/erp/attendance"
	"go-ngsc-erp/erp/login"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const customTimeFormat = "2006-01-02T15:04"
const CsvPath = "./attendance.csv"
const TimeLayout = time.RFC3339

var DailyMorningCron = "0 0 8 * * *"

//var DailyMorningCron = "0 * * * * *"

var DailyEveningCron = "0 45 17 * * *"

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
		fmt.Printf("Error when do login with user %v %v\n", credentials, err)
		csvLog.ErrorDetail = "LOGIN ERROR: " + err.Error()
		csvLog.Status = "ATTENDANCE FAILED"
		CsvWriterChan <- csvLog
		return
	}

	err = attendance.DoAttendance(credentials.Username, credentials.UserId, credentials.ArgId)
	if err != nil {
		fmt.Printf("Error when do attendance with user %v %v\n", credentials, err)
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
		fmt.Printf("lỗi phân tích chuỗi cron '%s': %w \n", cronString, err)
		return
	}

	// 2. Lấy thời điểm hiện tại.
	// Phương thức Next sẽ tính thời điểm chạy tiếp theo SAU thời điểm này.
	now := time.Now()

	// 3. Tính toán thời gian chạy tiếp theo.
	nextRunTime := schedule.Next(now)

	// 4. Trả về thời gian chạy tiếp theo và không có lỗi.
	fmt.Printf("nextRunTime %v \n", nextRunTime)
}

func (j *OneTimeJob) Run() {
	defer func() {
		fmt.Printf("[%s] 🗑️ Xóa Job Entry ID %d cho user %s\n", time.Now().Format("15:04:05"), j.ID, j.Username)
		j.Cron.Remove(j.ID)
	}()

	fmt.Printf("[%s] Start %s process user %v at %v\n", time.Now().Format("15:04:05"), j.ActionType, j.Username, time.Now())
	DoAction(j.ActionType, j.Credentials)
}

func RunJob() {

	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatal(err)
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	c := cron.New(cron.WithLocation(loc), cron.WithParser(parser))

	go WaitForWritingLog()

	printNextRunTime(DailyMorningCron)
	printNextRunTime(DailyEveningCron)

	_, err = c.AddFunc(DailyMorningCron, func() {
		currentTime := time.Now()
		fmt.Printf("\n--- [%s] Start morning routine ---\n", currentTime.Format("15:04:05"))
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
				fmt.Printf("Error adding CHECKIN Job for %s: %v\n", userCredential.Username, err)
			} else {
				oneTimeJob.ID = entryID
				fmt.Printf("   -> Scheduled CHECKIN for %s at %s (EntryID: %d)\n", userCredential.Username, newCronn, entryID)
			}
			return true
		})
		fmt.Println("   --- End morning routine ---")
		printNextRunTime(DailyMorningCron)
	})
	if err != nil {
		fmt.Printf("Error adding Morning Routine Job: %v\n", err)
	}

	_, err = c.AddFunc(DailyEveningCron, func() {
		currentTime := time.Now()
		fmt.Printf("\n--- [%s] Start evening routine ---\n", currentTime.Format("15:04:05"))
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
				fmt.Printf("Error adding CHECKOUT Job for %s: %v\n", userCredential.Username, err)
			} else {
				oneTimeJob.ID = entryID
				fmt.Printf("   -> Scheduled CHECKOUT for %s at %s (EntryID: %d)\n", userCredential.Username, newCronn, entryID)
			}
			return true
		})
		fmt.Println("   --- End evening routine ---")
		printNextRunTime(DailyEveningCron)
	})
	if err != nil {
		fmt.Printf("Error adding Evening Routine Job: %v\n", err)
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
