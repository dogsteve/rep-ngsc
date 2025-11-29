package main

import (
	"go-ngsc-erp/erp/app"
	"go-ngsc-erp/server"
)

func main() {
	go app.WaitForWritingLog()
	app.RunJob()
	server.StartServer()
}
