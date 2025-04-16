package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// UpgradeToWebSocket 升级HTTP连接到WebSocket
func UpgradeToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ReadWSMessage 读取并解析WebSocket消息
func ReadWSMessage[T any](conn *websocket.Conn) (Message[T], error) {
	var msg Message[T]
	_, data, err := conn.ReadMessage()
	if err != nil {
		return msg, err
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

// SendWSMessage 发送结构化WebSocket消息
func SendWSMessage(conn *websocket.Conn, code int, message string, data interface{}) error {
	msg := Message[interface{}]{
		Code:    code,
		Message: message,
		Data:    data,
	}
	return conn.WriteJSON(msg)
}

// SendError 发送错误消息
func SendError(conn *websocket.Conn, code int, errorMsg string) {
	_ = SendWSMessage(conn, code, errorMsg, nil)
}
