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

var DailyMorningCron = "0 0 8 * * *"

//var DailyMorningCron = "0 * * * * *"

var DailyEveningCron = "0 45 17 * * *"

//var DailyEveningCron = "0 15 * * * *"

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

func (j *OneTimeJob) Run() {
	defer func() {
		fmt.Printf("[%s] 🗑️ Xóa Job Entry ID %d cho user %s\n", time.Now().Format("15:04:05"), j.ID, j.Username)
		j.Cron.Remove(j.ID)
	}()

	fmt.Printf("[%s] Start %s process user %v at %v\n", time.Now().Format("15:04:05"), j.ActionType, j.Username, time.Now())
	DoAction(j.ActionType, j.Credentials)
}

func RunJob() {
	c := cron.New(cron.WithSeconds())

	go WaitForWritingLog()

	_, err := c.AddFunc(DailyMorningCron, func() {
		currentTime := time.Now()
		fmt.Printf("\n--- [%s] Start morning routine ---\n", currentTime.Format("15:04:05"))
		USER_STORE.Range(func(key, value interface{}) bool {
			addTime := time.Duration(generateRandomInt(1, 20)) * time.Minute
			newTime := currentTime.Add(addTime)

			newCronn := createSpecificCronStringFromTime(newTime)

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
