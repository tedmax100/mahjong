package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	logFile     *os.File
	gameLogger  *log.Logger
	mu          sync.Mutex
	logFilePath = "game.log"
)

// Init 初始化 logger，設定同時輸出到檔案和 stdout
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("無法開啟 log 檔案: %v", err)
	}

	// 設定標準 log 同時輸出到 stdout 和檔案
	multiWriter := &multiWriter{writers: []writerFunc{
		func(p []byte) (n int, err error) { return os.Stdout.Write(p) },
		func(p []byte) (n int, err error) { return logFile.Write(p) },
	}}

	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime)

	gameLogger = log.New(multiWriter, "", log.Ldate|log.Ltime)

	return nil
}

// ClearLog 清空 log 檔案（用於新局開始時）
func ClearLog() error {
	mu.Lock()
	defer mu.Unlock()

	// 關閉現有檔案
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("無法關閉 log 檔案: %v", err)
		}
	}

	// 以清空模式重新開啟檔案
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("無法清空 log 檔案: %v", err)
	}

	// 重新設定 writer
	multiWriter := &multiWriter{writers: []writerFunc{
		func(p []byte) (n int, err error) { return os.Stdout.Write(p) },
		func(p []byte) (n int, err error) { return logFile.Write(p) },
	}}

	log.SetOutput(multiWriter)
	gameLogger = log.New(multiWriter, "", log.Ldate|log.Ltime)

	// 寫入新局開始標記
	log.Printf("========== 新局開始 %s ==========", time.Now().Format("2006/01/02 15:04:05"))

	return nil
}

// Close 關閉 log 檔案
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Printf("關閉 log 檔案失敗: %v", err)
		}
		logFile = nil
	}
}

// multiWriter 同時寫入多個目標
type writerFunc func(p []byte) (n int, err error)

type multiWriter struct {
	writers []writerFunc
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}
