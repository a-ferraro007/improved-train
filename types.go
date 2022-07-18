package main

import (
	"math"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Pool struct {
	subwayLine             string
	clients                map[uuid.UUID]*Client
	broadcast              chan Message //[]byte
	broadcastV2            chan []*gtfs.TripUpdate_StopTimeUpdate
	register               chan *Client
	unregister             chan *Client
	activeTrains           map[string][]map[uuid.UUID]*Client
	activeTrainChannel     chan string
	cachedStopTimeUpdate   map[string]Message
	cachedStopTimeUpdateV2 map[string][]*gtfs.TripUpdate_StopTimeUpdate
}

type Client struct {
	UUID       uuid.UUID
	pool       *Pool
	conn       *websocket.Conn
	send       chan Message //[]byte
	sendV2     chan []*gtfs.TripUpdate_StopTimeUpdate
	stopId     string
	subwayLine string
	fetching   bool
	//poolMap    PoolClient
}

type PoolClient struct {
	pool    *Pool
	poolMap *map[string]*Pool
}

type Message struct {
	Message ArrivingTrain //[string]interface{}
	Client  *Client
}

type RespMsg struct {
	Message map[string]interface{}
}

type StopTimeUpdate struct {
	Id                     string                         `json:"id"`
	ArrivalTime            *int64                         `json:"arrivalTime"`
	DepartureTime          *int64                         `json:"departureTime"`
	Delay                  int32                          `json:"delay"`
	ArrivalTimeWithDelay   int64                          `json:"arrivalTimeDelay"`
	ConvertedArrivalTime   time.Time                      `json:"convertedArrivalTime"`
	ConvertedDepartureTime time.Time                      `json:"convertedDepartureTime"`
	TimeInMinutes          float64                        `json:"timeInMinutes"`
	GtfsDeparture          *gtfs.TripUpdate_StopTimeEvent `json:"departure"`
}

func (s *StopTimeUpdate) ConvertArrival() {
	s.ConvertedArrivalTime = time.Unix(int64(s.ArrivalTimeWithDelay), 0)
}

func (s *StopTimeUpdate) ConvertDeparture() {
	s.ConvertedDepartureTime = time.Unix(int64(*s.DepartureTime+int64(*s.GtfsDeparture.Delay)), 0)
}

func (s *StopTimeUpdate) AddDelay() {
	s.ArrivalTimeWithDelay = *s.ArrivalTime + int64(s.Delay)
}

func (s *StopTimeUpdate) ConvertTimeInMinutes() {
	s.TimeInMinutes = math.Round(time.Until(s.ConvertedArrivalTime).Minutes())
}

type ArrivingTrain struct {
	ClientID   uuid.UUID `json:"clientId"`
	SubwayLine string    `json:"subwayLine"`
	Trains     []*Train  `json:"trains"`
}

type Train struct {
	Direction string          `json:"direction"`
	Train     *StopTimeUpdate `json:"train"`
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
