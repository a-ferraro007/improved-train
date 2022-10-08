package utils

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/a-ferraro007/improved-train/pkg/types"
)

//ConvertToTrainSliceAndParse Function
func ConvertToTrainSliceAndParse(stopTimeUpdateSlice []*types.StopTimeUpdate) ([]*types.Train, types.ParsedByDirection) {
	unparsed := make([]*types.Train, 0)
	parsed := types.ParsedByDirection{Northbound: make([]*types.Train, 0), SouthBound: make([]*types.Train, 0)}
	for _, trip := range stopTimeUpdateSlice {
		train := &types.Train{}
		train.Train = trip
		train.Train.AddDelay()
		train.Train.ConvertArrivalNoDelay()
		train.Train.ConvertArrivalWithDelay()
		train.Train.ConvertTimeToMinutesNoDelay()
		train.Train.ConvertTimeToMinutesWithDelay()
		train.Train.ConvertDeparture()

		if train.Train.TimeInMinutes < 0 {
			//Sometimes time update data is stale so we skip any times that are in the past
			log.Printf("NEGATIVE TIME IN MINUTES: %v\n", train.Train.ConvertedArrivalTimeNoDelay)
			continue
		}

		idSplit := strings.Split(trip.ID, "")
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

//FindStopData Function
func FindStopData(update *gtfs.TripUpdate_StopTimeUpdate, stopID string) (bool, *types.StopTimeUpdate) {
	match := false
	stopTimeUpdate := types.StopTimeUpdate{}
	if strings.Contains(update.GetStopId(), stopID) {
		match = true
		stopTimeUpdate.ID = update.GetStopId()
		stopTimeUpdate.ArrivalTime = update.GetArrival().Time
		stopTimeUpdate.DepartureTime = update.GetDeparture().Time
		stopTimeUpdate.GtfsDeparture = update.GetDeparture()
		if update.Arrival.Delay != nil {
			stopTimeUpdate.Delay = *update.GetArrival().Delay
		}
	}

	return match, &stopTimeUpdate
}

//DefaultSort Function
func DefaultSort(parsed types.ParsedByDirection) types.ParsedByDirection {
	log.Println("DEFAULT SORT", len(parsed.Northbound))

	sort.SliceStable(parsed.Northbound, func(i, j int) bool {
		return parsed.Northbound[i].Train.TimeInMinutes < parsed.Northbound[j].Train.TimeInMinutes
	})

	sort.SliceStable(parsed.SouthBound, func(i, j int) bool {
		return parsed.SouthBound[i].Train.TimeInMinutes < parsed.SouthBound[j].Train.TimeInMinutes
	})

	return parsed
}

//DescendingSort Function
func DescendingSort(parsed types.ParsedByDirection) types.ParsedByDirection {
	log.Println("DESCENDING SORT", time.Now())
	return parsed
}

//TestGen Function
func TestGen(parsed types.ParsedByDirection) types.ParsedByDirection {
	return parsed
}
