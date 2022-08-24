package clientpool

import (
	"log"
	"sync"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/a-ferraro007/improved-train/pkg/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//PoolMap Struct
type PoolMap struct {
	Mutex sync.RWMutex
	Map   map[string]*Pool
}

//Pools Map
var Pools PoolMap

//Init Function
func Init() {
	log.Println("INITIALIZE POOLS")
	Pools.Map = make(map[string]*Pool)
}

//HandleNewConnection Function
func HandleNewConnection(subwayLine string, stopID string, conn *websocket.Conn) {
	if Pools.Map[subwayLine] == nil {
		createPool(subwayLine)
		insertIntoPool(subwayLine, stopID, conn)
	} else {
		insertIntoPool(subwayLine, stopID, conn)
	}
}

func createPool(subwayLine string) *Pool {
	Pools.Mutex.Lock()
	defer Pools.Mutex.Unlock()

	pool := newPool(subwayLine)
	Pools.Map[subwayLine] = pool
	log.Println(Pools.Map)
	go pool.run()
	go pool.fetchData()
	return pool
}

func insertIntoPool(subwayLine string, stopID string, conn *websocket.Conn) {
	Pools.Mutex.Lock()
	defer Pools.Mutex.Unlock()

	pool := Pools.Map[subwayLine]
	log.Println(Pools.Map)

	client := &Client{
		UUID:       uuid.New(),
		Pool:       pool,
		Conn:       conn,
		Send:       make(chan []*gtfs.TripUpdate_StopTimeUpdate),
		StopID:     stopID,
		SubwayLine: subwayLine,
		Config:     types.Config{StopID: stopID, SubwayLine: subwayLine, Sort: "ascending"},
		Fetching:   false,
	}
	client.ConfigureSort()
	client.ConfigureGenerator()

	mV2 := make([]*gtfs.TripUpdate_StopTimeUpdate, 0)
	mV2 = pool.CachedStopTimeUpdate[client.Config.SubwayLine]
	pool.Register <- client
	go client.read()
	go client.write(&mV2)
}

//DeletePool function
func (p *PoolMap) DeletePool(subwayLine string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.Map[subwayLine].Done <- true
	delete(p.Map, subwayLine)
	log.Println("POOL MAP: ", p.Map)
}
