package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Category string

const (
	CatSystem Category = "SYSTEM"
	CatDB     Category = "DB"
	CatAuth   Category = "AUTH"
	CatHTTP   Category = "HTTP"
	CatMap    Category = "MAP"
	CatMedia  Category = "MEDIA"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	infoFile    *os.File
	errorFile   *os.File
	mu          sync.Mutex
)

func Init(logDir string) error {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	var err error
	infoFile, err = os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open app.log: %w", err)
	}

	errorFile, err = os.OpenFile(filepath.Join(logDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		_ = infoFile.Close()
		return fmt.Errorf("open error.log: %w", err)
	}

	infoLogger = log.New(io.MultiWriter(os.Stdout, infoFile), "", 0)
	errorLogger = log.New(io.MultiWriter(os.Stderr, errorFile), "", 0)

	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if infoFile != nil {
		_ = infoFile.Close()
	}
	if errorFile != nil {
		_ = errorFile.Close()
	}
}

func Info(cat Category, format string, args ...any) {
	write(false, "INFO", cat, format, args...)
}

func Warn(cat Category, format string, args ...any) {
	write(true, "WARN", cat, format, args...)
}

func Error(cat Category, format string, args ...any) {
	write(true, "ERROR", cat, format, args...)
}

func Fatal(cat Category, format string, args ...any) {
	write(true, "FATAL", cat, format, args...)
	Close()
	os.Exit(1)
}

func write(isErr bool, level string, cat Category, format string, args ...any) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [%s] [%s] %s", ts, level, cat, msg)

	mu.Lock()
	defer mu.Unlock()

	if isErr {
		if errorLogger != nil {
			errorLogger.Println(line)
		} else {
			log.Println(line)
		}
		return
	}

	if infoLogger != nil {
		infoLogger.Println(line)
	} else {
		log.Println(line)
	}
}
