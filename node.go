package main

import (
	"context"
	"encoding/json"
	"time"
	"fmt"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"math/rand"

	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const ChainBufSize = 1024
const BlockSize = 512
const NbSample = 75

type TopicSubItem struct {
	topic 		*pubsub.Topic
	sub 		*pubsub.Subscription
}

type Host struct {
	message 		chan *Message
	ctx 			context.Context
	ps 				*pubsub.PubSub
	topicNames 		[]string
	topicsubList 	[]TopicSubItem
	self     		peer.ID
	nick     		string
	messageMetrics	*MessageGlobalMetrics
}

func CreateHost(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, nodeType string) (*Host, error){

	//========== Subscribe nodes to topics ==========
	roomNameList := make([]string, 0)
	rand.Seed(time.Now().UnixNano())

	switch nodeType {

	//Subscribe builders to all row and column
	case "builder":
		for i := 0; i < BlockSize; i++ {
			roomNameList = append(roomNameList, "builder:c" + strconv.Itoa(i))
			roomNameList = append(roomNameList, "builder:r" + strconv.Itoa(i))
		}

	//Subscribe validators to 2 random row and 2 random column
	case "validator":
		column1 := rand.Intn(BlockSize)
		column2 := rand.Intn(BlockSize)
		row1 := rand.Intn(BlockSize)
		row2 := rand.Intn(BlockSize)
		roomNameList = append(roomNameList, "builder:c" + strconv.Itoa(column1))
		roomNameList = append(roomNameList, "builder:c" + strconv.Itoa(column2))
		roomNameList = append(roomNameList, "builder:r" + strconv.Itoa(row1))
		roomNameList = append(roomNameList, "builder:r" + strconv.Itoa(row2))

	//Non validator subscribe to 75 random samples
	default:

		sample := make([]int,0)
		i := 0
		for i < NbSample {
			id := rand.Intn(BlockSize*BlockSize)
			if(findElementInt(sample, id) == -1){
				sample = append(sample, id)
				column := id%BlockSize
				row :=(id - column)/BlockSize
				if findElementString(roomNameList, "builder:c" + strconv.Itoa(column)) == -1{
					roomNameList = append(roomNameList, "builder:c" + strconv.Itoa(column))
				}
				if findElementString(roomNameList, "builder:r" + strconv.Itoa(row)) == -1{
					roomNameList = append(roomNameList, "builder:r" + strconv.Itoa(row))
				}
				i += 1
			}
		}

	}
	mesMetrics := InitMessageMetrics(nodeType, nickname)
	h := &Host{
		ctx:			ctx,
		ps: 			ps,
		topicNames: 	roomNameList,
		topicsubList: 	make([]TopicSubItem, 0),
		self:     		selfID,
		nick:     		nickname,
		message: 		make(chan *Message, ChainBufSize),
		messageMetrics:	mesMetrics,
	}

	for i := 0; i < len(roomNameList); i++ {
		h.AddSubTopic(roomNameList[i])
		go h.readLoop(roomNameList[i])
	}
	// start reading message from the subscription in a loop
	return h, nil
}

func (h *Host) AddSubTopic(roomName string) (error) {
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
		topic:		topic,
		sub:		sub,
	}

	h.topicsubList = append(h.topicsubList, *tsi)

	return nil
}

// Publish sends a message to the pubsub topic.
func (h *Host) Publish(message []byte, topic string, column int, row int) error {
	m := CreateMessage(CreateSample(column, row), topic, h.self, h.nick)
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


//This function handle message communication, process incomming message and send message for validator
func handleEventsValidator(cr *Host, file *os.File, debugMode bool, nodeRole string) {
	writer := csv.NewWriter(file)
	for {
		select {
		//========== Receive Message ==========
		case m := <-cr.message:
				// when we receive a message, print it to the message window
				timestamp := time.Now()
				timeString := timestamp.Format("2006-01-02 15:04:05")
				data := []string{timeString , "Received:", m.SenderID, "Topic:", m.Topic}
				err := writer.Write(data)
				if err != nil {
					log.Fatal("Error writing CSV:", err)
				}
				writer.Flush()
				if err := writer.Error(); err != nil {
					log.Fatal("Error flushing CSV writer:", err)
				}
				if debugMode {
					fmt.Println(timestamp, " / ", m.Topic, " / ",m.SenderID, " / ", m.Message)
				}
		}
	}
}

func handleEventsBuilder(cr *Host, file *os.File, debugMode bool, nodeRole string) {
	peerRefreshTicker := time.NewTicker( 1 * time.Millisecond)
	defer peerRefreshTicker.Stop()
	writer := csv.NewWriter(file)
	topic := "test1"
	row_id := 0
	column_id := 0


	for{
		if column_id == BlockSize{
			column_id = 0
			row_id += 1
		}

		if row_id == BlockSize{
			row_id += 0
		}
		//send sample to column topic
		topic = "builder:c" + strconv.Itoa(column_id)
		fmt.Println("BLOCK:test User:",cr.nick," Topic:", topic)
		err := cr.Publish([]byte(strconv.Itoa(column_id)), topic, column_id, row_id)
		if err != nil {
			fmt.Println("publish error: %s", err)
		}
		//send sample to row topic
		topic = "builder:r" + strconv.Itoa(row_id)
		fmt.Println("BLOCK:test User:",cr.nick," Topic:", topic)
		err = cr.Publish([]byte(strconv.Itoa(row_id)), topic, column_id, row_id)
		if err != nil {
			fmt.Println("publish error: %s", err)
		}
		
		column_id += 1		
													
		timestamp := time.Now()
		timeString := timestamp.Format("2006-01-02 15:04:05")
		data := []string{timeString , "PUT", strconv.Itoa(0), strconv.Itoa(0), strconv.Itoa(0), cr.nick, topic}

		err = writer.Write(data)
		if err != nil {
			log.Fatal("Error writing CSV:", err)
		}
								
		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatal("Error flushing CSV writer:", err)
		}
	}
}


//========== Util Function ==========
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