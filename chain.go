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

	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.

const ChainBufSize = 1280000000

type Chain struct {
	// Messages is a channel of messages received from other peers in the chat room
	block chan *Block

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

type Block struct {
	Message    []byte
	SenderID   string
	SenderNick string
}

func JoinChain(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*Chain, error) {
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

func handleEvents(cr *Chain, file *os.File, debugMode bool) {
	peerRefreshTicker := time.NewTicker( 1 * time.Second)
	defer peerRefreshTicker.Stop()

	//Open csv log file
	writer := csv.NewWriter(file)

	for {
	select {

	case m := <-cr.block:
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
			
	case <-peerRefreshTicker.C:
		err := cr.Publish(createAndConcatenate())
		if err != nil {
			fmt.Println("publish error: %s", err)
		}
				
		timestamp := time.Now()
		timeString := timestamp.Format("2006-01-02 15:04:05")
		data := []string{timeString , "Send", cr.nick}

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

func createAndConcatenate() []byte {
	// Initialize the random number generator
	rand.Seed(time.Now().UnixNano())

	// Define the length of each slice and the number of slices
	sliceLength := 42
	numSlices := 5*5

	// Create a slice to store the concatenated result
	concatenated := make([]byte, 0)

	// Create and concatenate the random slices
	for i := 0; i < numSlices; i++ {
		// Create a random slice
		randomSlice := make([]byte, sliceLength)
		rand.Read(randomSlice)

		// Concatenate the random slice to the result
		concatenated = append(concatenated, randomSlice...)
	}

	return concatenated
}