package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func convertToTrainSliceAndParse(stopTimeUpdateSlice []*StopTimeUpdate) ([]*Train, ParsedByDirection) {
	unparsed := make([]*Train, 0)
	parsed := ParsedByDirection{}
	for _, trip := range stopTimeUpdateSlice {
		train := &Train{}
		train.Train = trip
		train.Train.AddDelay()
		train.Train.ConvertArrival()
		train.Train.ConvertDeparture()
		train.Train.ConvertTimeInMinutes()
		if train.Train.TimeInMinutes < 0 {
			log.Println("neg", train.Train)
			log.Printf("NEGATIVE TIME IN MINUTES: %v\n", train.Train.ConvertedArrivalTime)
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

func findStopData(update *gtfs.TripUpdate_StopTimeUpdate, stopID string) (bool, *StopTimeUpdate) {
	match := false
	stopTimeUpdate := StopTimeUpdate{}
	//log.Printf("CLIENT STOP ID: %v\n UPDATE STOP ID: %v\n", stopID, update.GetStopId())
	if strings.Contains(update.GetStopId(), stopID) {
		match = true

		if update.GetDeparture() != nil {
			stopTimeUpdate.Id = update.GetStopId()
			stopTimeUpdate.ArrivalTime = update.GetArrival().Time
			stopTimeUpdate.DepartureTime = update.GetDeparture().Time
			stopTimeUpdate.Delay = update.GetDeparture().GetDelay()
			stopTimeUpdate.GtfsDeparture = update.GetDeparture()
		} else {
			fmt.Println("NO DEPARTURE")
			stopTimeUpdate.Id = update.GetStopId()
			stopTimeUpdate.ArrivalTime = update.GetArrival().Time
		}
	}

	return match, &stopTimeUpdate
}

func defaultSort(parsed ParsedByDirection) ParsedByDirection {
	log.Println("DEFAULT SORT", time.Now())
	sort.SliceStable(parsed.Northbound, func(i, j int) bool {
		return parsed.Northbound[i].Train.TimeInMinutes < parsed.SouthBound[j].Train.TimeInMinutes
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
