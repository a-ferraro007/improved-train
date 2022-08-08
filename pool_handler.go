package main

import (
	"log"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (p *PoolMap) Init() {
	log.Println("INITIALIZE POOL MAP")
	p.Map = make(map[string]*Pool)
}

func (p *PoolMap) createPool(subwayLine string, headSignMap map[string]TripHeadSign) *Pool {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	pool := newPool(subwayLine, headSignMap)
	p.Map[subwayLine] = pool
	log.Println(p.Map)
	go pool.run()
	go pool.fetchData()
	return pool
}

func (p *PoolMap) insertIntoPool(subwayLine string, stopId string, conn *websocket.Conn) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	pool := p.Map[subwayLine]
	log.Println(p.Map)

	client := &Client{
		UUID:       uuid.New(),
		pool:       pool,
		conn:       conn,
		send:       make(chan []*gtfs.TripUpdate_StopTimeUpdate),
		stopId:     stopId,
		subwayLine: subwayLine,
		config:     Config{stopId: stopId, subwayLine: subwayLine, sort: "ascending"},
		fetching:   false,
	}
	client.configureSort()
	client.configureGenerator()

	mV2 := make([]*gtfs.TripUpdate_StopTimeUpdate, 0)
	mV2 = pool.cachedStopTimeUpdate[client.config.subwayLine]
	pool.register <- client
	go client.read()
	go client.write(&mV2)
}

func (p *PoolMap) deletePool(subwayLine string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.Map[subwayLine].done <- true
	delete(p.Map, subwayLine)
	log.Println("POOL MAP: ", p.Map)
}
