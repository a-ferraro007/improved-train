package types

import (
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/google/uuid"
)

// Config struct holds all client related data
type Config struct {
	StopID     string
	SubwayLine string
	Sort       string
	Generate   string
	SortFunc   func(parsed TrainsByDirection) TrainsByDirection
	Generator  func(parsed TrainsByDirection) TrainsByDirection
	Limit      int
	Minimal    bool
	Headsigns  map[string]string
	//use this generator property to keep custom property generators
	//seperate of the sorting function config property.
}

// SortPrototype struct
type SortPrototype func(parsed TrainsByDirection) TrainsByDirection

// RespMsg struct
type RespMsg struct {
	Message map[string]interface{}
}

type Delay struct {
	Delay       int32 `json:"delay"`
	Uncertainty int32 `json:"uncertainty"`
}

// StopTimeUpdate struct
type StopTimeUpdate struct {
	ID                     string               `json:"id"`
	Trip                   *gtfs.TripDescriptor `json:"trip"`
	Headsign               string               `json:"headsign"`
	IsArriving             bool                 `json:"isArriving"`
	ArrivalTime            int64                `json:"arrivalTime"`
	DepartureTime          int64                `json:"departureTime"`
	ArrivalDelay           Delay                `json:"arrivalDelay"`
	DepartureDelay         Delay                `json:"departureDelay"`
	ArrivalTimeLocal       time.Time            `json:"arrivalTimeLocal"`
	DepartureTimeLocal     time.Time            `json:"departureTimeLocal"`
	ArrivalTimeInMinutes   float64              `json:"arrivalTimeInMinutes"`
	DepartureTimeInMinutes float64              `json:"departureTimeInMinutes"`
	SecondsUntilArrival    int64                `json:"secondsUntilArrival"`
}

type MinimalStopTimeUpdate struct {
	ArrivalTime          int64
	SecondsUntilArrival  int64
	Headsign             string  `json:"headsign"`
	IsArriving           bool    `json:"isArriving"`
	ArrivalTimeInMinutes float64 `json:"arrivalTimeInMinutes"`
}

type CombinedStopTimeUpdate struct {
	Type                  bool
	Direction             string
	StopTimeUpdate        *StopTimeUpdate
	MinimalStopTimeUpdate *MinimalStopTimeUpdate
}

// ConvertArrivalTimeToLocal Func
func (s *StopTimeUpdate) ConvertArrivalTimeToLocal() {
	s.ArrivalTimeLocal = time.Unix(s.ArrivalTime, 0)
}

// ConvertSecondsUntilArrival Func
func (s *StopTimeUpdate) ConvertSecondsUntilArrival() {
	local := time.Unix(s.ArrivalTime, 0)
	s.SecondsUntilArrival = int64(time.Until(local).Seconds())
}

// ConvertSecondsUntilArrival Func
func (m *MinimalStopTimeUpdate) ConvertSecondsUntilArrival() {
	local := time.Unix(m.ArrivalTime, 0)
	m.SecondsUntilArrival = int64(time.Until(local).Seconds())
}

// ConvertDepartureTimeToLocal Func
func (s *StopTimeUpdate) ConvertDepartureTimeToLocal() {
	s.DepartureTimeLocal = time.Unix(s.DepartureTime, 0)
}

// ConvertArrivalTimeToMinutes Func
func (s *StopTimeUpdate) ConvertArrivalTimeToMinutes() {
	local := time.Unix(s.ArrivalTime, 0)
	seconds := int64(time.Until(local).Seconds())
	s.ArrivalTimeInMinutes = float64((time.Duration(seconds)*time.Second + time.Minute - 1) / time.Minute)
}

// ConvertArrivalTimeToMinutes Func
func (m *MinimalStopTimeUpdate) ConvertArrivalTimeToMinutes() {
	local := time.Unix(m.ArrivalTime, 0)
	seconds := int64(time.Until(local).Seconds())
	m.ArrivalTimeInMinutes = float64((time.Duration(seconds)*time.Second + time.Minute - 1) / time.Minute)
}

// ConvertDepartureTimeToMinutes Func
func (s *StopTimeUpdate) ConvertDepartureTimeToMinutes() {
	d := time.Until(s.DepartureTimeLocal)
	s.DepartureTimeInMinutes = float64((d + time.Minute - 1) / time.Minute)
}

// ProcessStopTimeUpdate
func (s *StopTimeUpdate) ProcessStopTimeUpdate() {
	s.ConvertArrivalTimeToLocal()
	s.ConvertSecondsUntilArrival()
	s.ConvertArrivalTimeToMinutes()
	s.ConvertDepartureTimeToLocal()
	s.ConvertDepartureTimeToMinutes()
}

func (m *MinimalStopTimeUpdate) ProcessStopTimeUpdate() {
	m.ConvertArrivalTimeToMinutes()
	m.ConvertSecondsUntilArrival()
}

// NextTrain struct
type NextTrain struct {
	ClientID          uuid.UUID `json:"clientId"`
	SubwayLine        string    `json:"subwayLine"`
	TrainsByDirection `json:"trainsByDirection"`
}

// Train Struct
type Train struct {
	Type                   int
	Direction              string                 `json:"direction"`
	CombinedStopTimeUpdate CombinedStopTimeUpdate `json:"stopTimeUpdate"`
}

// TrainByDirection Struct
type TrainsByDirection struct {
	North []Train `json:"north"`
	South []Train `json:"south"`
	//Add ability to attach a custom data type here so I can
	//use the config struct to write functions that can combine
	//different data feeds into a single return object.
}

type serviceAlertHeader struct{}
