package main

import (
	"log"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
)

func newPool(subwayLine string, headSignMap map[string]TripHeadSign) *Pool {
	return &Pool{
		subwayLine:         subwayLine,
		clients:            make(map[uuid.UUID]*Client), //make(map[*Client]bool),
		broadcast:          make(chan []*gtfs.TripUpdate_StopTimeUpdate),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		activeTrains:       make(map[string][]map[uuid.UUID]*Client), //Do we need activeTrains anymore
		activeTrainChannel: make(chan string),
		//This probably doesn't need to be a map anymore since every pool is scoped to a subwayline
		cachedStopTimeUpdate: make(map[string][]*gtfs.TripUpdate_StopTimeUpdate),
		ticker:               time.NewTicker(10 * time.Second),
		done:                 make(chan bool),
		TripHeadSignMap:      headSignMap,
	}
}

func (p *Pool) run() {
	for {
		select {
		case client := <-p.register:
			p.clients[client.UUID] = client
			log.Println("Register", len(p.clients))
		case client := <-p.unregister:
			if _, ok := p.clients[client.UUID]; ok {
				for _, c := range p.clients {
					line := client.config.subwayLine
					if client.UUID == c.UUID {
						log.Printf("___________ REMOVING CLIENT:   %v ___________\n", client.UUID)
						delete(p.clients, client.UUID)
						close(client.send)

						if len(p.clients) <= 0 {
							Pools.deletePool(line)
							return
						}
					}
				}
			}
		//change name
		case broadcast := <-p.broadcast:
			p.cachedStopTimeUpdate[p.subwayLine] = broadcast
			for _, client := range p.clients {
				log.Println("CLIENT SEND V2: ", client.UUID)
				client.send <- broadcast
			}
		}
	}
}

func (pool *Pool) fetchData() {
	defer log.Printf("Closing Transit Data: %v\n", pool.subwayLine)
	log.Printf("Start Fetching Transit Times For POOL: %v \n", pool.subwayLine)

	//Need to send two messages to start for some reason or
	//else the client doesn't receive the first for ~20 secs
	i := 0
	for i < 2 {
		transitData := handleFetchTransitData(pool.subwayLine)
		pool.broadcast <- transitData
		i++
	}

	for {
		select {
		case <-pool.done:
			return
		case time := <-pool.ticker.C:
			log.Printf("TIME: %v\n", time)
			transitData := handleFetchTransitData(pool.subwayLine)
			pool.broadcast <- transitData
		}
	}
}
