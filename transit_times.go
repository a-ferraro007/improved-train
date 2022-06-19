package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
)

func transitTimes(subwayLine string, stopId string, id uuid.UUID) ArrivingTrain {
	client := &http.Client{}
	var arrivingTrain = &ArrivingTrain{}

	alertReqURL := SUBWAY_LINE_REQUEST_URLS["SERVICE"]
	req, err := http.NewRequest("GET", alertReqURL, nil)

	req.Header.Set("x-api-key", MTA_API_KEY)
	if err != nil {
		log.Println("LINE 104")
		log.Default().Println(err)
		return *arrivingTrain
	}

	reqURL := SUBWAY_LINE_REQUEST_URLS[subwayLine]
	req2, err2 := http.NewRequest("GET", reqURL, nil)

	req2.Header.Set("x-api-key", MTA_API_KEY)
	if err2 != nil {
		log.Println("LINE 113")
		log.Default().Println(err2)
		return *arrivingTrain
	}

	arrivingTrain.ClientID = id
	arrivingTrain.SubwayLine = subwayLine
	arrivingTrain.Trains = fetch(client, req2, stopId)
	return *arrivingTrain
}

func fetch(client *http.Client, req *http.Request, stopId string) []*Train {
	var stopTimeUpdateSlice []*StopTimeUpdate
	trainSlice := make([]*Train, 0)
	resp, err := client.Do(req)

	if err != nil {
		log.Println("LINE 128")
		log.Default().Println(err)
		return trainSlice
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("LINE 137")
		log.Default().Println(err)
		return trainSlice
	}

	feed := gtfs.FeedMessage{}

	err = proto.Unmarshal(body, &feed)
	if err != nil {
		log.Println("LINE 145")
		log.Default().Println(err)
		return trainSlice
	}

	for _, entity := range feed.Entity {
		tripUpdate := entity.TripUpdate
		if tripUpdate != nil {
			for _, tripUpdate := range tripUpdate.GetStopTimeUpdate() {
			 match, stopTimeUpdate := findStopData(tripUpdate, stopId)
				if match {
					stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
				}
			}
		}
	}

	if len(stopTimeUpdateSlice) != 0 {
		//log.Println(len(stopTimeUpdateSlice))
		for _, trip := range stopTimeUpdateSlice {
			var train = &Train{}

			if (strings.Count(trip.Id, "N") >= 1 && strings.Count(trip.Id, "N") <= 2) {
				train.Train = trip
				train.Direction = "Manhattan"
				train.Train.AddDelay()
				train.Train.ConvertArrival()
				train.Train.ConvertDeparture()
				train.Train.ConvertTimeInMinutes()
			}

			if (strings.Count(trip.Id, "S") >= 1 && strings.Count(trip.Id, "S") <= 2) {
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
				log.Println(train.Train.TimeInMinutes)
			}
		}
	}
	return trainSlice
}

func findStopData(update *gtfs.TripUpdate_StopTimeUpdate, stopID string) (bool, *StopTimeUpdate) {
	match := false
	stopTimeUpdate := StopTimeUpdate{}

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

func getServiceError(client *http.Client, req *http.Request)  {
	resp, err := client.Do(req)
	if err != nil {

		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := Resp{}
	json.Unmarshal(body, &result)

	for _, entity := range result.Entity {
		for _, informedEntity  := range entity.Alert.InformedEntity {
			if strings.Contains(informedEntity.RouteID, "L") {
				fmt.Println(informedEntity.RouteID,   " " ,  entity.Alert.TransitRealtimeMercuryAlert.AlertType)
			}
		}
	}
}