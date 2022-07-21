package main

import (
	"encoding/json"
	"log"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/gorilla/websocket"
)

func (client *Client) read() {
	defer func() {
		log.Printf("Closing Read Client: %v \n", client.UUID)
		client.pool.unregister <- client
		client.conn.Close()
	}()

	for {
		m := &RespMsg{}

		_, d, readerErr := client.conn.ReadMessage()
		if readerErr != nil {
			log.Println(readerErr)
			return
		}
		err := json.Unmarshal(d, &m.Message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (client *Client) write(cachedGTFSData *[]*gtfs.TripUpdate_StopTimeUpdate) {
	defer log.Printf("Closing Write Client: %v \n", client.UUID)

	stopTimeUpdateSlice := make([]*StopTimeUpdate, 0)
	arrivingTrain := &ArrivingTrain{ClientID: client.UUID, SubwayLine: client.config.subwayLine}

	if len(*cachedGTFSData) != 0 {
		log.Printf("CACHE HIT: %v\n", client.UUID)
		for _, tripUpdate := range *cachedGTFSData {
			match, stopTimeUpdate := findStopData(tripUpdate, client.config.stopId)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice)
			arrivingTrain.Trains = unparsed
			arrivingTrain.ParsedTrains = client.config.funct(parsed)
		}

		i := 0
		for i < 2 {
			client.writeJSON(Message{Client: client, Message: *arrivingTrain})
			i++
		}
	}
	log.Println("FETCH")
	for {
		data, ok := <-client.send
		//log.Printf("DATA RECEIVED FOR CLIENT: %v\n", client.UUID)
		if !ok {
			client.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		arrivingTrain.Trains = make([]*Train, 0)
		stopTimeUpdateSlice = make([]*StopTimeUpdate, 0)

		for _, tripUpdate := range data {
			match, stopTimeUpdate := findStopData(tripUpdate, client.config.stopId)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}
		log.Println("FETCH 2", stopTimeUpdateSlice)
		if len(stopTimeUpdateSlice) > 0 {
			log.Println("FETCH 3")
			unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice)
			arrivingTrain.Trains = unparsed
			log.Println("FETCH 4", unparsed)
			arrivingTrain.ParsedTrains = parsed
		}
		log.Println("WRITE JSON")
		client.writeJSON(Message{Client: client, Message: *arrivingTrain})
	}
}

func (client *Client) writeJSON(msg Message) {
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
	log.Printf("JSON: %v\n bytes written: %v\n", msg.Message, l)
}

//This is probably really over kill but this lets you write special
//parsers/sorters serverside and apply them to multiple clients at once.
//Eventaully use an enum/constant for the sort configurations/functions.
func (client *Client) configureSort() {
	if client.config.sort == "ascend" {
		client.config.funct = sort
	}
}
