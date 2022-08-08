package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

/********************
MOVE ALL OF THIS INTO A CLOUDFLARE WORKER?


subway line needs to map to SUBWAY_LINE_REQUEST_URLS constant since this
is how the pools are segmented.
var SUBWAY_LINE_REQUEST_URLS = map[string]string {
 "L": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l",
 "ACE": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace",
 "BDFM": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm",
 "G": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g",
 "JZ": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz",
 "NQRW": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw",
 "NUMBERS": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs",
 "SERVICE": "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/camsys%2Fsubway-alerts.json",
}
********************/

type Station struct {
	StationId      string `json:"stationId"`
	ComplexId      string `json:"complexId"`
	StopId         string `json:"stopId"`
	SubwayLine     string `json:"subwayLine"`
	StopName       string `json:"stopName"`
	Borough        string `json:"borough"`
	Routes         string `json:"routes"`
	Lattitude      string `json:"lattitude"`
	Longitude      string `json:"longitude"`
	NorthDirection string `json:"northDirectionLabel"`
	SouthDirection string `json:"southDirectionLabel"`
}

type SubwayLineMap struct {
	NUMBERS ParsedStationMap
	ACE     ParsedStationMap
	BDFM    ParsedStationMap
	NQRW    ParsedStationMap
	L       ParsedStationMap
	G       ParsedStationMap
	S       ParsedStationMap
	JZ      ParsedStationMap
	SERVICE ParsedStationMap
}

func parseStaticTripsCSV(data [][]string) map[string]TripHeadSign {
	tripSubwayMap := make(map[string]TripHeadSign, 0)
	tripHeadSign := TripHeadSign{}
	for i, line := range data {
		var direction string
		var headSign string
		//var shapeId string
		var routeId string
		if i > 0 {
			for j, field := range line {
				switch {
				case j == 0:
					//Mapping GS to S for now. Need to update to account
					//For both Shuttles
					if field == "GS" {
						routeId = "S"
					} else {
						routeId = field
					}
				case j == 3:
					headSign = field
				case j == 4:
					direction = field
				}
			}
		}
		if routeId != "" || headSign != "" {

			switch {
			case direction == "0":
				tripHeadSign.North = headSign
			case direction == "1":
				tripHeadSign.South = headSign
			}
			tripSubwayMap[routeId] = tripHeadSign
		}
	}
	return tripSubwayMap
}

func parseStaticStationCSV(data [][]string) []Station {
	stationList := make([]Station, 0)
	for i, line := range data {
		if i > 0 {
			station := Station{}
			for j, field := range line {
				switch {
				case j == 0:
					station.StationId = field
				case j == 1:
					station.ComplexId = field
				case j == 2:
					station.StopId = field
				case j == 4:
					station.SubwayLine = field
				case j == 5:
					station.StopName = field
				case j == 6:
					station.Borough = field
				case j == 7:
					station.Routes = field
				case j == 9:
					station.Lattitude = field
				case j == 10:
					station.Longitude = field
				case j == 11:
					station.NorthDirection = field
				case j == 12:
					station.SouthDirection = field
				}
			}
			stationList = append(stationList, station)
		}
	}
	return stationList
}

func createStationToSubwayLineMap(stations []Station) SubwayLineMap {
	stationMap := SubwayLineMap{
		NUMBERS: ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		ACE:     ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		BDFM:    ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		NQRW:    ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		L:       ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		G:       ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		SERVICE: ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		JZ:      ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		S:       ParsedStationMap{Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
	}

	leftover := make([]Station, 0)
	for _, station := range stations {
		routes := station.Routes
		trim := strings.ToUpper(strings.ReplaceAll(routes, " ", ""))

		//not sure what subway line this is right now
		if trim == "SIR" {
			leftover = append(leftover, station)
			continue
		}

		if strings.Contains("L", trim) {
			stationMap.L.Stations = append(stationMap.L.Stations, station)
			stationMap.L.StationsByBorough[station.Borough] = append(stationMap.L.StationsByBorough[station.Borough], station)
			//stationMap.L = append(stationMap.L, station)
			//stationMap.L[station.borough] = append()
		} else if strings.Contains("G", trim) {
			//stationMap.G = append(stationMap.G, station)
			stationMap.G.Stations = append(stationMap.G.Stations, station)
			stationMap.G.StationsByBorough[station.Borough] = append(stationMap.G.StationsByBorough[station.Borough], station)
		} else if strings.Contains("S", trim) {
			//stationMap.S = append(stationMap.S, station)
			stationMap.S.Stations = append(stationMap.S.Stations, station)
			stationMap.S.StationsByBorough[station.Borough] = append(stationMap.S.StationsByBorough[station.Borough], station)
		} else if containsAny("ACE", trim) {
			//stationMap.ACE = append(stationMap.ACE, station)
			stationMap.ACE.Stations = append(stationMap.ACE.Stations, station)
			stationMap.ACE.StationsByBorough[station.Borough] = append(stationMap.ACE.StationsByBorough[station.Borough], station)
		} else if containsAny("BDFM", trim) {
			//stationMap.BDFM = append(stationMap.BDFM, station)
			stationMap.BDFM.Stations = append(stationMap.BDFM.Stations, station)
			stationMap.BDFM.StationsByBorough[station.Borough] = append(stationMap.BDFM.StationsByBorough[station.Borough], station)
		} else if containsAny("JZ", trim) {
			//stationMap.JZ = append(stationMap.JZ, station)
			stationMap.JZ.Stations = append(stationMap.JZ.Stations, station)
			stationMap.JZ.StationsByBorough[station.Borough] = append(stationMap.JZ.StationsByBorough[station.Borough], station)
		} else if containsAny("NQRW", trim) {
			//stationMap.NQRW = append(stationMap.NQRW, station)
			stationMap.NQRW.Stations = append(stationMap.NQRW.Stations, station)
			stationMap.NQRW.StationsByBorough[station.Borough] = append(stationMap.NQRW.StationsByBorough[station.Borough], station)
		} else if containsAny("1234567", trim) {
			//stationMap.NUMBERS = append(stationMap.NUMBERS, station)
			stationMap.NUMBERS.Stations = append(stationMap.NUMBERS.Stations, station)
			stationMap.NUMBERS.StationsByBorough[station.Borough] = append(stationMap.NUMBERS.StationsByBorough[station.Borough], station)
		}
	}

	return stationMap
}

func containsAny(str string, substr string) bool {
	for _, l := range str {
		if strings.Contains(substr, string(l)) {
			return true
		}
	}
	return false
}

func readCSV(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err) //change error handling, probably don't want to crash the server for this?
	}
	return data
}

func Process() StaticData {
	stationData := readCSV("./google_transit/stations.csv")
	stations := parseStaticStationCSV(stationData)
	stationSubwaLineMap := createStationToSubwayLineMap(stations)

	trips := readCSV("./google_transit/trips.csv")
	subwayTripMap := parseStaticTripsCSV(trips)

	return StaticData{
		StationMap:    stationSubwaLineMap,
		SubwayTripMap: subwayTripMap,
	}
}
