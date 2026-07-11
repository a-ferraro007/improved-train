package clientpool

import (
	"encoding/json"
	"log"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/a-ferraro007/improved-train/pkg/types"
	"github.com/a-ferraro007/improved-train/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client struct holds all client related data
type Client struct {
	UUID       uuid.UUID
	Pool       *Pool
	Conn       *websocket.Conn
	Send       chan []*gtfs.TripUpdate
	StopID     string
	SubwayLine string
	Config     types.Config
	Fetching   bool
}

// Message Struct
type Message struct {
	Message types.NextTrain
	Client  *Client
}

func (client *Client) read() {
	defer func() {
		log.Default().Println("Closing Read Client: ", client.UUID)
		client.Pool.Unregister <- client
		client.Conn.Close()
		log.Default().Println("Client Closed: ", client.UUID)
	}()

	for {
		m := &types.RespMsg{}

		_, d, readerErr := client.Conn.ReadMessage()
		if readerErr != nil {
			log.Println(readerErr)
			return
		}
		err := json.Unmarshal(d, &m.Message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (client *Client) write(cachedGTFSData *[]*gtfs.TripUpdate) {
	defer log.Default().Println("Closing write for ClientId: ", client.UUID)
	log.Default().Println("Writing to ClientId: ", client.UUID)
	nextTrain := types.NextTrain{ClientID: client.UUID, SubwayLine: client.Config.SubwayLine}
	if len(*cachedGTFSData) != 0 {
		stopTimeUpdates := make([]types.CombinedStopTimeUpdate, 0)
		log.Default().Println("Cache Hit")
		for _, tripUpdate := range *cachedGTFSData {
			trip := tripUpdate.GetTrip()
			for _, stopTime := range tripUpdate.GetStopTimeUpdate() {
				stopTimeUpdate, err := utils.ParseTripUpdate(trip, stopTime, client.Config.StopID, client.Config.Minimal, client.Config.Headsigns)
				if err == nil {
					stopTimeUpdates = append(stopTimeUpdates, stopTimeUpdate)
				}
			}
		}

		if len(stopTimeUpdates) > 0 {
			trainsByDirection := client.Config.SortFunc(utils.ConvertToTrainSliceAndParse(stopTimeUpdates))
			nextTrain.TrainsByDirection = utils.ReturnLimit(trainsByDirection, client.Config.Limit)

		}

		for i := 0; i < 2; i++ {
			client.writeJSON(Message{Client: client, Message: nextTrain})
		}
	}

	for {
		data, ok := <-client.Send
		if !ok {
			log.Default().Println("Error writing to ClientId: ", client.UUID)
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		nextTrain.TrainsByDirection.North = make([]types.Train, 0)
		nextTrain.TrainsByDirection.South = make([]types.Train, 0)
		stopTimeUpdates := make([]types.CombinedStopTimeUpdate, 0)

		for _, tripUpdate := range data {
			trip := tripUpdate.GetTrip()
			for _, stopTime := range tripUpdate.GetStopTimeUpdate() {
				stopTimeUpdate, err := utils.ParseTripUpdate(trip, stopTime, client.Config.StopID, client.Config.Minimal, client.Config.Headsigns)
				if err == nil {
					stopTimeUpdates = append(stopTimeUpdates, stopTimeUpdate)
				}
			}
		}

		if len(stopTimeUpdates) > 0 {
			trainsByDirection := client.Config.SortFunc(utils.ConvertToTrainSliceAndParse(stopTimeUpdates))
			nextTrain.TrainsByDirection = utils.ReturnLimit(trainsByDirection, client.Config.Limit)
		}

		println("IS MIN: ", client.Config.Minimal)
		println("IS LIMIT: ", client.Config.Limit)
		client.writeJSON(Message{Client: client, Message: nextTrain})
	}
}

func (client *Client) writeJSON(msg Message) {
	w, errWriter := client.Conn.NextWriter(websocket.TextMessage)
	if errWriter != nil {
		log.Println(errWriter)
		return
	}

	json, jsonErr := json.Marshal(msg.Message)
	if jsonErr != nil {
		log.Println(jsonErr)
		return
	}
	l, errNW := w.Write(json)
	if errNW != nil {
		log.Println(errNW)
		return
	}
	log.Printf("JSON: %v\n bytes written: %v\n", msg.Message, l)
}

func (client *Client) SortConfig() {
	switch client.Config.Sort {
	case "descending":
		client.Config.SortFunc = utils.DescendingSort
	default:
		client.Config.SortFunc = utils.DefaultSort
	}
}

// GeneratorConfig is also probably overkill
func (client *Client) GeneratorConfig() {
	switch client.Config.Generate {
	case "test":
		client.Config.Generator = utils.TestGen
	default:
		client.Config.Generator = nil
	}
}
