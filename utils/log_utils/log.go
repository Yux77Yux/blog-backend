package log_utils

import (
	"log"
	"os"
)

var (
	Logger  *log.Logger
	logFile *os.File
)

func init() {
	dir, _ := os.Getwd()
	var err error
	logFile, err = os.OpenFile("./log/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(dir)
		log.Fatalf("error opening log file: %v", err)
	}

	Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}
