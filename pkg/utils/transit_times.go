package utils

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	proto "github.com/golang/protobuf/proto"
)

func FetchTransitData(subwayLine string) []*gtfs.TripUpdate {
	client := &http.Client{}
	tripUpdates := make([]*gtfs.TripUpdate, 0)

	reqURL := SUBWAY_LINE_REQUEST_URLS[subwayLine]
	req, err := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("x-api-key", MTA_API_KEY)
	if err != nil {
		log.Default().Println("Error fetching transit data: ", err)
		return tripUpdates
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Default().Println(err)
		return tripUpdates
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Default().Println(err)
		return tripUpdates
	}

	return parseResponse(body, tripUpdates)
}

func parseResponse(body []byte, tripUpdates []*gtfs.TripUpdate) []*gtfs.TripUpdate {
	feed := gtfs.FeedMessage{}

	err := proto.Unmarshal(body, &feed)
	if err != nil {
		log.Default().Println("Error parsing StopTimeUpdate: ", err)
		return tripUpdates
	}

	for _, entity := range feed.Entity {
		tripUpdate := entity.TripUpdate
		if tripUpdate != nil {
			tripUpdates = append(tripUpdates, tripUpdate)
		}
	}
	return tripUpdates
}
