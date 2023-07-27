package main

import (
	"time"
	"math/rand"
	"strconv"

	"github.com/libp2p/go-libp2p/core/peer"

)

const sampleLength = 42
//========== Struct definition ==========
type Message struct {
	Message    []byte
	SenderID   string
	SenderNick string
	Topic 	   string
}

//Sample struct
type Sample struct {
	data []byte
	column int 
	row int
	block int
}

//Block struct
type Block struct {
	sampleList []Sample
	numberColumn int
	numberRow int
}

//Create a sample based on its size and place in the block
func CreateSample(column int, row int) (*Sample){
	rand.Seed(time.Now().UnixNano())
	sliceLength := sampleLength
	randomSlice := make([]byte, sliceLength)
	rand.Read(randomSlice)

	s:= &Sample{
		data:		randomSlice,
		column:		column,
		row:		row,
	}
	return s
}

//Create an empty block
func CreateBlock(size int) (*Block){
	b := &Block{
		sampleList:		nil,
		numberColumn: 	size,
		numberRow:		size,
	}
	return b
}

/*
func (b *Block) AddSample(column int, row int, sample *Sample) {
	b.sampleList = append(b.sampleList, sample)
}
*/

func CreateMessage(sample *Sample, topic string, sender peer.ID, nick string) (*Message){

		message := make([]byte, 0)
		message = append(message, []byte(strconv.Itoa(sample.block))...)
		message = append(message, []byte(strconv.Itoa(sample.row))...)
		message = append(message, []byte(strconv.Itoa(sample.column))...)
		message = append(message, sample.data...)

		m := &Message{
		Message:    message,
		SenderID:   sender.Pretty(),
		SenderNick: nick,
		Topic: 		topic,
	}

	return m
}