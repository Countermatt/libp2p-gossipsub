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

	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const ChainBufSize = 1024


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

}

func CreateHost(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomNameList []string) (*Host, error){

	h := &Host{
		ctx:			ctx,
		ps: 			ps,
		topicNames: 	roomNameList,
		topicsubList: 	make([]TopicSubItem, 0),
		self:     		selfID,
		nick:     		nickname,
		message: 		make(chan *Message, ChainBufSize),
	}

	for i := 0; i < len(roomNameList); i++ {
		h.AddSubTopic(roomNameList[i])
	}
	// start reading message from the subscription in a loop
	go h.readLoop()
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
func (h *Host) Publish(message []byte, topic string) error {
	m := Message{
		Message:    message,
		SenderID:   h.self.Pretty(),
		SenderNick: h.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return h.topicsubList[findElement(h.topicNames, topic)].topic.Publish(h.ctx, msgBytes)
}
 /*
func (h *Host) ListPeers() []peer.ID {
	return h.ps.ListPeers(topicName(h.roomName))
}
*/
func (h *Host) readLoop() {
	for {

		for i := 0; i < len(h.topicsubList); i++ {
			msg, err := h.topicsubList[i].sub.Next(h.ctx)
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
				continue
			}
			// send valid messages onto the Messages channel
			h.message <- cm

		}
	}
}

func handleEvents(cr *Host, file *os.File, debugMode bool, nodeRole string) {
	peerRefreshTicker := time.NewTicker( 1 * time.Second)
	defer peerRefreshTicker.Stop()
	writer := csv.NewWriter(file)

	for {
		select {
		//========== Receive Message ==========
		case m := <-cr.message:
			//if nodeRole != "builder" {
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
			//}

		//========== Send Message ==========
		case <-peerRefreshTicker.C:
				if nodeRole == "builder" {

					fmt.Println("BLOCK:test User:",cr.nick)
					
					err := cr.Publish([]byte("test"), "test")
					if err != nil {
						fmt.Println("publish error: %s", err)
					}
							
					timestamp := time.Now()
					timeString := timestamp.Format("2006-01-02 15:04:05")
					data := []string{timeString , "PUT", strconv.Itoa(0), strconv.Itoa(0), strconv.Itoa(0), cr.nick}

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



//========== Util Function ==========
func findElement(list []string, target string) int {
    for index, value := range list {
        if value == target {
            return index // Found the element, return its index
        }
    }
    return -1 // Element not found, return -1
}