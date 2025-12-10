package logger

import (
	"os"
	"strings"
	"testing"
)

func TestLoggerLifecycle(t *testing.T) {
	// Setup temporary log file
	tmpFile, err := os.CreateTemp("", "test_game_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Override logFilePath
	originalLogFilePath := logFilePath
	logFilePath = tmpPath
	defer func() { logFilePath = originalLogFilePath }()

	// Test Init
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write something using standard log (which is redirected)
	testMsg := "Test log message"
	gameLogger.Println(testMsg)

	// Test ClearLog
	if err := ClearLog(); err != nil {
		t.Fatalf("ClearLog failed: %v", err)
	}

	// Verify file content after clear (should have "新局開始")
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "新局開始") {
		t.Errorf("Log file should contain '新局開始' after ClearLog, got: %s", string(content))
	}

	// Test Close
	Close()

	// Verify we can init again
	if err := Init(); err != nil {
		t.Fatalf("Re-Init failed: %v", err)
	}
	Close()
}
