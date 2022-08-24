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

//Client struct holds all client related data
type Client struct {
	UUID       uuid.UUID
	Pool       *Pool
	Conn       *websocket.Conn
	Send       chan []*gtfs.TripUpdate_StopTimeUpdate
	StopID     string
	SubwayLine string
	Config     types.Config
	Fetching   bool
}

//Message Struct
type Message struct {
	Message types.UpcomingTrain //Confusing naming convention
	Client  *Client
}

func (client *Client) read() {
	defer func() {
		log.Printf("Closing Read Client: %v \n", client.UUID)
		client.Pool.Unregister <- client
		client.Conn.Close()
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

func (client *Client) write(cachedGTFSData *[]*gtfs.TripUpdate_StopTimeUpdate) {
	defer log.Printf("Closing Write Client: %v \n", client.UUID)

	stopTimeUpdateSlice := make([]*types.StopTimeUpdate, 0)
	upcomingTrain := &types.UpcomingTrain{ClientID: client.UUID, SubwayLine: client.Config.SubwayLine}

	if len(*cachedGTFSData) != 0 {
		log.Printf("WRITE FROM CACHE: %v\n", client.UUID)
		for _, tripUpdate := range *cachedGTFSData {
			match, stopTimeUpdate := utils.FindStopData(tripUpdate, client.Config.StopID)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			unparsed, parsed := utils.ConvertToTrainSliceAndParse(stopTimeUpdateSlice)
			upcomingTrain.Trains = unparsed
			upcomingTrain.ParsedTrains = client.Config.Funct(parsed)
		}

		i := 0
		for i < 2 {
			client.writeJSON(Message{Client: client, Message: *upcomingTrain})
			i++
		}
	}
	log.Printf("START WRITING FOR CLIENT: %v \n", client.UUID)
	for {
		data, ok := <-client.Send
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		//not sure if rezeroing the slices is actually necessary
		upcomingTrain.Trains = make([]*types.Train, 0)
		upcomingTrain.ParsedTrains.Northbound = make([]*types.Train, 0)
		upcomingTrain.ParsedTrains.SouthBound = make([]*types.Train, 0)
		stopTimeUpdateSlice = make([]*types.StopTimeUpdate, 0)

		for _, tripUpdate := range data {
			match, stopTimeUpdate := utils.FindStopData(tripUpdate, client.Config.StopID)
			if match {
				stopTimeUpdateSlice = append(stopTimeUpdateSlice, stopTimeUpdate)
			}
		}

		if len(stopTimeUpdateSlice) > 0 {
			unparsed, parsed := utils.ConvertToTrainSliceAndParse(stopTimeUpdateSlice)
			upcomingTrain.Trains = unparsed
			upcomingTrain.ParsedTrains = client.Config.Funct(parsed)
		}

		client.writeJSON(Message{Client: client, Message: *upcomingTrain})
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

//ConfigureSort is probably really over kill but this lets you write special
//parsers/sorters serverside and apply them to multiple clients at once.
//Eventaully use an enum/constant for the sort configurations/functions.
//Can probably use this to return custom data types i.e. mixing train times
// and service alert information.
func (client *Client) ConfigureSort() {
	switch client.Config.Sort {
	case "descending":
		client.Config.Funct = utils.DescendingSort
	default:
		client.Config.Funct = utils.DefaultSort
	}
}

//ConfigureGenerator is also probably overkill
func (client *Client) ConfigureGenerator() {
	switch client.Config.Generate {
	case "test":
		client.Config.Generator = utils.TestGen
	default:
		client.Config.Generator = nil
	}
}
