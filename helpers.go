package main

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func convertToTrainSliceAndParse(stopTimeUpdateSlice []*StopTimeUpdate) ([]*Train, ParsedByDirection) {
	unparsed := make([]*Train, 0)
	parsed := ParsedByDirection{Northbound: make([]*Train, 0), SouthBound: make([]*Train, 0)}
	for _, trip := range stopTimeUpdateSlice {
		train := &Train{}
		train.Train = trip
		train.Train.AddDelay()
		train.Train.ConvertArrivalNoDelay()
		train.Train.ConvertArrivalWithDelay()
		train.Train.ConvertTimeToMinutesNoDelay()
		train.Train.ConvertTimeToMinutesWithDelay()
		//train.Train.ConvertDeparture()

		if train.Train.TimeInMinutes < 0 {
			//Sometimes time update data is stale so we skip any times that are in the past
			log.Printf("NEGATIVE TIME IN MINUTES: %v\n", train.Train.ConvertedArrivalTimeNoDelay)
			continue
		}

		idSplit := strings.Split(trip.Id, "")
		direction := strings.ToLower(idSplit[len(idSplit)-1])

		//Create helper for this to parse Northbound & Southbound trains
		if direction == "n" {
			train.Direction = "Manhattan"
			train.DirectionV2 = "N" //Use an Enum for this?
			parsed.Northbound = append(parsed.Northbound, train)
		} else if direction == "s" {
			train.Direction = "Brooklyn"
			train.DirectionV2 = "S" //Use an Enum for this?
			parsed.SouthBound = append(parsed.SouthBound, train)
		}

		unparsed = append(unparsed, train)
	}

	return unparsed, parsed
}

//Get rid boolean return value
func findStopData(update *gtfs.TripUpdate_StopTimeUpdate, stopID string) (bool, *StopTimeUpdate) {
	match := false
	stopTimeUpdate := StopTimeUpdate{}
	if strings.Contains(update.GetStopId(), stopID) {
		match = true

		stopTimeUpdate.Id = update.GetStopId()
		stopTimeUpdate.ArrivalTime = update.GetArrival().Time
		//stopTimeUpdate.DepartureTime = update.GetDeparture().Time
		//stopTimeUpdate.GtfsDeparture = update.GetDeparture()
		if update.Arrival.Delay != nil {
			stopTimeUpdate.Delay = *update.GetArrival().Delay
		}
	}

	return match, &stopTimeUpdate
}

func defaultSort(parsed ParsedByDirection) ParsedByDirection {
	log.Println("DEFAULT SORT", len(parsed.Northbound))

	sort.SliceStable(parsed.Northbound, func(i, j int) bool {
		return parsed.Northbound[i].Train.TimeInMinutes < parsed.Northbound[j].Train.TimeInMinutes
	})

	sort.SliceStable(parsed.SouthBound, func(i, j int) bool {
		return parsed.SouthBound[i].Train.TimeInMinutes < parsed.SouthBound[j].Train.TimeInMinutes
	})

	return parsed
}

func descendingSort(parsed ParsedByDirection) ParsedByDirection {
	log.Println("DESCENDING SORT", time.Now())
	return parsed
}

func testGen(parsed ParsedByDirection) ParsedByDirection {
	return parsed
}
