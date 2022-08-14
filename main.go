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
		group := r.URL.Query()["group"][0]
		subway := r.URL.Query()["subway"][0]
		stopId := r.URL.Query()["stopId"][0]

		log.Println("group", group)
		if Pools.Map[group] == nil {
			Pools.createPool(group, StaticData.SubwayTripMap)
			Pools.insertIntoPool(group, subway, stopId, conn)
		} else {
			Pools.insertIntoPool(group, subway, stopId, conn)
		}
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("TRASNIT DATE")
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		stopId := r.URL.Query()["stopId"][0]
		group := r.URL.Query()["group"][0]
		subway := r.URL.Query()["subway"][0]

		stopTimeUpdateSlice := make([]*StopTimeUpdate, 0)

		data := handleFetchTransitData(group)
		log.Println(len(data))
		for _, tripUpdate := range data {
			stopTimeUpdate := findStopData(tripUpdate, stopId)
			if stopTimeUpdate != nil {
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

		unparsed, parsed := convertToTrainSliceAndParse(stopTimeUpdateSlice, subway, StaticData.SubwayTripMap)
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
