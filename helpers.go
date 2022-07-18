package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func returnTrainSlice(stopTimeUpdateSlice []*StopTimeUpdate) []*Train {
	trainSlice := make([]*Train, 0)

	for _, trip := range stopTimeUpdateSlice {
		var train = &Train{}

		//Create helper for this to parse Northbound & Southbound trains
		if strings.Count(trip.Id, "N") >= 1 && strings.Count(trip.Id, "N") <= 2 {
			train.Train = trip
			train.Direction = "Manhattan"
			train.Train.AddDelay()
			train.Train.ConvertArrival()
			train.Train.ConvertDeparture()
			train.Train.ConvertTimeInMinutes()
		}

		if strings.Count(trip.Id, "S") >= 1 && strings.Count(trip.Id, "S") <= 2 {
			train.Train = trip
			train.Direction = "Brooklyn"
			train.Train.AddDelay()
			train.Train.ConvertArrival()
			train.Train.ConvertDeparture()
			train.Train.ConvertTimeInMinutes()
		}

		if train.Train.TimeInMinutes >= 0 {
			trainSlice = append(trainSlice, train)
		} else {
			log.Printf("NEGATIVE TIME IN MINUTES: %v\n", train.Train.ConvertedArrivalTime)
		}
	}
	//log.Printf("TRAIN SLICE: %v\n", trainSlice)
	return trainSlice
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
