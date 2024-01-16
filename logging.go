package main

import (
	"encoding/json"
	"log"
	"time"
)

type MessageType int

const (
	BuilderPublishRow MessageType = iota
	BuilderPublishColumn
)

type LogEntry struct {
	Timestamp   string      `json:"timestamp"`
	SenderID    string      `json:"SenderID"`
	RowColumnID int         `json:"RowColumnID"`
	Topic       string      `json:"Topic"`
	MessageType MessageType `json:"MessageType"` // message type
}

func formatJSONLogMessageSend(SenderID string, rowColumnID int, Topic string, messageType MessageType) string {
	// Custom log entry struct
	logEntry := LogEntry{
		Timestamp:   time.Now().Format(time.RFC3339Nano),
		SenderID:    SenderID,
		RowColumnID: rowColumnID,
		Topic:       Topic,
		MessageType: messageType,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}
