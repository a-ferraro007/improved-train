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

func (client *Client) writeV2(cachedGTFSData *[]*gtfs.TripUpdate_StopTimeUpdate) {
	defer log.Printf("Closing Write Client: %v \n", client.UUID)

	stopTimeUpdateSlice := make([]*StopTimeUpdate, 0)
	arrivingTrain := &ArrivingTrain{ClientID: client.UUID, SubwayLine: client.subwayLine}

	if len(*cachedGTFSData) != 0 {
		log.Printf("CACHE HIT: %v\n", client.UUID)
		for _, tripUpdate := range *cachedGTFSData {
			match, stopTimeUpdate := findStopData(tripUpdate, client.stopId)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			arrivingTrain.Trains = returnTrainSlice(stopTimeUpdateSlice)
		}
		i := 0
		for i < 2 {
			writeJSON(client, Message{Client: client, Message: *arrivingTrain})
			i++
		}
	}

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
			match, stopTimeUpdate := findStopData(tripUpdate, client.stopId)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			arrivingTrain.Trains = returnTrainSlice(stopTimeUpdateSlice)
		}

		writeJSON(client, Message{Client: client, Message: *arrivingTrain})
	}
}

func writeJSON(client *Client, msg Message) {
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
