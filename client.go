package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (client *Client) read() {
	defer func(){
		log.Printf("Closing Read Client: %v \n", client.UUID)
		client.pool.unregister <- client
		client.conn.Close()
	}()

	for  {
		//log.Println("+++++++++++++++++++++++++++++++++")
		m := &RespMsg{}
		//log.Println(m)

		_, d, readerErr := 	client.conn.ReadMessage()
		if readerErr != nil {
			log.Println(readerErr)
			return
		}
		err := json.Unmarshal(d, &m.Message)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println(m.Message)
		//log.Println("--------------------------------")
	}
}

func (client *Client) write(){
	defer func(){
		log.Printf("Closing Write Client: %v \n", client.UUID)
		client.conn.Close()
	}()

	for {
		select {
			case msg, ok := <-client.send:
				log.Println("PONG")
				if !ok {
					client.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, errWriter := client.conn.NextWriter(websocket.TextMessage)
				if errWriter != nil {
					log.Println(errWriter)
					return
				}

				json, jsonErr := json.Marshal(msg.Message)
				if jsonErr != nil {
					log.Println(jsonErr)
					return
				}
				l, errNW := w.Write(json)
				if errNW != nil {
					log.Println(errNW)
					return
				}
				log.Println(l)
		}
	}
}

func transitData(client *Client) {
	log.Printf("Start Fetching Transit Times For Client: %v, %v \n", client.UUID, client.fetching)
	defer func(){
		log.Printf("Closing Transit Data: %v \n", client.UUID)
		//client.conn.Close()
	}()

	client.fetching = true
	var msg *Message

	transitData := transitTimes(client.subwayLine, client.stopId, client.UUID)
	log.Printf("DIRECTION: %v - TIME IN MINUTES: %v ",transitData.Trains[0].Direction, transitData.Trains[0].Train.TimeInMinutes)

	msg = &Message{Client: client, Message: transitData}
	client.pool.broadcast <- *msg

	t := time.Second * 20
	for range	time.Tick(t){
		if !client.fetching {
			log.Printf("Fetching False: %v \n", client.UUID)
			return
		}

		transitData := transitTimes(client.subwayLine, client.stopId, client.UUID)
		log.Printf("DIRECTION: %v - TIME IN MINUTES: %v ",transitData.Trains[0].Direction, transitData.Trains[0].Train.TimeInMinutes)

		msg = &Message{Client: client, Message: transitData}
		client.pool.broadcast <- *msg
	}
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w,r, nil)
	if err != nil {
		log.Println(err)
	}
	urlQuery := r.URL.Query()

	log.Println(r.URL.Query())
	client := &Client{
			UUID: uuid.New(),
			pool: pool,
			conn: conn,
			send: make(chan Message),//make(chan []byte),
			stopId: urlQuery["stopId"][0],
			subwayLine: urlQuery["subwayLine"][0],
			fetching: false,
		}

	clientSubwaySlice := pool.activeTrains[client.subwayLine]
	//Need to keep track of Train lines, not StopID's
	if(len(clientSubwaySlice) == 0) {
		log.Println("Client Subway Line")
		newClient := make(map[uuid.UUID]*Client)
		newClient[client.UUID] = client
		pool.activeTrains[client.subwayLine] = append(pool.activeTrains[client.subwayLine], newClient)
		log.Println(pool.activeTrains["L"])
		go transitData(client)
	} else {
		log.Println("APPEND Client Subway Line")
		newClient := make(map[uuid.UUID]*Client)
		newClient[client.UUID] = client
		pool.activeTrains[client.subwayLine] = append(pool.activeTrains[client.subwayLine], newClient)
		//pool.activeTrains[client.subwayLine] = append(pool.activeTrains[client.subwayLine], client)
		log.Println(pool.activeTrains["L"])
	}

	go client.write()
	go client.read()

	client.pool.register <- client
}

	//pool.activeTrains[client.subwayLine] = client
	//pool.activeTrainChannel <- client.stopId