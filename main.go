package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var Pools PoolMap

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Println("MTA SERVER v0.1.7")
	Pools.Init()
	StaticData := Process() //process once when the server starts up

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ERROR UPGRADING WEBSOCKET: %v", err)
			return
		}

		subwayLine := r.URL.Query()["subwayLine"][0]
		stopId := r.URL.Query()["stopId"][0]

		if Pools.Map[subwayLine] == nil {
			Pools.createPool(subwayLine, StaticData.SubwayTripMap)
			Pools.insertIntoPool(subwayLine, stopId, conn)
		} else {
			Pools.insertIntoPool(subwayLine, stopId, conn)
		}
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("TRASNIT DATE")
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		stopId := r.URL.Query()["stopId"][0]
		subwayLine := r.URL.Query()["subwayLine"][0]
		stopTimeUpdateSlice := make([]*StopTimeUpdate, 0)

		data := handleFetchTransitData(subwayLine)
		log.Println(len(data))
		for _, tripUpdate := range data {
			match, stopTimeUpdate := findStopData(tripUpdate, stopId)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		log.Println(stopTimeUpdateSlice)
		if len(stopTimeUpdateSlice) <= 0 {
			log.Println("LESS")
			json, _ := json.Marshal("empty")
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
		}

		unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice, subwayLine, StaticData.SubwayTripMap)
		trains := unparsed
		parsedTrains := defaultSort(parsed)

		m := Message{Message: ArrivingTrain{
			Trains:       trains,
			ParsedTrains: parsedTrains,
		}}

		json, _ := json.Marshal(m.Message)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)

	})

	http.HandleFunc("/static", func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		json, _ := json.Marshal(StaticData)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
