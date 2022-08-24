package clientpool

import (
	"log"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/a-ferraro007/improved-train/pkg/utils"
	"github.com/google/uuid"
)

//Pool struct holds all pool related data
type Pool struct {
	SubwayLine           string
	Clients              map[uuid.UUID]*Client
	Broadcast            chan []*gtfs.TripUpdate_StopTimeUpdate
	Register             chan *Client
	Unregister           chan *Client
	ActiveTrains         map[string][]map[uuid.UUID]*Client
	ActiveTrainChannel   chan string
	CachedStopTimeUpdate map[string][]*gtfs.TripUpdate_StopTimeUpdate
	Ticker               *time.Ticker
	Done                 chan bool
}

func newPool(subwayLine string) *Pool {
	return &Pool{
		SubwayLine:         subwayLine,
		Clients:            make(map[uuid.UUID]*Client), //make(map[*Client]bool),
		Broadcast:          make(chan []*gtfs.TripUpdate_StopTimeUpdate),
		Register:           make(chan *Client),
		Unregister:         make(chan *Client),
		ActiveTrains:       make(map[string][]map[uuid.UUID]*Client), //Do we need activeTrains anymore
		ActiveTrainChannel: make(chan string),
		//This probably doesn't need to be a map anymore since every pool is scoped to a subwayline
		CachedStopTimeUpdate: make(map[string][]*gtfs.TripUpdate_StopTimeUpdate),
		Ticker:               time.NewTicker(10 * time.Second),
		Done:                 make(chan bool),
	}
}

func (p *Pool) run() {
	for {
		select {
		case client := <-p.Register:
			p.Clients[client.UUID] = client
			log.Println("Register", len(p.Clients))
		case client := <-p.Unregister:
			if _, ok := p.Clients[client.UUID]; ok {
				for _, c := range p.Clients {
					line := client.Config.SubwayLine
					if client.UUID == c.UUID {
						log.Printf("___________ REMOVING CLIENT:   %v ___________\n", client.UUID)
						delete(p.Clients, client.UUID)
						close(client.Send)

						if len(p.Clients) <= 0 && line != "" {
							Pools.DeletePool(line)
							return
						}
					}
				}
			}
		//change name
		case broadcast := <-p.Broadcast:
			p.CachedStopTimeUpdate[p.SubwayLine] = broadcast
			for _, client := range p.Clients {
				log.Println("CLIENT SEND V2: ", client.UUID)
				client.Send <- broadcast
			}
		}
	}
}

func (p *Pool) fetchData() {
	defer log.Printf("Closing Transit Data: %v\n", p.SubwayLine)
	log.Printf("Start Fetching Transit Times For POOL: %v \n", p.SubwayLine)

	//Need to send two messages to start for some reason or
	//else the client doesn't receive the first for ~20 secs
	i := 0
	for i < 2 {
		transitData := utils.HandleFetchTransitData(p.SubwayLine)
		p.Broadcast <- transitData
		i++
	}

	for {
		select {
		case <-p.Done:
			return
		case time := <-p.Ticker.C:
			log.Printf("TIME: %v\n", time)
			transitData := utils.HandleFetchTransitData(p.SubwayLine)
			p.Broadcast <- transitData
		}
	}
}
