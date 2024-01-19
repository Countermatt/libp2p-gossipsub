package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

type MessageType int

const (
	BuilderPublishRow MessageType = iota
	BuilderPublishColumn
	ValidatorReceiveRow
	ValidatorReceiveColumn
	BuilderPublishHeader
	ValidatorReceiveHeader
	ValidatorPublishRow
	ValidatorPublishColumn
	RegularReceiveRow
	RegularReceiveColumn
)

type LogEvent struct {
	Timestamp string `json:"timestamp"`
	EventType int    `json:"eventType"`
	BlockId   int    `json:"blockId"`
}

type LogEntry struct {
	Timestamp   string      `json:"timestamp"`
	SenderID    string      `json:"SenderID"`
	RowColumnID int         `json:"RowColumnID"`
	Topic       string      `json:"Topic"`
	MessageType MessageType `json:"MessageType"` // message type
}

type LogEntryHeader struct {
	Timestamp   string      `json:"timestamp"`
	SenderID    string      `json:"SenderID"`
	Topic       string      `json:"Topic"`
	Block       string      `json:"Block"`
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

func formatJSONLogHeaderSend(SenderID string, Topic string, Block int, messageType MessageType) string {
	// Custom log entry struct
	logEntry := LogEntryHeader{
		Timestamp:   time.Now().Format(time.RFC3339Nano),
		SenderID:    SenderID,
		Topic:       Topic,
		Block:       strconv.Itoa(Block),
		MessageType: BuilderPublishHeader,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}

func formatJSONLogEvent(eventType int, blockId int) string {
	// Custom log entry struct
	logEntry := LogEvent{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		EventType: eventType,
		BlockId:   blockId,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}
