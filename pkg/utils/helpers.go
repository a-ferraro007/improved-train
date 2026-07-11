package utils

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/a-ferraro007/improved-train/pkg/types"
)

// ConvertToTrainSliceAndParse Function
func ConvertToTrainSliceAndParse(stopTimeUpdates []types.CombinedStopTimeUpdate) types.TrainsByDirection {
	trainsByDirection := types.TrainsByDirection{North: make([]types.Train, 0), South: make([]types.Train, 0)}

	for _, stopTime := range stopTimeUpdates {
		train := types.Train{}
		if stopTime.Type == true {
			stopTime.MinimalStopTimeUpdate.ProcessStopTimeUpdate()
			if stopTime.MinimalStopTimeUpdate.SecondsUntilArrival <= 30 && stopTime.MinimalStopTimeUpdate.SecondsUntilArrival > -30 {
				stopTime.MinimalStopTimeUpdate.IsArriving = true
			}
			train.CombinedStopTimeUpdate.MinimalStopTimeUpdate = stopTime.MinimalStopTimeUpdate
		} else {
			stopTime.StopTimeUpdate.ProcessStopTimeUpdate()
			if stopTime.StopTimeUpdate.SecondsUntilArrival <= 30 && stopTime.StopTimeUpdate.SecondsUntilArrival > -30 {
				stopTime.StopTimeUpdate.IsArriving = true
			}
			train.CombinedStopTimeUpdate.StopTimeUpdate = stopTime.StopTimeUpdate
		}
		switch stopTime.Direction {
		case "n":
			train.Direction = "N"
			trainsByDirection.North = append(trainsByDirection.North, train)
		case "s":
			train.Direction = "S"
			trainsByDirection.South = append(trainsByDirection.South, train)
		default:
			log.Default().Println("Error: Direction unknown: ", stopTime.Direction)
		}
	}

	return trainsByDirection
}

// ParseTripUpdate Function
func ParseTripUpdate(trip *gtfs.TripDescriptor, gtfsStopTimeUpdate *gtfs.TripUpdate_StopTimeUpdate, stopID string, minimal bool, headsigns map[string]string) (types.CombinedStopTimeUpdate, error) {

	if gtfsStopTimeUpdate != nil && strings.Contains(gtfsStopTimeUpdate.GetStopId(), stopID) {
		// log.Default().Println("Found Trip for:", gtfsStopTimeUpdate.GetStopId())
		id := gtfsStopTimeUpdate.GetStopId()
		tripId := strings.Split(trip.GetTripId(), "_")[1]
		arrival := gtfsStopTimeUpdate.GetArrival()
		departure := gtfsStopTimeUpdate.GetDeparture()
		direction := strings.ToLower(strings.Split(id, "")[len(strings.Split(id, ""))-1])

		ret := types.CombinedStopTimeUpdate{
			Type:                  minimal,
			Direction:             direction,
			MinimalStopTimeUpdate: &types.MinimalStopTimeUpdate{Headsign: headsigns[tripId]},
			StopTimeUpdate:        &types.StopTimeUpdate{Headsign: headsigns[tripId]},
		}

		if minimal == true {
			if arrival.Time != nil {
				ret.MinimalStopTimeUpdate.ArrivalTime = arrival.GetTime()
			}
			ret.StopTimeUpdate = nil
		} else {
			ret.StopTimeUpdate.ID = id
			ret.StopTimeUpdate.Trip = trip

			if departure != nil {
				if departure.Delay != nil {
					ret.StopTimeUpdate.DepartureDelay.Delay = departure.GetDelay()
				}
				if departure.Uncertainty != nil {
					ret.StopTimeUpdate.DepartureDelay.Uncertainty = departure.GetUncertainty()
				}
				if departure.Time != nil {
					ret.StopTimeUpdate.DepartureTime = departure.GetTime()
				}
			}

			if arrival != nil {
				if arrival.Delay != nil {
					ret.StopTimeUpdate.ArrivalDelay.Delay = arrival.GetDelay()
				}
				if arrival.Uncertainty != nil {
					ret.StopTimeUpdate.ArrivalDelay.Uncertainty = arrival.GetUncertainty()
				}
				if arrival.Time != nil {
					ret.StopTimeUpdate.ArrivalTime = arrival.GetTime()
				}
			}
			ret.MinimalStopTimeUpdate = nil
		}

		return ret, nil
	}

	return types.CombinedStopTimeUpdate{}, errors.New("")
}

// ReturnLimit Function
func ReturnLimit(trainsByDirection types.TrainsByDirection, limit int) types.TrainsByDirection {
	if limit == 0 {
		return trainsByDirection
	}
	north := trainsByDirection.North
	south := trainsByDirection.South

	if limit < len(north) {
		north = north[0:limit]
	}
	if limit < len(south) {
		south = south[0:limit]
	}

	return types.TrainsByDirection{
		North: north,
		South: south,
	}
}

// DefaultSort Function
func DefaultSort(parsed types.TrainsByDirection) types.TrainsByDirection {
	log.Println("Default sort")

	// sort.SliceStable(parsed.North, func(i, j int) bool {
	// 	return parsed.North[i].StopTimeUpdate.ArrivalTimeInMinutesWithDelay < parsed.North[j].StopTimeUpdate.ArrivalTimeInMinutesWithDelay
	// })

	// sort.SliceStable(parsed.South, func(i, j int) bool {
	// 	return parsed.South[i].StopTimeUpdate.ArrivalTimeInMinutesWithDelay < parsed.South[j].StopTimeUpdate.ArrivalTimeInMinutesWithDelay
	// })

	return parsed
}

// DescendingSort Function
func DescendingSort(parsed types.TrainsByDirection) types.TrainsByDirection {
	log.Println("Descending sort", time.Now())
	return parsed
}

// TestGen Function
func TestGen(parsed types.TrainsByDirection) types.TrainsByDirection {
	return parsed
}
