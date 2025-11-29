package server

import (
	"fmt"
	"go-ngsc-erp/erp/app"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const timeLayout = "2006-01-02 15:04:05"

func StartServer() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		var userCredentials app.UserCredentials
		err := render.Decode(r, &userCredentials)
		if err != nil {
			// Trả về lỗi 400 Bad Request nếu JSON không hợp lệ
			http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
			return
		}
		app.USER_STORE.Store(userCredentials.Username, userCredentials)
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		userResponse := make([]app.UserCredentials, 0)
		app.USER_STORE.Range(func(key, value interface{}) bool {
			userResponse = append(userResponse, value.(app.UserCredentials))
			return true
		})
		render.JSON(w, r, userResponse)
	})

	r.Post("/cron", func(w http.ResponseWriter, r *http.Request) {
		var cron CronnJobConfig
		err := render.Decode(r, &cron)
		if err != nil {
			// Trả về lỗi 400 Bad Request nếu JSON không hợp lệ
			http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
			return
		}
		if cron.DailyMorningCron == "" {
			app.DailyMorningCron = cron.DailyMorningCron
		}
		if cron.DailyEveningCron != "" {
			app.DailyEveningCron = cron.DailyEveningCron
		}

	})

	r.Get("/statistic", func(w http.ResponseWriter, r *http.Request) {
		result, err := app.ReadCSVAndMap()
		if err != nil {
			// Trả về lỗi 400 Bad Request nếu JSON không hợp lệ
			http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
			return
		}
		render.JSON(w, r, result)
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
