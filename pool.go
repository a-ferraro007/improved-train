package main

import (
	"log"

	"github.com/google/uuid"
)

func newPool() *Pool {
	return &Pool{
		clients: make(map[*Client]bool),
		broadcast: make(chan Message), //make(chan []byte),
		register: make(chan *Client),
		unregister: make(chan *Client),
		activeTrains: make(map[string][]map[uuid.UUID]*Client),
		activeTrainChannel: make(chan string),
		cachedStopTimeUpdate: make(map[string]Message), //invalidate cache after X amount of time
	}
}

func (p *Pool) run() {
	for {
		select {
			case client := <- p.register:
				p.clients[client] = true
				if len(p.activeTrains[client.subwayLine]) > 1 {
					log.Println("PING")
					client.send <- p.cachedStopTimeUpdate[client.subwayLine]
				}
			case client := <- p.unregister:
				if _,ok := p.clients[client]; ok {
					for i, c := range p.activeTrains[client.subwayLine]  {
						log.Printf("------ LOOP %v ------\n", c[client.UUID])
						line := client.subwayLine
						if c[client.UUID] != nil {
							log.Printf("___________ UNACTIVE  %v ___________\n", c[client.UUID])
							wasFetching := false
							if client.fetching {
								wasFetching = true
								client.fetching = false
							}

							delete(c, client.UUID)
							delete(p.clients, client)
							close(client.send)

							p.activeTrains[line][i] = p.activeTrains[line][len(p.activeTrains[line])-1]
							p.activeTrains[line] = p.activeTrains[line][:len(p.activeTrains[line])-1]

							log.Printf("___________ Active Trains: %v ___________\n Length: %v ___________\n", p.activeTrains[line], len(p.activeTrains[line]))

							if len(p.activeTrains[line]) != 0 && wasFetching{
								for _, value := range p.activeTrains[line][0]{
									next := value
									log.Default().Printf("_________NEXT CLIENT_________\n%v\n_______________________", next)
									go next.handleBroadcasting()
									break
								}
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
			}
	}
}

					//log.Printf("Client: %v \n, Message: %+v \n", client.UUID, message.Message)
					//log.Println(" ")
					//log.Println(" ")
					//log.Printf("Subway Line: %v\n, Message Client Id: %v\n", message.Client.subwayLine, message.Client.UUID)