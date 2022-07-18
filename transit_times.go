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
)

//ArrivingTrain
func transitTimes(subwayLine string) []*gtfs.TripUpdate_StopTimeUpdate {
	client := &http.Client{}
	//var arrivingTrain = &ArrivingTrain{}
	var stopTimeUpdateSlice []*gtfs.TripUpdate_StopTimeUpdate

	alertReqURL := SUBWAY_LINE_REQUEST_URLS["SERVICE"]
	req, err := http.NewRequest("GET", alertReqURL, nil)

	req.Header.Set("x-api-key", MTA_API_KEY)
	if err != nil {
		log.Println("LINE 104")
		log.Default().Println(err)
		return stopTimeUpdateSlice
	}

	reqURL := SUBWAY_LINE_REQUEST_URLS[subwayLine]
	req2, err2 := http.NewRequest("GET", reqURL, nil)

	req2.Header.Set("x-api-key", MTA_API_KEY)
	if err2 != nil {
		log.Println("LINE 113")
		log.Default().Println(err2)
		return stopTimeUpdateSlice
	}

	//arrivingTrain.ClientID = id
	//arrivingTrain.SubwayLine = subwayLine
	//arrivingTrain.Trains = fetch(client, req2)
	return fetch(client, req2) //*arrivingTrain
}
//[]*Train
func fetch(client *http.Client, req *http.Request) []*gtfs.TripUpdate_StopTimeUpdate {
	var stopTimeUpdateSlice []*gtfs.TripUpdate_StopTimeUpdate //[]*StopTimeUpdate
	resp, err := client.Do(req)

	if err != nil {
		log.Println("LINE 128")
		log.Default().Println(err)
		return stopTimeUpdateSlice
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("LINE 137")
		log.Default().Println(err)
		return stopTimeUpdateSlice
	}

	feed := gtfs.FeedMessage{}

	err = proto.Unmarshal(body, &feed)
	if err != nil {
		log.Println("LINE 145")
		log.Default().Println(err)
		return stopTimeUpdateSlice
	}

	for _, entity := range feed.Entity {
		tripUpdate := entity.TripUpdate
		if tripUpdate != nil {
			stopTimeUpdateSlice = append(stopTimeUpdateSlice, tripUpdate.GetStopTimeUpdate()...)
		}
	}

	return stopTimeUpdateSlice
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
