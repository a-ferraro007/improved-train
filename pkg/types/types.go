package types

import (
	"math"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
)

//Config struct holds all client related data
type Config struct {
	StopID     string
	SubwayLine string
	Sort       string
	Generate   string
	Funct      func(parsed ParsedByDirection) ParsedByDirection
	Generator  func(parsed ParsedByDirection) ParsedByDirection
	//use this generator property to keep custom property generators
	//seperate of the sorting function config property.
}

//SortPrototype struct
type SortPrototype func(parsed ParsedByDirection) ParsedByDirection

//RespMsg struct
type RespMsg struct {
	Message map[string]interface{}
}

//StopTimeUpdate struct
type StopTimeUpdate struct {
	Trip                          *gtfs.TripDescriptor           `json:"trip"`
	ID                            string                         `json:"id"`
	ArrivalTime                   *int64                         `json:"arrivalTime"`
	DepartureTime                 *int64                         `json:"departureTime"`
	Delay                         int32                          `json:"delay"`
	ArrivalTimeWithDelay          int64                          `json:"arrivalTimeDelay"`
	ConvertedArrivalTimeWithDelay time.Time                      `json:"convertedArrivalTimeWithDelay "`
	ConvertedArrivalTimeNoDelay   time.Time                      `json:"convertedArrivalTimeNoDelay"`
	ConvertedDepartureTime        time.Time                      `json:"convertedDepartureTime"`
	TimeInMinutes                 float64                        `json:"timeInMinutes"`
	TimeInMinutesNoDelay          float64                        `json:"timeInMinutesNoDelay"`
	GtfsDeparture                 *gtfs.TripUpdate_StopTimeEvent `json:"departure"`
}

//Still unsure about how all these time/delay conversions
//should be handled. Merge all of these into 1 conversion function

//ConvertArrivalNoDelay Func
func (s *StopTimeUpdate) ConvertArrivalNoDelay() {
	s.ConvertedArrivalTimeNoDelay = time.Unix(int64(*s.ArrivalTime), 0)
}

//ConvertTimeToMinutesWithDelay Func
func (s *StopTimeUpdate) ConvertTimeToMinutesWithDelay() {
	s.TimeInMinutes = math.Floor(time.Until(s.ConvertedArrivalTimeWithDelay).Minutes()) + 1
}

//ConvertArrivalWithDelay Func
func (s *StopTimeUpdate) ConvertArrivalWithDelay() {
	s.ConvertedArrivalTimeWithDelay = time.Unix((s.ArrivalTimeWithDelay), 0)
}

//ConvertDeparture Func
func (s *StopTimeUpdate) ConvertDeparture() {
	s.ConvertedDepartureTime = time.Unix(int64(*s.DepartureTime+int64(s.Delay)), 0)
}

//ConvertTimeToMinutesNoDelay Func
func (s *StopTimeUpdate) ConvertTimeToMinutesNoDelay() {
	s.TimeInMinutesNoDelay = math.Floor(time.Until(s.ConvertedArrivalTimeNoDelay).Minutes()) + 1
}

//AddDelay Func
func (s *StopTimeUpdate) AddDelay() {
	s.ArrivalTimeWithDelay = *s.ArrivalTime + int64(s.Delay)
}

//UpcomingTrain struct
type UpcomingTrain struct {
	ClientID     uuid.UUID         `json:"clientId"`
	SubwayLine   string            `json:"subwayLine"`
	Trains       []*Train          `json:"trains"` //Return all trains to do whatever clientside
	ParsedTrains ParsedByDirection `json:"parsedTrains"`
}

//Train Struct
type Train struct {
	DirectionV2 string          `json:"directionV2"`
	Direction   string          `json:"direction"`
	Train       *StopTimeUpdate `json:"train"`
}

//ParsedByDirection Struct
type ParsedByDirection struct {
	Northbound []*Train `json:"northbound"` //sorted by the default sorting
	SouthBound []*Train `json:"southbound"` //sorted by the default sorting
	//Add ability to attach a custom data type here so I can
	//use the config struct to write functions that can combine
	//different data feeds into a single return object.
}

type serviceAlertHeader struct{}
