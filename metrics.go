package main

import (
	"os"
	"encoding/csv"
	"log"
	"strconv"

)

type MessageGlobalMetrics struct {
	fileName string
	numberSend int	
	errorSend int
	numberReceived int
	errorReceived int
}

func InitMessageMetrics(nodeRole string, nick string) (*MessageGlobalMetrics) {


	m := &MessageGlobalMetrics{
		fileName:			nodeRole + "-" + nick + "-MessageGlobal.csv",
		numberSend: 		0,
		errorSend:			0,
		numberReceived: 	0,
		errorReceived:		0,
	}
	return m
}

func (m *MessageGlobalMetrics) AddSend() {
	m.numberSend += 1
}
func (m *MessageGlobalMetrics) AddReceived() {
	m.numberReceived += 1
}

func (m *MessageGlobalMetrics) AddErrorSend() {
	m.errorSend += 1
}

func (m *MessageGlobalMetrics) AddErrorReceived() {
	m.errorReceived += 1
}

func (m *MessageGlobalMetrics) WriteMessageGlobalCSV(cpuLoad int) {

	file, err := os.Create(m.fileName)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	data := [][]string{
		{"# Message send", "# Send error", "# Message received", "# Received error", "Average cpu load"},
		{strconv.Itoa(m.numberSend), strconv.Itoa(m.errorSend), strconv.Itoa(m.numberReceived), strconv.Itoa(m.errorReceived), strconv.Itoa(cpuLoad)},
	}
	err = writer.WriteAll(data)
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing CSV writer:", err)
	}
}