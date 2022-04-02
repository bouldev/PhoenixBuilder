package connection

import (
	"fmt"
	"sync"
)

const (
	MAX_READ_QUEUE_SIZE = 0
)

type Mux struct {
	baseChannel ReliableConnetion
	SubChannels map[byte]ReliableConnetion
	rchan       map[byte]chan []byte
	wsync       sync.Mutex
	isClosed    bool
}

func NewMux(baseChannel ReliableConnetion) *Mux {
	m := &Mux{
		baseChannel: baseChannel,
		SubChannels: make(map[byte]ReliableConnetion),
		rchan:       make(map[byte]chan []byte),
		wsync:       sync.Mutex{},
		isClosed:    false,
	}
	go func() {
		for {
			data, err := baseChannel.RecvFrame()
			if err != nil {
				return
			}
			key := data[0]
			payload := data[1:]
			if _, ok := m.rchan[key]; ok {
				m.rchan[key] <- payload
			} else {
				m.GetSubChannel(key)
				m.rchan[key] <- payload
			}
		}
	}()
	// go func() {
	// 	for {
	// 		data := <-m.wchan
	// 		if m.baseChannel.SendFrame(data) != nil {
	// 			return
	// 		}
	// 	}
	// }()
	return m
}

func (m *Mux) GetSubChannel(key byte) ReliableConnetion {
	sc, ok := m.SubChannels[key]
	if ok {
		return sc
	}
	c := &SubChannel{
		idenficationKey: key,
		mux:             m,
		baseChannel:     m.baseChannel,
	}
	m.SubChannels[key] = c
	if _, ok := m.rchan[key]; !ok {
		m.rchan[key] = make(chan []byte, MAX_READ_QUEUE_SIZE)
	}
	return c
}

type SubChannel struct {
	idenficationKey byte
	mux             *Mux
	baseChannel     ReliableConnetion
}

func (sc *SubChannel) Init() error {
	return nil
}

func (sc *SubChannel) SendFrame(data []byte) error {
	if sc.mux.isClosed {
		return fmt.Errorf("mux is closed")
	}
	sc.mux.wsync.Lock()
	defer sc.mux.wsync.Unlock()
	return sc.baseChannel.SendFrame(append([]byte{sc.idenficationKey}, data...))
}

func (sc *SubChannel) RecvFrame() ([]byte, error) {
	if sc.mux.isClosed {
		return nil, fmt.Errorf("mux is closed")
	}
	r := <-sc.mux.rchan[sc.idenficationKey]
	if r == nil {
		return nil, fmt.Errorf("no data")
	}
	return r, nil
}
