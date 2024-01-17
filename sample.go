package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const sampleLength = 512

// ========== Struct definition ==========
type Message struct {
	Message    []byte
	SenderID   string
	SenderNick string
	Topic      string
	Id         string
	Block      string
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

type Header struct {
	data []byte
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

func CreateMessage(parcel *Parcel, topic string, sender peer.ID, nick string, id int, block int) *Message {

	message := make([]byte, 0)
	message = append(message, []byte(strconv.Itoa(parcel.colRow))...)
	message = append(message, 0x00)
	message = append(message, 0xFF)
	message = append(message, 0x00)
	message = append(message, []byte(strconv.Itoa(parcel.block))...)
	message = append(message, 0x00)
	message = append(message, 0xFF)
	message = append(message, 0x00)
	message = append(message, []byte(strconv.Itoa(parcel.size))...)
	message = append(message, 0x00)
	message = append(message, 0xFF)
	message = append(message, 0x00)
	message = append(message, []byte(strconv.Itoa(parcel.first))...)

	m := &Message{
		Message:    message,
		SenderID:   sender.ShortString(),
		SenderNick: nick,
		Topic:      topic,
		Id:         strconv.Itoa(id),
		Block:      strconv.Itoa(block),
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

func readMessage(parcel []byte) (int, int, int, int) {
	delimiter := []byte{0x00, 0xFF, 0x00}
	elements := bytes.Split(parcel, delimiter)
	colRow, err := strconv.Atoi(string(elements[0]))
	if err != nil {
		fmt.Println("Error:", err)
	}
	block, err := strconv.Atoi(string(elements[1]))
	if err != nil {
		fmt.Println("Error:", err)
	}
	size, err := strconv.Atoi(string(elements[2]))
	if err != nil {
		fmt.Println("Error:", err)
	}
	first, err := strconv.Atoi(string(elements[3]))
	if err != nil {
		fmt.Println("Error:", err)
	}
	return colRow, block, size, first
}

func CreateMessageHeader(topic string, sender peer.ID, nick string, block int) *Message {

	m := &Message{
		Message:    make([]byte, sizeHeader),
		SenderID:   sender.ShortString(),
		SenderNick: nick,
		Topic:      topic,
		Id:         strconv.Itoa(-1),
		Block:      strconv.Itoa(block),
	}

	return m
}
