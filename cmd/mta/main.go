package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/a-ferraro007/improved-train/pkg/clientpool"
	"github.com/a-ferraro007/improved-train/pkg/types"
	"github.com/a-ferraro007/improved-train/pkg/utils"
	"github.com/gorilla/websocket"
)

//Pools is a map that holds all of the running pools organized by subwaygroup
//var Pools clientpool.PoolMap

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Println("MTA SERVER v0.1.4")
	clientpool.Init()
	stations := utils.Process() //process once when the server starts up

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ERROR UPGRADING WEBSOCKET: %v", err)
			return
		}

		subwayLine := r.URL.Query()["subwayLine"][0]
		stopID := r.URL.Query()["stopID"][0]

		clientpool.HandleNewConnection(subwayLine, stopID, conn)
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("TRASNIT DATE")
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		stopID := r.URL.Query()["stopID"][0]
		subwayLine := r.URL.Query()["subwayLine"][0]
		stopTimeUpdateSlice := make([]*types.StopTimeUpdate, 0)

		data := utils.HandleFetchTransitData(subwayLine)
		log.Println(len(data))
		for _, tripUpdate := range data {
			match, stopTimeUpdate := utils.FindStopData(tripUpdate, stopID)
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

		unparsed, parsed := utils.ConvertToTrainSliceAndParse(stopTimeUpdateSlice)
		trains := unparsed
		parsedTrains := utils.DefaultSort(parsed)

		m := clientpool.Message{Message: types.UpcomingTrain{
			Trains:       trains,
			ParsedTrains: parsedTrains,
		}}

		json, _ := json.Marshal(m.Message)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)

	})

	http.HandleFunc("/stations", func(w http.ResponseWriter, r *http.Request) {
		log.Println(w, "Hello, i'm a golang microservice")
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		json, _ := json.Marshal(stations)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
