package main

import (
	"time"
	"math/rand"
)

//========== Struct definition ==========

type Message struct {
	Message    []byte
	SenderID   string
	SenderNick string
}

//Sample struct
type Sample struct {
	data []byte
	column int 
	row int
}

//Block struct
type Block struct {
	sampleList []Sample
	numberColumn int
	numberRow int
}

//Create a sample based on its size and place in the block
func CreateSample(column int, row int, size int) (*Sample){
	rand.Seed(time.Now().UnixNano())
	sliceLength := size
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