package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const ChainBufSize = 1024
const NbSample = 75

type TopicSubItem struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

type Host struct {
	message        chan *Message
	ctx            context.Context
	ps             *pubsub.PubSub
	topicNames     []string
	topicsubList   []TopicSubItem
	self           peer.ID
	nick           string
	messageMetrics *MessageGlobalMetrics
}

func CreateHost(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, nodeType string, BlockSize int) (*Host, error) {

	//========== Subscribe nodes to topics ==========
	roomNameList := make([]string, 0)
	rand.Seed(time.Now().UnixNano())

	switch nodeType {

	//Subscribe builders to all row and column
	case "builder":
		for i := 0; i < BlockSize; i++ {
			roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(i))
			roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(i))
		}

	//Subscribe validators to 2 random row and 2 random column
	case "validator":
		column1 := rand.Intn(BlockSize)
		column2 := rand.Intn(BlockSize)
		row1 := rand.Intn(BlockSize)
		row2 := rand.Intn(BlockSize)
		roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(column1))
		roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(column2))
		roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(row1))
		roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(row2))

	//Non validator subscribe to 75 random samples
	default:

		sample := make([]int, 0)
		i := 0
		for i < NbSample {
			id := rand.Intn(BlockSize * BlockSize)
			if findElementInt(sample, id) == -1 {
				sample = append(sample, id)
				column := id % BlockSize
				row := (id - column) / BlockSize
				if findElementString(roomNameList, "builder:c"+strconv.Itoa(column)) == -1 {
					roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(column))
				}
				if findElementString(roomNameList, "builder:r"+strconv.Itoa(row)) == -1 {
					roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(row))
				}
				i += 1
			}
		}

	}
	mesMetrics := InitMessageMetrics(nodeType, nickname)
	h := &Host{
		ctx:            ctx,
		ps:             ps,
		topicNames:     roomNameList,
		topicsubList:   make([]TopicSubItem, 0),
		self:           selfID,
		nick:           nickname,
		message:        make(chan *Message, ChainBufSize),
		messageMetrics: mesMetrics,
	}

	for i := 0; i < len(roomNameList); i++ {
		h.AddSubTopic(roomNameList[i])
		go h.readLoop(roomNameList[i])
	}
	// start reading message from the subscription in a loop
	return h, nil
}

func (h *Host) AddSubTopic(roomName string) error {
	// join the pubsub topic
	topic, err := h.ps.Join(topicName(roomName))
	if err != nil {
		return err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	tsi := &TopicSubItem{
		topic: topic,
		sub:   sub,
	}

	h.topicsubList = append(h.topicsubList, *tsi)

	return nil
}

// Publish sends a message to the pubsub topic.
func (h *Host) Publish(topic string, colRow int, first int, block int, size int) error {
	m := CreateMessage(CreateParcel(colRow, block, size, first), topic, h.self, h.nick, first, block)
	msgBytes, err := json.Marshal(m)
	if err != nil {
		h.messageMetrics.AddErrorSend()
		return err
	}
	h.messageMetrics.AddSend()
	return h.topicsubList[findElementString(h.topicNames, topic)].topic.Publish(h.ctx, msgBytes)
}

/*
	func (h *Host) ListPeers() []peer.ID {
		return h.ps.ListPeers(topicName(h.roomName))
	}
*/
func (h *Host) readLoop(topic string) {
	for {
		msg, err := h.topicsubList[findElementString(h.topicNames, topic)].sub.Next(h.ctx)
		if err != nil {
			close(h.message)
			h.messageMetrics.AddErrorReceived()
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == h.self {
			continue
		}
		cm := new(Message)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			h.messageMetrics.AddErrorReceived()
			continue
		}
		// send valid messages onto the Messages channel
		h.messageMetrics.AddReceived()
		h.message <- cm

	}
}

//===================================
//========== Handle Events ==========
//===================================

// This function handle message communication, process incomming message and send message for validator
func handleEventsValidator(cr *Host, file *os.File, debugMode bool, nodeRole string, sizeParcel int, sizeBlock int, colRow int) {
	writer := csv.NewWriter(file)
	block := 0
	print(sizeParcel)
	nb_id := sizeBlock * 4 / sizeParcel
	id := 0
	data := []string{"TimeStamp", "Block", "Id", "Topic"}
	err := writer.Write(data)
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing CSV writer:", err)
	}

	for {
		select {
		//========== Receive Message ==========
		case m := <-cr.message:
			if id == nb_id {
				block += 1
			}

			// when we receive a message, print it to the message window
			data := []string{strconv.FormatInt(time.Now().Unix(), 10), strconv.Itoa(block), m.Id, m.Topic}

			err := writer.Write(data)
			if err != nil {
				log.Fatal("Error writing CSV:", err)
			}
			writer.Flush()
			if err := writer.Error(); err != nil {
				log.Fatal("Error flushing CSV writer:", err)
			}
			if debugMode {
				fmt.Println(time.Now(), "/ BLOCK:", block, "/ Id:", m.Id, "/ Topic:", m.Topic)

				if err != nil {
					fmt.Println("publish error: %s", err)
				}
			}
			id += 1
		}
	}
}

func handleEventsBuilder(cr *Host, file *os.File, debugMode bool, nodeRole string, sizeParcel int, sizeBlock int, colRow int) {
	peerRefreshTicker := time.NewTicker(1 * time.Millisecond)
	defer peerRefreshTicker.Stop()
	writer := csv.NewWriter(file)
	row_sample_list := idListRow(sizeParcel, sizeBlock)
	col_sample_list := idListCol(sizeParcel, sizeBlock)
	id := 0
	block := 0

	data := []string{"TimeStamp", "Block", "Id", "Sample col/row", "Topic"}

	err := writer.Write(data)
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing CSV writer:", err)
	}

	for {
		if id == len(row_sample_list) {
			id = 0
			block += 1
		}

		//send sample to column topic
		topic := "builder:c" + strconv.Itoa(id%sizeBlock)
		err := cr.Publish(topic, 0, id, block, sizeBlock)
		if debugMode {
			fmt.Println(time.Now(), "/ BLOCK:", block, "/ Col Id:", id, "/", len(col_sample_list), "/ Topic:", topic)
			if err != nil {
				fmt.Println("publish error: %s", err)
			}
		}
		data := []string{strconv.FormatInt(time.Now().Unix(), 10), strconv.Itoa(block), strconv.Itoa(id), strconv.Itoa(len(col_sample_list)), topic}

		err = writer.Write(data)
		if err != nil {
			log.Fatal("Error writing CSV:", err)
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatal("Error flushing CSV writer:", err)
		}

		//send sample to row topic
		topic = "builder:r" + strconv.Itoa((id-id%sizeBlock)/sizeBlock)
		if debugMode {
			fmt.Println(time.Now(), "/ BLOCK:", block, "/ Row Id:", id, "/", len(row_sample_list), "/ Topic:", topic)

			if err != nil {
				fmt.Println("publish error: %s", err)
			}
		}
		data = []string{strconv.FormatInt(time.Now().Unix(), 10), strconv.Itoa(block), strconv.Itoa(id), strconv.Itoa(len(row_sample_list)), topic}
		err = cr.Publish(topic, 1, id, block, sizeBlock)

		err = writer.Write(data)
		if err != nil {
			log.Fatal("Error writing CSV:", err)
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatal("Error flushing CSV writer:", err)
		}
		id += 1

	}
}

// ========== Util Function ==========
func findElementString(list []string, target string) int {
	for index, value := range list {
		if value == target {
			return index // Found the element, return its index
		}
	}
	return -1 // Element not found, return -1
}

func findElementInt(list []int, target int) int {
	for index, value := range list {
		if value == target {
			return index // Found the element, return its index
		}
	}
	return -1 // Element not found, return -1
}

func idListRow(sizeParcel int, sizeBlock int) []int {
	var result []int
	id := 0
	for id < sizeBlock*sizeBlock {
		result = append(result, id)
		id += sizeParcel
	}
	return result
}

func idListCol(sizeParcel int, sizeBlock int) []int {
	var result []int
	id := 0
	col := 0
	row := 0
	for col < sizeBlock {
		result = append(result, id)
		for i := 0; i < sizeParcel; i++ {
			row += 1
			if row == sizeBlock {
				col += 1
				row = 0
			}
		}
		id = row*sizeBlock + col
	}
	return result
}
