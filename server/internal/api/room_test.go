package api

import (
	"encoding/json"
	"mahjong/internal/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateRoom(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	hub := websocket.NewHub()
	r.POST("/api/rooms", CreateRoom(hub))

	// Create request payload
	reqBody := `{"userId": "user123", "userName": "TestUser"}`
	req, _ := http.NewRequest("POST", "/api/rooms", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Assert response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success to be true")
	}

	roomID, ok := response["roomId"].(string)
	if !ok || roomID == "" {
		t.Error("Expected valid roomId in response")
	}
}

func TestGetRoom(t *testing.T) {
	// Setup Gin and Hub
	gin.SetMode(gin.TestMode)
	r := gin.New()
	hub := websocket.NewHub()
	
	// Pre-create a room
	room := hub.CreateRoom("123456")
	if room == nil {
		t.Fatal("Failed to create test room")
	}

	r.GET("/api/rooms/:roomId", GetRoom(hub))

	// Test Case 1: Existing Room
	req, _ := http.NewRequest("GET", "/api/rooms/123456", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for existing room, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["success"] != true {
		t.Error("Expected success=true for existing room")
	}

	// Test Case 2: Non-existing Room
	req2, _ := http.NewRequest("GET", "/api/rooms/999999", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existing room, got %d", w2.Code)
	}
}
