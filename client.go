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
		//log.Println(m.Message["ready"])
		//if m.Message["ready"] == true {
		//	broadcast(client)
		//}
		//log.Println("--------------------------------")
	}
}

func (client *Client) write(cachedMsg *Message){
	defer func(){
		log.Printf("Closing Write Client: %v \n", client.UUID)
	}()
	cache := cachedMsg.Message

	if cache.Trains != nil {
		log.Printf("CACHE HIT: %v\n", client.UUID)
		writeJSON(client, *cachedMsg)
	}

	for {
		msg, ok := <-client.send
		log.Printf("PONG: %v\n", client.UUID)
		if !ok {
			client.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		writeJSON(client, msg)
	}
}

func writeJSON(client *Client, msg Message) {
	w, errWriter := client.conn.NextWriter(websocket.TextMessage)
	if errWriter != nil {
		log.Println(errWriter)
		return
	}

	log.Printf("JSON: %v", msg.Message)

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
	log.Printf("bytes written: %v\n", l)
}

func (client *Client) handleBroadcasting() {
	log.Printf("Start Fetching Transit Times For Client: %v, %v \n", client.UUID, client.fetching)
	defer func(){
		log.Printf("Closing Transit Data: %v \n", client.UUID)
	}()
	client.fetching = true
	ticker := time.NewTicker(10 * time.Second)
	i := 0

	//Need to send two messages to start for some reason or
	//else the client doesn't receive the first for ~20 secs
	for i < 2 {
		transitData := transitTimes(client.subwayLine, client.stopId, client.UUID)
		client.broadcast(transitData)
		i++
	}

	for {
		if !client.fetching {
			ticker.Stop()
			log.Printf("Fetching False: %v \n", client.UUID)
			return
		}

		time := <-ticker.C
		log.Printf("TIME: %v\n", time)
		transitData := transitTimes(client.subwayLine, client.stopId, client.UUID)
		client.broadcast(transitData)
	}
}

func (client *Client) broadcast(transitData ArrivingTrain) bool {
	var msg *Message
	emptyMessage := ArrivingTrain{
		ClientID: client.UUID,
		SubwayLine: client.subwayLine,
		Trains: make([]*Train, 0),
	}

	if len(transitData.Trains) > 0 {
		log.Printf("DIRECTION: %v - TIME IN MINUTES: %v ",transitData.Trains[0].Direction, transitData.Trains[0].Train.TimeInMinutes)

		msg = &Message{Client: client, Message: transitData}
	} else {
		log.Printf("EMPTY MESSAGE: %v", emptyMessage)
		msg = &Message{Client: client, Message: emptyMessage}
	}
	client.pool.broadcast <- *msg
	return true
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


	var m Message
	if(len(clientSubwaySlice) == 0) {
		log.Println("Client Subway Line")
		newClient := make(map[uuid.UUID]*Client)
		newClient[client.UUID] = client
		pool.activeTrains[client.subwayLine] = append(pool.activeTrains[client.subwayLine], newClient)

		go client.handleBroadcasting()
		go client.read()
		go client.write(&m)

		client.pool.register <- client
		} else {
			log.Println("APPEND Client Subway Line")
			newClient := make(map[uuid.UUID]*Client)
			newClient[client.UUID] = client
			pool.activeTrains[client.subwayLine] = append(pool.activeTrains[client.subwayLine], newClient)
			m = pool.cachedStopTimeUpdate[client.subwayLine]
			//client.pool.register <- client
			//go client.write(&m)
			go client.read()
			go client.write(&m)
			client.pool.register <- client
		}
}

	//pool.activeTrains[client.subwayLine] = client
	//pool.activeTrainChannel <- client.stopId