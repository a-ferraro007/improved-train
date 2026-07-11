package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/a-ferraro007/improved-train/pkg/clientpool"
	"github.com/a-ferraro007/improved-train/pkg/types"
	"github.com/a-ferraro007/improved-train/pkg/utils"
	"github.com/gorilla/websocket"
)

// var Pools clientpool.PoolMap
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Println("Train Time Server v0.3.0")
	m := utils.CreateHeadSignsMap()
	clientpool.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Default().Println("Error upgrading http connection: ", err)
			return
		}

		subwayLine := r.URL.Query().Get("subwayLine")
		stopID := r.URL.Query().Get("stopID")
		min := false
		if r.URL.Query().Get("min") != "" {
			var err error
			min, err = strconv.ParseBool(r.URL.Query().Get("min"))
			if err != nil {
				log.Default().Println(err)
			}
		}

		if stopID == "" {
			log.Default().Println("Missing stopId")
			return
		}

		if subwayLine == "" {
			log.Default().Println("Missing subwayLine")
			return
		}

		limit := 0
		if len(r.URL.Query()["limit"]) > 0 {
			limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
			log.Default().Println("Limit Set: ", limit)
		}

		clientpool.HandleNewConnection(conn, subwayLine, stopID, limit, min, m)
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")

		subwayLine := r.URL.Query().Get("subwayLine")
		stopID := r.URL.Query().Get("stopID")
		min := false
		if r.URL.Query().Get("min") != "" {
			var err error
			min, err = strconv.ParseBool(r.URL.Query().Get("min"))
			if err != nil {
				log.Default().Println(err)
			}
		}

		if stopID == "" {
			log.Default().Println("Missing stopId")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":"missing stopId"}`))
			return
		}

		if subwayLine == "" {
			log.Default().Println("Missing subwayLine")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":"missing subwayLine"}`))
			return
		}

		data := utils.FetchTransitData(subwayLine)

		stopTimeUpdates := make([]types.CombinedStopTimeUpdate, 0)
		for _, tripUpdate := range data {
			trip := tripUpdate.GetTrip()
			// log.Default().Println(trip.GetTripId())
			for _, stopTime := range tripUpdate.GetStopTimeUpdate() {
				stopTimeUpdate, err := utils.ParseTripUpdate(trip, stopTime, stopID, min, m)
				if err == nil {
					stopTimeUpdates = append(stopTimeUpdates, stopTimeUpdate)
				}
			}
		}

		if len(stopTimeUpdates) <= 0 {
			empty := []types.StopTimeUpdate{}
			json, _ := json.Marshal(empty)
			w.Header().Set("Content-Type", "application/json")
			w.Write(json)
			return
		}

		limit := 0
		if len(r.URL.Query()["limit"]) > 0 {
			limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
			log.Default().Println("Limit Set: ", limit)
		}

		trainsByDirection := utils.ReturnLimit(utils.DefaultSort(utils.ConvertToTrainSliceAndParse(stopTimeUpdates)), limit)
		json, _ := json.Marshal(trainsByDirection)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)

	})

	// http.HandleFunc("/stations", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Println(w, "Stations Endpoint")
	// 	(w).Header().Set("Access-Control-Allow-Origin", "*")
	// 	json, _ := json.Marshal(stations)
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(json)
	// })

	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
