package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	message      chan *Message
	ctx          context.Context
	ps           *pubsub.PubSub
	topicNames   []string
	topicsubList []TopicSubItem
	self         peer.ID
	nick         string
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
		roomNameList = append(roomNameList, "builder:header_dis")
	//Subscribe validators to 2 random row and 2 random column
	case "validator":
		//column1 := rand.Intn(BlockSize)
		//column2 := rand.Intn(BlockSize)
		//row1 := rand.Intn(BlockSize)
		//row2 := rand.Intn(BlockSize)
		column1 := 1
		column2 := 2
		row1 := 1
		row2 := 2

		//builder topic
		roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(column1))
		roomNameList = append(roomNameList, "builder:c"+strconv.Itoa(column2))
		roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(row1))
		roomNameList = append(roomNameList, "builder:r"+strconv.Itoa(row2))

		//header topic
		roomNameList = append(roomNameList, "builder:header_dis")

		//validator topic
		roomNameList = append(roomNameList, "validator:c"+strconv.Itoa(column1))
		roomNameList = append(roomNameList, "validator:c"+strconv.Itoa(column2))
		roomNameList = append(roomNameList, "validator:r"+strconv.Itoa(row1))
		roomNameList = append(roomNameList, "validator:r"+strconv.Itoa(row2))

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
				if findElementString(roomNameList, "validator:c"+strconv.Itoa(column)) == -1 {
					roomNameList = append(roomNameList, "validator:c"+strconv.Itoa(column))
				}
				if findElementString(roomNameList, "validator:r"+strconv.Itoa(row)) == -1 {
					roomNameList = append(roomNameList, "validator:r"+strconv.Itoa(row))
				}
				i += 1
			}
		}
		roomNameList = append(roomNameList, "builder:header_dis")

	}
	h := &Host{
		ctx:          ctx,
		ps:           ps,
		topicNames:   roomNameList,
		topicsubList: make([]TopicSubItem, 0),
		self:         selfID,
		nick:         nickname,
		message:      make(chan *Message, ChainBufSize),
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
func (h *Host) Publish(topic string, colRow int, first int, block int, size int, logger *log.Logger, builder bool) error {
	m := CreateMessage(CreateParcel(colRow, block, size, first), topic, h.self, h.nick, first, block)
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if builder {
		if colRow == 1 {
			logger.Println(formatJSONLogMessageSend(h.nick, colRow, topic, MessageType(1)))
		} else {
			logger.Println(formatJSONLogMessageSend(h.nick, colRow, topic, MessageType(0)))
		}
	} else {
		if colRow == 1 {
			logger.Println(formatJSONLogMessageSend(h.nick, colRow, topic, MessageType(7)))
		} else {
			logger.Println(formatJSONLogMessageSend(h.nick, colRow, topic, MessageType(6)))
		}
	}

	return h.topicsubList[findElementString(h.topicNames, topic)].topic.Publish(h.ctx, msgBytes)
}

func (h *Host) PublishHeader(topic string, block int, logger *log.Logger) error {
	m := CreateMessageHeader(topic, h.self, h.nick, block)
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	logger.Println(formatJSONLogHeaderSend(h.nick, topic, block, BuilderPublishHeader))
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
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == h.self {
			continue
		}
		cm := new(Message)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			fmt.Println("AAAAA")
			continue
		}
		// send valid messages onto the Messages channel
		h.message <- cm

	}
}

//===================================
//========== Handle Events ==========
//===================================

// This function handle message communication, process incomming message and send message for validator
func handleEventsValidator(cr *Host, file_log *os.File, debugMode bool, nodeRole string, sizeParcel int, sizeBlock int, colRow int, logger *log.Logger, duration int) {
	block := 0
	//nb_id := sizeBlock * 2 / sizeParcel
	expeDurationTicker := time.NewTicker(time.Duration(duration) * time.Second)

	for {
		select {
		case <-expeDurationTicker.C:
			if debugMode {
				fmt.Println("Exit part")
			}
			return
		//========== Receive Message ==========
		case m := <-cr.message:

			if m.Topic == "builder:header_dis" {
				block += 1
				logger.Println(formatJSONLogHeaderSend(m.SenderID, m.Topic, block, MessageType(4)))
			}
			if m.Topic != "builder:header_dis" {
				idBlock, _ := strconv.Atoi(m.Block)
				colRow, _, _, _ := readMessage(m.Message)
				if colRow == 1 {
					logger.Println(formatJSONLogMessageSend(m.SenderID, colRow, m.Topic, MessageType(2)))
				} else {
					logger.Println(formatJSONLogMessageSend(m.SenderID, colRow, m.Topic, MessageType(3)))
				}
				if idBlock == -1 {
					return
				}
				topic := "validator:" + strings.Split(m.Topic, ":")[1]
				id := strings.Split(m.Topic, ":")[1][1:]
				idi, err := strconv.Atoi(id)
				err = cr.Publish(topic, 0, idi, block, sizeBlock, logger, false)
				if err != nil {
					log.Fatal("Publish failed column", err)
				}
				// when we receive a message, print it to the message window
				if debugMode {
					timestamp := time.Now().UnixNano() / int64(time.Millisecond)
					fmt.Println(timestamp, "/ BLOCK:", m.Block, "/ Id:", m.Id, "/ Topic:", m.Topic)
				}
			}
		}
	}
}

func handleEventsBuilder(cr *Host, file *os.File, debugMode bool, sizeParcel int, sizeBlock int, duration int, logger *log.Logger) {
	peerRefreshTicker := time.NewTicker(1 * time.Millisecond)
	defer peerRefreshTicker.Stop()
	row_sample_list := idListRow(sizeParcel, sizeBlock)
	col_sample_list := idListCol(sizeParcel, sizeBlock)
	id := 0
	block := 0

	blockGenerationTicker := time.NewTicker(Blocktime * time.Second)
	expeDurationTicker := time.NewTicker(time.Duration(duration) * time.Second)
	defer expeDurationTicker.Stop()
	defer blockGenerationTicker.Stop()

	for {
		select {
		case <-blockGenerationTicker.C:
			block += 1
			id = 0
			cr.PublishHeader("builder:header_dis", block, logger)

			for id < sizeBlock {
				// ====================send sample to column topic ====================
				topic := "builder:c" + strconv.Itoa(id)
				err := cr.Publish(topic, 0, id, block, sizeBlock, logger, true)
				if err != nil {
					log.Fatal("Publish failed column", err)
				}

				//Debug Log
				if debugMode {
					timestamp := time.Now().UnixNano() / int64(time.Millisecond)
					fmt.Println(timestamp, "/ BLOCK:", block, "/ Col Id:", id, "/", len(col_sample_list), "/ Topic:", topic)
				}

				// ====================send sample to row topic ====================
				topic = "builder:r" + strconv.Itoa(id)
				err = cr.Publish(topic, 1, id, block, sizeBlock, logger, true)
				if err != nil {
					log.Fatal("Publish failed row", err)
				}

				//Debug Log
				if debugMode {
					timestamp := time.Now().UnixNano() / int64(time.Millisecond)
					fmt.Println(timestamp, "/ BLOCK:", block, "/ Row Id:", id, "/", len(row_sample_list), "/ Topic:", topic)
				}

				//Write message sent to Log file
				id += 1
			}
		case <-expeDurationTicker.C:
			if debugMode {
				fmt.Println("Exit part")
			}
			return

		}
	}

}

func handleEventsNonValidator(cr *Host, file_log *os.File, debugMode bool, nodeRole string, sizeParcel int, sizeBlock int, colRow int, logger *log.Logger, duration int) {
	block := 0
	print(sizeParcel)
	//nb_id := sizeBlock * 2 / sizeParcel
	expeDurationTicker := time.NewTicker(time.Duration(duration) * time.Second)

	for {
		select {
		case <-expeDurationTicker.C:
			if debugMode {
				fmt.Println("Exit part")
			}
			return
		//========== Receive Message ==========
		case m := <-cr.message:
			if m.Topic == "builder:header_dis" {
				block += 1
				logger.Println(formatJSONLogHeaderSend(m.SenderID, m.Topic, block, MessageType(2)))
			}

			if m.Topic != "builder:header_dis" {
				idBlock, _ := strconv.Atoi(m.Block)
				colRow, _, _, _ := readMessage(m.Message)
				if colRow == 1 {
					logger.Println(formatJSONLogMessageSend(m.SenderID, colRow, m.Topic, MessageType(8)))
				} else {
					logger.Println(formatJSONLogMessageSend(m.SenderID, colRow, m.Topic, MessageType(9)))
				}
				if idBlock == -1 {
					return
				}
				// when we receive a message, print it to the message window
				if debugMode {
					timestamp := time.Now().UnixNano() / int64(time.Millisecond)
					fmt.Println(timestamp, "/ BLOCK:", m.Block, "/ Id:", m.Id, "/ Topic:", m.Topic)
				}
			}
		}
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
