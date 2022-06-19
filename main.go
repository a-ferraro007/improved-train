package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func main() {
	pool := newPool()
	go pool.run()

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request){
		log.Println("TRASNIT DATE")
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions{
			log.Println("PREFLIGHT", r)
		} else {
			log.Println(r.Method)
		}
		data := transitTimes("L", "L12", uuid.New())
		//log.Printf("%+v",data)
		json, _ := json.Marshal(data)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(w, "Hello, i'm a golang microservice")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("REQUS", r)
		serveWs(pool, w, r)
	})

	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}