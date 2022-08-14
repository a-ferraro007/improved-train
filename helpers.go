package main

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func convertToTrainSliceAndParse(stopTimeUpdateSlice []*StopTimeUpdate, subway string, subwayTripMap map[string]TripHeadSign) ([]*Train, ParsedByDirection) {

	log.Println("CONVERT", subway, subwayTripMap[subway])
	unparsed := make([]*Train, 0)
	parsed := ParsedByDirection{Northbound: make([]*Train, 0), SouthBound: make([]*Train, 0)}
	for _, stopTimeUpdate := range stopTimeUpdateSlice {
		train := &Train{}
		train.Train = stopTimeUpdate
		if train.Train.ArrivalTime == nil {
			continue
		}
		location, _ := time.LoadLocation("America/New_York")
		train.Train.ConvertArrival()
		train.Train.ConvertDeparture()
		train.Train.ConvertTimeToMinutes()

		if time.Now().In(location).After(train.Train.ConvertedDepartureTime) {
			//Sometimes time update data is stale so we skip any times that are in the past
			log.Printf("NEGATIVE TIME IN MINUTES: %v\n", train.Train.ConvertedDepartureTime)
			continue
		}
		//There's definitely a direction field somewhere in this data to use instead
		idSplit := strings.Split(stopTimeUpdate.Id, "")
		direction := strings.ToLower(idSplit[len(idSplit)-1])

		//Create helper for this to parse Northbound & Southbound trains
		if direction == "n" {
			train.HeadSign = subwayTripMap[subway].North
			parsed.Northbound = append(parsed.Northbound, train)
		} else if direction == "s" {
			train.HeadSign = subwayTripMap[subway].South
			parsed.SouthBound = append(parsed.SouthBound, train)
		}

		unparsed = append(unparsed, train)
	}

	return unparsed, parsed
}

//Get rid boolean return value
func findStopData(update *gtfs.TripUpdate_StopTimeUpdate, stopID string) *StopTimeUpdate {
	stopTimeUpdate := StopTimeUpdate{}
	if strings.Contains(update.GetStopId(), stopID) {
		stopTimeUpdate.Id = update.GetStopId()

		if update.Arrival != nil {
			stopTimeUpdate.ArrivalTime = update.GetArrival().Time
			if update.GetArrival().Delay != nil {
				stopTimeUpdate.ArrivalDelay = *update.GetArrival().Delay
				*stopTimeUpdate.ArrivalTime += int64(stopTimeUpdate.ArrivalDelay)
			}
		}
		if update.Departure != nil {
			stopTimeUpdate.DepartureTime = update.GetDeparture().Time
			if update.GetDeparture().Delay != nil {
				stopTimeUpdate.DepartureDelay = *update.GetDeparture().Delay
				*stopTimeUpdate.DepartureTime += int64(stopTimeUpdate.DepartureDelay)
			}
		}
	}

	return &stopTimeUpdate
}

func defaultSort(parsed ParsedByDirection) ParsedByDirection {
	log.Println("DEFAULT SORT", len(parsed.Northbound))

	sort.SliceStable(parsed.Northbound, func(i, j int) bool {
		return *parsed.Northbound[i].Train.DepartureTime < *parsed.Northbound[j].Train.DepartureTime
	})

	sort.SliceStable(parsed.SouthBound, func(i, j int) bool {
		return *parsed.SouthBound[i].Train.DepartureTime < *parsed.SouthBound[j].Train.DepartureTime
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
