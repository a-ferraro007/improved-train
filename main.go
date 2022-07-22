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
	Pools.Init()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ERROR UPGRADING WEBSOCKET: %v", err)
			return
		}

		subwayLine := r.URL.Query()["subwayLine"][0]
		stopId := r.URL.Query()["stopId"][0]

		if Pools.Map[subwayLine] == nil {
			Pools.createPool(subwayLine)
			Pools.insertIntoPool(subwayLine, stopId, conn)
		} else {
			Pools.insertIntoPool(subwayLine, stopId, conn)
		}
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request) {
		log.Println("TRASNIT DATE")
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			log.Println("PREFLIGHT", r)
		} else {
			log.Println(r.Method)
		}
		data := transitTimes("L")
		json, _ := json.Marshal(data)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(w, "Hello, i'm a golang microservice")
	})

	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
