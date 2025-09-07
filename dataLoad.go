package main

import (
	"math/rand"
	"time"
)

var (
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type DataLoad []byte

func (p *DataLoad) Resize(size uint16) {
	buf := make([]byte, size)
	if _, err := rng.Read(buf); err != nil {
		log.Errorf("error resizing payload: %v", err)
		return
	}
	*p = DataLoad(buf)
}
