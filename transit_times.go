package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	proto "github.com/golang/protobuf/proto"
)

func handleFetchTransitData(subwayLine string) []*gtfs.TripUpdate_StopTimeUpdate {
	client := &http.Client{}

	alertReqURL := SUBWAY_LINE_REQUEST_URLS["SERVICE"]
	req, err := http.NewRequest("GET", alertReqURL, nil)

	req.Header.Set("x-api-key", MTA_API_KEY)
	if err != nil {
		log.Println("LINE 104")
		log.Default().Println(err)
		return make([]*gtfs.TripUpdate_StopTimeUpdate, 0)
	}

	reqURL := SUBWAY_LINE_REQUEST_URLS[subwayLine]
	req2, err2 := http.NewRequest("GET", reqURL, nil)

	req2.Header.Set("x-api-key", MTA_API_KEY)
	if err2 != nil {
		log.Println("LINE 113")
		log.Default().Println(err2)
		return make([]*gtfs.TripUpdate_StopTimeUpdate, 0)
	}

	return fetch(client, req2)
}

func fetch(client *http.Client, req *http.Request) []*gtfs.TripUpdate_StopTimeUpdate {
	stopTimeUpdateSlice := make([]*gtfs.TripUpdate_StopTimeUpdate, 0)
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

	//log.Println(stopTimeUpdateSlice)
	return stopTimeUpdateSlice
}
