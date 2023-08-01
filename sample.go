package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const sampleLength = 42

// ========== Struct definition ==========
type Message struct {
	Message    []byte
	SenderID   string
	SenderNick string
	Topic      string
}

// Sample struct
type Sample struct {
	data   []byte
	column int
	row    int
	block  int
}

// Parcel struct
type Parcel struct {
	colRow int //0 = col; 1 = row
	block  int
	size   int
	first  int //first parcel if
	data   []byte
}

// Block struct
type Block struct {
	sampleList   []Sample
	numberColumn int
	numberRow    int
}

// Create a sample based on its size and place in the block
func CreateSample(column int, row int) *Sample {
	rand.Seed(time.Now().UnixNano())
	sliceLength := sampleLength
	randomSlice := make([]byte, sliceLength)
	rand.Read(randomSlice)

	s := &Sample{
		data:   randomSlice,
		column: column,
		row:    row,
	}
	return s
}

// Create an empty block
func CreateBlock(size int) *Block {
	b := &Block{
		sampleList:   nil,
		numberColumn: size,
		numberRow:    size,
	}
	return b
}

/*
func (b *Block) AddSample(column int, row int, sample *Sample) {
	b.sampleList = append(b.sampleList, sample)
}
*/

func CreateMessage(parcel *Parcel, topic string, sender peer.ID, nick string) *Message {

	message := make([]byte, 0)
	message = append(message, []byte(strconv.Itoa(parcel.colRow))...)
	message = append(message, []byte(strconv.Itoa(parcel.block))...)
	message = append(message, []byte(strconv.Itoa(parcel.size))...)
	message = append(message, []byte(strconv.Itoa(parcel.first))...)

	m := &Message{
		Message:    message,
		SenderID:   sender.Pretty(),
		SenderNick: nick,
		Topic:      topic,
	}

	return m
}

// Create a parcel based on its size and place in the block
func CreateParcel(colRow int, block int, size int, first int) *Parcel {
	rand.Seed(time.Now().UnixNano())
	sliceLength := sampleLength
	randomSlice := make([]byte, sliceLength*size)
	rand.Read(randomSlice)

	p := &Parcel{
		colRow: colRow,
		block:  block,
		size:   size,
		first:  first,
		data:   randomSlice,
	}
	return p
}
