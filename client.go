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
	arrivingTrain := &ArrivingTrain{ClientID: client.UUID, SubwayLine: client.config.group}

	if len(*cachedGTFSData) != 0 {
		log.Printf("WRITE FROM CACHE: %v\n", client.UUID)
		for _, tripUpdate := range *cachedGTFSData {
			stopTimeUpdate := findStopData(tripUpdate, client.config.stopId)
			if stopTimeUpdate != nil {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		//log.Println(client.)
		if len(stopTimeUpdateSlice) > 0 {
			unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice, client.config.subway, client.pool.TripHeadSignMap)
			arrivingTrain.Trains = unparsed
			arrivingTrain.ParsedTrains = client.config.funct(parsed)
		}

		i := 0

		for i < 2 {
			client.writeJSON(Message{Client: client, Message: *arrivingTrain})
			i++
		}
	}
	log.Printf("START WRITING FOR CLIENT: %v \n", client.UUID)
	for {
		data, ok := <-client.send
		if !ok {
			client.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		//not sure if rezeroing the slices is actually necessary
		arrivingTrain.Trains = make([]*Train, 0)
		arrivingTrain.ParsedTrains.Northbound = make([]*Train, 0)
		arrivingTrain.ParsedTrains.SouthBound = make([]*Train, 0)
		stopTimeUpdateSlice = make([]*StopTimeUpdate, 0)

		for _, tripUpdate := range data {
			stopTimeUpdate := findStopData(tripUpdate, client.config.stopId)
			if stopTimeUpdate != nil {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice, client.config.subway, client.pool.TripHeadSignMap)
			arrivingTrain.Trains = unparsed
			arrivingTrain.ParsedTrains = client.config.funct(parsed)
		}
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
//Can probably use this to return custom data types i.e. mixing train times
// and service alert information.
func (client *Client) configureSort() {
	switch client.config.sort {
	case "descending":
		client.config.funct = descendingSort
	default:
		client.config.funct = defaultSort
	}
}

func (client *Client) configureGenerator() {
	switch client.config.generate {
	case "test":
		client.config.generator = testGen
	default:
		client.config.generator = nil
	}
}
