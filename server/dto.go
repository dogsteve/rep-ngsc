package server

type CronnJobConfig struct {
	DailyMorningCron string `json:"dailyMorningCron"`
	DailyEveningCron string `json:"dailyEveningCron"`
}
