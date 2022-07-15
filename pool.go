package main

import (
	"log"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
)

func newPool(subwayLine string) *Pool {
	return &Pool{
		subwayLine: subwayLine,
		clients: make(map[*Client]bool),
		broadcast: make(chan Message), //make(chan []byte),
		broadcastV2: make(chan []*gtfs.TripUpdate_StopTimeUpdate),
		register: make(chan *Client),
		unregister: make(chan *Client),
		activeTrains: make(map[string][]map[uuid.UUID]*Client), //Do we need activeTrains anymore
		activeTrainChannel: make(chan string),
		cachedStopTimeUpdate: make(map[string]Message), //invalidate cache after X amount of time
		//This probably doesn't need to be a map anymore since every pool is scoped to a subwayline
		cachedStopTimeUpdateV2: make(map[string][]*gtfs.TripUpdate_StopTimeUpdate),
	}
}

func (p *Pool) run() {
	go p.fetchData()
	for {
		select {
			case client := <- p.register:
				log.Println("Register")
				p.clients[client] = true
				//if len(p.activeTrains[client.subwayLine]) == 0 {
				//	log.Println("FETCH")

				//}
				if len(p.activeTrains[client.subwayLine]) > 1 {
					log.Println("IMMEDIATELY RETURN CACHED GTFS DATA")
					client.send <- p.cachedStopTimeUpdate[client.subwayLine]
				}
			case client := <- p.unregister:
				if _,ok := p.clients[client]; ok {
					for i, c := range p.activeTrains[client.subwayLine]  {
						log.Printf("------ LOOP %v ------\n", c[client.UUID])
						line := client.subwayLine
						if c[client.UUID] != nil {
							log.Printf("___________ UNACTIVE  %v ___________\n", c[client.UUID])

							delete(c, client.UUID)
							delete(p.clients, client)
							close(client.send)

							p.activeTrains[line][i] = p.activeTrains[line][len(p.activeTrains[line])-1]
							p.activeTrains[line] = p.activeTrains[line][:len(p.activeTrains[line])-1]

							log.Printf("___________ Active Trains: %v ___________\n Length: %v ___________\n", p.activeTrains[line], len(p.activeTrains[line]))

							if len(p.activeTrains[line]) <= 0  {
								pm := *client.poolMap.poolMap
								pm[line] = nil
								return
							}
						}
					}
				}
			case message := <- p.broadcast:
				p.cachedStopTimeUpdate[message.Message.SubwayLine] = message
				log.Printf("SET CACHE: %v\n", p.cachedStopTimeUpdate[message.Message.SubwayLine])
				log.Printf("CLIENTS: %v \n", p.clients)
				for client := range p.clients {
					log.Printf("FOR CLIENT: %v \n", client.UUID)
					if client.subwayLine == message.Client.subwayLine {
						select {
							case client.send <- message:
								log.Printf("SEND: %v \n", message)
							default:
								client.fetching = false
						}
					}
				}
						//change name
			case broadcastPool := <- p.broadcastV2:
				p.cachedStopTimeUpdateV2[p.subwayLine] = broadcastPool
				for client := range p.clients {
					client.sendV2 <- broadcastPool
				}
			}
	}
}

					//log.Printf("Client: %v \n, Message: %+v \n", client.UUID, message.Message)
					//log.Println(" ")
					//log.Println(" ")
					//log.Printf("Subway Line: %v\n, Message Client Id: %v\n", message.Client.subwayLine, message.Client.UUID)


func (pool *Pool) fetchData() {
	log.Printf("Start Fetching Transit Times For POOL: %v \n", pool.subwayLine)
	defer func(){
		log.Printf("Closing Transit Data: %v\n", pool.subwayLine)
	}()
	//client.fetching = true
	ticker := time.NewTicker(10 * time.Second)
	i := 0

	//Need to send two messages to start for some reason or
	//else the client doesn't receive the first for ~20 secs
	for i < 2 {
		transitData := transitTimes(pool.subwayLine)
		//client.broadcast(transitData)
		pool.broadcastV2 <- transitData
		i++
	}

	for {
		time := <-ticker.C
		log.Printf("TIME: %v\n", time)
		transitData := transitTimes(pool.subwayLine)
		pool.broadcastV2 <- transitData
	}
}

//func (pool *Pool) broadcastData(transitData ArrivingTrain) bool {
//	var msg *Message
//	emptyMessage := ArrivingTrain{
//		ClientID: client.UUID,
//		SubwayLine: client.subwayLine,
//		Trains: make([]*Train, 0),
//	}

//	if len(transitData.Trains) > 0 {
//		log.Printf("DIRECTION: %v - TIME IN MINUTES: %v ",transitData.Trains[0].Direction, transitData.Trains[0].Train.TimeInMinutes)

//		msg = &Message{Client: client, Message: transitData}
//	} else {
//		log.Printf("EMPTY MESSAGE: %v", emptyMessage)
//		msg = &Message{Client: client, Message: emptyMessage}
//	}
//	pool.broadcast <- *msg
//	return true
//}