package fileprocessing

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type Station struct {
	StationId      string `json:"stationId"`
	ComplexId      string `json:"complexId"`
	GtfsStopId     string `json:"gtfsStopId"`
	SubwayLine     string `json:"subwayLine"`
	StopName       string `json:"stopName"`
	Borough        string `json:"borough"`
	Lattitude      string `json:"lattitude"`
	Longitude      string `json:"longitude"`
	NorthDirection string `json:"northDirectionLabel"`
	SouthDirection string `json:"southDirectionLabel"`
}

func createStationObject(data [][]string) []Station {
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
					station.GtfsStopId = field
				case j == 4:
					station.SubwayLine = field
				case j == 5:
					station.StopName = field
				case j == 6:
					station.Borough = field
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

func Process() []Station {
	f, err := os.Open("./fileprocessing/stations.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	stations := createStationObject(data)

	// print the array
	fmt.Printf("%+v\n", stations)
}
