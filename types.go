package main

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Pool struct {
	subwayLine           string
	clients              map[uuid.UUID]*Client
	broadcast            chan []*gtfs.TripUpdate_StopTimeUpdate
	register             chan *Client
	unregister           chan *Client
	activeTrains         map[string][]map[uuid.UUID]*Client
	activeTrainChannel   chan string
	cachedStopTimeUpdate map[string][]*gtfs.TripUpdate_StopTimeUpdate
	ticker               *time.Ticker
	done                 chan bool
	TripHeadSignMap      map[string]TripHeadSign
}

type Client struct {
	UUID     uuid.UUID
	pool     *Pool
	conn     *websocket.Conn
	send     chan []*gtfs.TripUpdate_StopTimeUpdate
	stopId   string
	group    string
	subway   string
	config   Config
	fetching bool
}

type Config struct {
	stopId    string
	group     string
	subway    string
	sort      string
	generate  string
	funct     func(parsed ParsedByDirection) ParsedByDirection
	generator func(parsed ParsedByDirection) ParsedByDirection
	//use this generator property to keep custom property generators
	//seperate of the sorting function config property.
}
type SortPrototype func(parsed ParsedByDirection) ParsedByDirection

type PoolMap struct {
	Mutex sync.RWMutex
	Map   map[string]*Pool
}

type Message struct {
	Message ArrivingTrain //Confusing naming convention
	Client  *Client
}

type RespMsg struct {
	Message map[string]interface{}
}

type StopTimeUpdate struct {
	Trip                   *gtfs.TripDescriptor `json:"trip"`
	Id                     string               `json:"id"`
	ArrivalTime            *int64               `json:"arrivalTime"`
	DepartureTime          *int64               `json:"departureTime"`
	ArrivalDelay           int32                `json:"arrivalDelay"`
	DepartureDelay         int32                `json:"departureDelay"`
	ConvertedArrivalTime   time.Time            `json:"convertedArrivalTime"`
	ConvertedDepartureTime time.Time            `json:"convertedDepartureTime"`
	TimeInMinutes          float64              `json:"timeInMinutes"`
}

func (s *StopTimeUpdate) ConvertArrival() {
	if s.ArrivalTime == nil {
		log.Println(s.ArrivalTime)
		return
	}
	location, _ := time.LoadLocation("America/New_York")
	s.ConvertedArrivalTime = time.Unix(int64(*s.ArrivalTime), 0).In(location)
}

func (s *StopTimeUpdate) ConvertDeparture() {
	if s.DepartureTime == nil {
		log.Println(s.DepartureTime)
		return
	}
	location, _ := time.LoadLocation("America/New_York")
	s.ConvertedDepartureTime = time.Unix(int64(*s.DepartureTime), 0).In(location)
}

func (s *StopTimeUpdate) ConvertTimeToMinutes() {
	if s.ConvertedDepartureTime == time.Date(0001, 01, 01, 00, 00, 00, +0000, time.Local) {
		log.Println(s.DepartureTime)
		return
	}
	s.TimeInMinutes = math.Floor(time.Until(s.ConvertedDepartureTime).Minutes()) + 1
}

type ArrivingTrain struct {
	ClientID     uuid.UUID         `json:"clientId"`
	SubwayLine   string            `json:"group"`
	Trains       []*Train          `json:"trains"` //Return all trains to do whatever clientside
	ParsedTrains ParsedByDirection `json:"parsedTrains"`
	Direction    string            `json:"direction"`
}

type Train struct {
	HeadSign string          `json:"headsign"`
	Train    *StopTimeUpdate `json:"train"`
}

type ParsedByDirection struct {
	Northbound []*Train `json:"northbound"` //sorted by the default sorting
	SouthBound []*Train `json:"southbound"` //sorted by the default sorting
	//Add ability to attach a custom data type here so I can
	//use the config struct to write functions that can combine
	//different data feeds into a single return object.
}

type BoroughStationMap struct {
	seen              map[string]bool
	Stations          []Station            `json:"stations"`
	StationsByBorough map[string][]Station `json:"stationsByBorough"`
}

type StaticData struct {
	StationMap    SubwayLineMap           `json:"stationMap"`
	SubwayTripMap map[string]TripHeadSign `json:"subwayTripMap"`
}

type TripHeadSign struct {
	North string
	South string
}

type ServiceAlertHeader struct{}

type Station struct {
	StationId      string `json:"stationId"`
	ComplexId      string `json:"complexId"`
	StopId         string `json:"stopId"`
	SubwayLine     string `json:"subwayLine"`
	StopName       string `json:"stopName"`
	Borough        string `json:"borough"`
	Routes         string `json:"routes"`
	Lattitude      string `json:"lattitude"`
	Longitude      string `json:"longitude"`
	NorthDirection string `json:"northDirectionLabel"`
	SouthDirection string `json:"southDirectionLabel"`
}

type SubwayLineMap struct {
	One     *BoroughStationMap
	Two     *BoroughStationMap
	Three   *BoroughStationMap
	Four    *BoroughStationMap
	Five    *BoroughStationMap
	Six     *BoroughStationMap
	Seven   *BoroughStationMap
	A       *BoroughStationMap
	C       *BoroughStationMap
	E       *BoroughStationMap
	B       *BoroughStationMap
	D       *BoroughStationMap
	F       *BoroughStationMap
	M       *BoroughStationMap
	N       *BoroughStationMap
	Q       *BoroughStationMap
	R       *BoroughStationMap
	W       *BoroughStationMap
	L       *BoroughStationMap
	G       *BoroughStationMap
	S       *BoroughStationMap
	J       *BoroughStationMap
	Z       *BoroughStationMap
	SERVICE BoroughStationMap
}
