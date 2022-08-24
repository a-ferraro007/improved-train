package utils

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

/*
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
*/

//Station struct
type Station struct {
	StationID      string `json:"stationId"`
	ComplexID      string `json:"complexId"`
	StopID         string `json:"stopID"`
	SubwayLine     string `json:"subwayLine"`
	StopName       string `json:"stopName"`
	Borough        string `json:"borough"`
	Routes         string `json:"routes"`
	Lattitude      string `json:"lattitude"`
	Longitude      string `json:"longitude"`
	NorthDirection string `json:"northDirectionLabel"`
	SouthDirection string `json:"southDirectionLabel"`
}

//SubwayStationMap is a slice of stations mapped to a
//subway line
type SubwayStationMap struct {
	L       []Station
	ACE     []Station
	BDFM    []Station
	G       []Station
	JZ      []Station
	NQRW    []Station
	NUMBERS []Station
	SHUTTLE []Station
	SERVICE []Station
}

func createSliceOfStations(data [][]string) []Station {
	stationList := make([]Station, 0)
	for i, line := range data {
		if i > 0 {
			station := Station{}
			for j, field := range line {
				switch {
				case j == 0:
					station.StationID = field
				case j == 1:
					station.ComplexID = field
				case j == 2:
					station.StopID = field
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

func createStationToSubwayLineMap(stations []Station) SubwayStationMap {
	stationMap := SubwayStationMap{
		L:       make([]Station, 0),
		ACE:     make([]Station, 0),
		BDFM:    make([]Station, 0),
		G:       make([]Station, 0),
		JZ:      make([]Station, 0),
		NQRW:    make([]Station, 0),
		NUMBERS: make([]Station, 0),
		SHUTTLE: make([]Station, 0),
		SERVICE: make([]Station, 0),
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
			stationMap.L = append(stationMap.L, station)
		} else if strings.Contains("G", trim) {
			stationMap.G = append(stationMap.G, station)
		} else if strings.Contains("S", trim) {
			stationMap.SHUTTLE = append(stationMap.SHUTTLE, station)
		} else if containsAny("ACE", trim) {
			stationMap.ACE = append(stationMap.ACE, station)
		} else if containsAny("BDFM", trim) {
			stationMap.BDFM = append(stationMap.BDFM, station)
		} else if containsAny("JZ", trim) {
			stationMap.JZ = append(stationMap.JZ, station)
		} else if containsAny("NQRW", trim) {
			stationMap.NQRW = append(stationMap.NQRW, station)
		} else if containsAny("1234567", trim) {
			stationMap.NUMBERS = append(stationMap.NUMBERS, station)
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

//Process function
func Process() SubwayStationMap {
	f, err := os.Open("./static_transit/stations.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	stations := createSliceOfStations(data)
	return createStationToSubwayLineMap(stations)
}
