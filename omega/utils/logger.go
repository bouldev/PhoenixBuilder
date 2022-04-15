package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"phoenixbuilder/omega/defines"
)

type MultipleLogger struct {
	Loggers []defines.LineDst
}

func (bl *MultipleLogger) Write(line string) {
	for _, logger := range bl.Loggers {
		logger.Write(line)
	}
}

type LogLineWrapper struct {
	log     *log.Logger
	closeFn func() error
}

func (w *LogLineWrapper) Write(data string) {
	w.log.Print(data)
}

func (w *LogLineWrapper) Close() error {
	return w.closeFn()
}

func NewFileNormalLogger(fileName string) *LogLineWrapper {
	if fileName == "omega_storage/logs/ChatLogger" {
		fmt.Println(fileName)
	}

	logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil && os.IsNotExist(err) {
		panic(fmt.Sprintf("Logger: cannot create or append file for logger %v (%v)", fileName, err))
	}
	log_ := log.New(logFile, "", log.Ldate|log.Ltime)
	log_.SetFlags(log.Ldate | log.Ltime)

	return &LogLineWrapper{
		log:     log_,
		closeFn: func() error { return logFile.Close() },
	}
}

func NewIONormalLogger(w io.Writer) *LogLineWrapper {
	log_ := log.New(w, "", log.Ldate|log.Ltime)
	log_.SetFlags(log.Ldate | log.Ltime)
	return &LogLineWrapper{
		log:     log_,
		closeFn: func() error { return nil },
	}
}
