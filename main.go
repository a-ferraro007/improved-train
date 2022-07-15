package main

import (
	"log"
	"net/http"
)



func main() {
  poolTrainMap := make(map[string]*Pool)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("WEBSOCKET REQUST", r)
		subwayLine := r.URL.Query()["subwayLine"][0]

		if poolTrainMap[subwayLine] == nil {
			log.Println("RUN POOL", r)
			pool := newPool(subwayLine)
			poolTrainMap[subwayLine] = pool
			go pool.run()
			addClientToPool(pool, w, r, &poolTrainMap)
		} else {
			log.Printf("ADD TO POOL %v\n", poolTrainMap)
			pool := poolTrainMap[subwayLine]
			addClientToPool(pool, w, r,  &poolTrainMap)
		}
	})

	http.HandleFunc("/transit", func(w http.ResponseWriter, r *http.Request){
		//log.Println("TRASNIT DATE")
		//(w).Header().Set("Access-Control-Allow-Origin", "*")
		//if r.Method == http.MethodOptions{
		//	log.Println("PREFLIGHT", r)
		//} else {
		//	log.Println(r.Method)
		//}
		//data := transitTimes("L", "L12", uuid.New())
		////log.Printf("%+v",data)
		//json, _ := json.Marshal(data)
		//w.Header().Set("Content-Type", "application/json")
		//w.Write(json)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(w, "Hello, i'm a golang microservice")
	})


	log.Println("Server Running On Port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}