package main

import (
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
}

type Client struct {
	UUID       uuid.UUID
	pool       *Pool
	conn       *websocket.Conn
	send       chan []*gtfs.TripUpdate_StopTimeUpdate
	stopId     string
	subwayLine string
	config     Config
	fetching   bool
}

type Config struct {
	stopId     string
	subwayLine string
	sort       string
	generate   string
	funct      func(parsed ParsedByDirection) ParsedByDirection
	generator  func(parsed ParsedByDirection) ParsedByDirection
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
	Trip                          *gtfs.TripDescriptor `json:"trip"`
	Id                            string               `json:"id"`
	ArrivalTime                   *int64               `json:"arrivalTime"`
	DepartureTime                 *int64               `json:"departureTime"`
	Delay                         int32                `json:"delay"`
	ArrivalTimeWithDelay          int64                `json:"arrivalTimeDelay"`
	ConvertedArrivalTimeWithDelay time.Time            `json:"convertedArrivalTimeWithDelay "`
	ConvertedArrivalTimeNoDelay   time.Time            `json:"convertedArrivalTimeNoDelay"`
	//ConvertedDepartureTime        time.Time                      `json:"convertedDepartureTime"`
	TimeInMinutes        float64                        `json:"timeInMinutes"`
	TimeInMinutesNoDelay float64                        `json:"timeInMinutesNoDelay"`
	GtfsDeparture        *gtfs.TripUpdate_StopTimeEvent `json:"departure"`
}

//Still unsure about how all these time/delay conversions
//should be handled. Merge all of these into 1 conversion function
func (s *StopTimeUpdate) ConvertArrivalNoDelay() {
	s.ConvertedArrivalTimeNoDelay = time.Unix(int64(*s.ArrivalTime), 0)
}

func (s *StopTimeUpdate) ConvertTimeToMinutesWithDelay() {
	s.TimeInMinutes = math.Floor(time.Until(s.ConvertedArrivalTimeWithDelay).Minutes()) + 1
}

func (s *StopTimeUpdate) ConvertArrivalWithDelay() {
	s.ConvertedArrivalTimeWithDelay = time.Unix((s.ArrivalTimeWithDelay), 0)
}

//func (s *StopTimeUpdate) ConvertDeparture() {
//	s.ConvertedDepartureTime = time.Unix(int64(*s.DepartureTime+int64(s.Delay)), 0)
//}

func (s *StopTimeUpdate) ConvertTimeToMinutesNoDelay() {
	s.TimeInMinutesNoDelay = math.Floor(time.Until(s.ConvertedArrivalTimeNoDelay).Minutes()) + 1
}

func (s *StopTimeUpdate) AddDelay() {
	s.ArrivalTimeWithDelay = *s.ArrivalTime + int64(s.Delay)
}

type ArrivingTrain struct {
	ClientID     uuid.UUID         `json:"clientId"`
	SubwayLine   string            `json:"subwayLine"`
	Trains       []*Train          `json:"trains"` //Return all trains to do whatever clientside
	ParsedTrains ParsedByDirection `json:"parsedTrains"`
}

type Train struct {
	DirectionV2 string          `json:"directionV2"`
	Direction   string          `json:"direction"`
	Train       *StopTimeUpdate `json:"train"`
}

type ParsedByDirection struct {
	Northbound []*Train `json:"northbound"` //sorted by the default sorting
	SouthBound []*Train `json:"southbound"` //sorted by the default sorting
	//Add ability to attach a custom data type here so I can
	//use the config struct to write functions that can combine
	//different data feeds into a single return object.
}

type ServiceAlertHeader struct{}

type Resp struct {
	Header struct {
		GtfsRealtimeVersion              string `json:"gtfs_realtime_version"`
		Incrementality                   string `json:"incrementality"`
		Timestamp                        int    `json:"timestamp"`
		TransitRealtimeMercuryFeedHeader struct {
			MercuryVersion string `json:"mercury_version"`
		} `json:"transit_realtime.mercury_feed_header"`
	} `json:"header"`
	Entity []struct {
		ID    string `json:"id"`
		Alert struct {
			ActivePeriod []struct {
				Start int `json:"start"`
			} `json:"active_period"`
			InformedEntity []struct {
				AgencyID                             string `json:"agency_id"`
				RouteID                              string `json:"route_id"`
				TransitRealtimeMercuryEntitySelector struct {
					SortOrder string `json:"sort_order"`
				} `json:"transit_realtime.mercury_entity_selector"`
			} `json:"informed_entity"`
			HeaderText struct {
				Translation []struct {
					Text     string `json:"text"`
					Language string `json:"language"`
				} `json:"translation"`
			} `json:"header_text"`
			DescriptionText struct {
				Translation []struct {
					Text     string `json:"text"`
					Language string `json:"language"`
				} `json:"translation"`
			} `json:"description_text"`
			TransitRealtimeMercuryAlert struct {
				CreatedAt           int    `json:"created_at"`
				UpdatedAt           int    `json:"updated_at"`
				AlertType           string `json:"alert_type"`
				DisplayBeforeActive int    `json:"display_before_active"`
			} `json:"transit_realtime.mercury_alert"`
		} `json:"alert,omitempty"`
	} `json:"entity"`
	Arriving ArrivingTrain `json:"arriving"`
}
