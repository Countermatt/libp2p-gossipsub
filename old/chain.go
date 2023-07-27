package main

import (
	"context"
	"encoding/json"
	"time"
	"fmt"
	"math/rand"
	"encoding/csv"
	"log"
	"os"
	"strconv"

	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.

const ChainBufSize = 1280000000

type SampleGiven struct {
	Block int
	Column int 
	Row int
}

type TopicItem struct {
	topic *pubsub.Topic
}

type Chain struct {
	// Messages is a channel of messages received from other peers in the chat room
	block chan *Block

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	//topic []TopicItem
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string

	//Builder Item
	SampleList []SampleGiven
	sizeBlock int
}

type Block struct {
	Message    []byte
	SenderID   string
	SenderNick string
}

func JoinChain(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string, sizeBlock int) (*Chain, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &Chain{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		block: make(chan *Block, ChainBufSize),
		sizeBlock: sizeBlock,
	}

	// start reading block from the subscription in a loop
	go cr.readLoop()
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *Chain) Publish(message []byte) error {
	m := Block{
		Message:    message,
		SenderID:   cr.self.Pretty(),
		SenderNick: cr.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.topic.Publish(cr.ctx, msgBytes)
}

func (cr *Chain) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.roomName))
}

func (cr *Chain) readLoop() {
	for {
		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			close(cr.block)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}
		cm := new(Block)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		cr.block <- cm
	}
}

func handleEvents(cr *Chain, file *os.File, debugMode bool, nodeRole string) {
	peerRefreshTicker := time.NewTicker( 1 * time.Millisecond)
	defer peerRefreshTicker.Stop()


	//for builder
	Sendsample := createsample()
	blockNumber := 0
	row := 0
	column := 0
	//Open csv log file
	writer := csv.NewWriter(file)
	rand.Seed(time.Now().UnixNano())
	for {

		select {

		case m := <-cr.block:
			if nodeRole != "builder" {
				// when we receive a message, print it to the message window
				timestamp := time.Now()
				timeString := timestamp.Format("2006-01-02 15:04:05")
				data := []string{timeString , "Received", m.SenderID}

				err := writer.Write(data)
				if err != nil {
					log.Fatal("Error writing CSV:", err)
				}

				writer.Flush()
				if err := writer.Error(); err != nil {
					log.Fatal("Error flushing CSV writer:", err)
				}
				if debugMode {
					fmt.Println(timestamp, "  ", m.SenderID)
				}
			}
		case <-peerRefreshTicker.C:
				if nodeRole == "builder" {


					Sample := SampleGiven{blockNumber, row,  column}

					if (row%cr.sizeBlock == 0 &&  column%cr.sizeBlock == 0 && column>0  && row>0){
						blockNumber += 1
						column = 0
						row = 0
					}
					if (row%cr.sizeBlock == 0 && row >0){
						column += 1
						row = 0
					}

					row += 1
					fmt.Println("BLOCK:", Sample.Block, "/SAMPLE:", Sample.Column, Sample.Row)
					
					cr.SampleList = append(cr.SampleList, Sample)
					err := cr.Publish(publishBlock(Sample, Sendsample))
					if err != nil {
						fmt.Println("publish error: %s", err)
					}
							
					timestamp := time.Now()
					timeString := timestamp.Format("2006-01-02 15:04:05")
					data := []string{timeString , "PUT", strconv.Itoa(Sample.Block), strconv.Itoa(Sample.Column), strconv.Itoa(Sample.Row), cr.nick}

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
	}
}
func createsample() []byte {
	rand.Seed(time.Now().UnixNano())
	sliceLength := 42
	randomSlice := make([]byte, sliceLength)
	rand.Read(randomSlice)
	return randomSlice
}

func publishBlock(Sample SampleGiven, Sendsample []byte) []byte {
	// Initialize the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a slice to store the concatenated result
	concatenated := make([]byte, 0)

	concatenated = append(concatenated, byte(Sample.Block))
	concatenated = append(concatenated, byte(Sample.Column))
	concatenated = append(concatenated, byte(Sample.Row))
	concatenated = append(concatenated, Sendsample...)
	return concatenated
}


func containsTuple(list []SampleGiven, target SampleGiven) bool {
	for _, tuple := range list {
		if tuple == target {
			return true
		}
	}
	return false
}

func randomIntInRange(x int) int {
	rand.Seed(time.Now().UnixNano()) // Initialize the random number generator with a seed based on the current time
	return rand.Intn(x + 1) // Generate a random number between 0 and x (inclusive)
}