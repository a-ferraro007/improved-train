package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

//MOVE ALL OF THIS INTO A CLOUDFLARE WORKER?

func parseStaticTripsCSV(data [][]string) map[string]TripHeadSign {
	tripSubwayMap := make(map[string]TripHeadSign, 0)
	tripHeadSign := TripHeadSign{}
	for i, line := range data {
		var direction string
		var headSign string
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
	subwayLineMap := SubwayLineMap{
		One:   &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Two:   &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Three: &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Four:  &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Five:  &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Six:   &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Seven: &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		A:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		C:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		E:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		B:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		D:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		F:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		M:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		J:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Z:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		L:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		G:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		N:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		Q:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		R:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		W:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
		S:     &ParsedStationMap{seen: make(map[string]bool, 0), Stations: make([]Station, 0), StationsByBorough: map[string][]Station{}},
	}

	leftover := make([]Station, 0)
	for _, station := range stations {
		routes := station.Routes
		split := strings.Split(strings.ToUpper(routes), " ")
		for _, trim := range split {
			//not sure what subway line this is right now
			if trim == "SIR" {
				leftover = append(leftover, station)
				continue
			}

			switch {
			case trim == One:
				subwayLineMap.One.Stations = append(subwayLineMap.One.Stations, station)
				subwayLineMap.One.StationsByBorough[station.Borough] = append(subwayLineMap.One.StationsByBorough[station.Borough], station)
				subwayLineMap.One.seen[trim] = true
			case trim == Two:
				subwayLineMap.Two.Stations = append(subwayLineMap.Two.Stations, station)
				subwayLineMap.Two.StationsByBorough[station.Borough] = append(subwayLineMap.Two.StationsByBorough[station.Borough], station)
				subwayLineMap.Two.seen[trim] = true
			case trim == Three:
				subwayLineMap.Three.Stations = append(subwayLineMap.Three.Stations, station)
				subwayLineMap.Three.StationsByBorough[station.Borough] = append(subwayLineMap.Three.StationsByBorough[station.Borough], station)
				subwayLineMap.Three.seen[trim] = true
			case trim == Four:
				subwayLineMap.Four.Stations = append(subwayLineMap.Four.Stations, station)
				subwayLineMap.Four.StationsByBorough[station.Borough] = append(subwayLineMap.Four.StationsByBorough[station.Borough], station)
				subwayLineMap.Four.seen[trim] = true
			case trim == Five:
				subwayLineMap.Five.Stations = append(subwayLineMap.Five.Stations, station)
				subwayLineMap.Five.StationsByBorough[station.Borough] = append(subwayLineMap.Five.StationsByBorough[station.Borough], station)
				subwayLineMap.Five.seen[trim] = true
			case trim == Six:
				subwayLineMap.Six.Stations = append(subwayLineMap.Six.Stations, station)
				subwayLineMap.Six.StationsByBorough[station.Borough] = append(subwayLineMap.Six.StationsByBorough[station.Borough], station)
				subwayLineMap.Six.seen[trim] = true
			case trim == Seven:
				subwayLineMap.Seven.Stations = append(subwayLineMap.Seven.Stations, station)
				subwayLineMap.Seven.StationsByBorough[station.Borough] = append(subwayLineMap.Seven.StationsByBorough[station.Borough], station)
				subwayLineMap.Seven.seen[trim] = true
			case trim == A:
				subwayLineMap.A.Stations = append(subwayLineMap.A.Stations, station)
				subwayLineMap.A.StationsByBorough[station.Borough] = append(subwayLineMap.A.StationsByBorough[station.Borough], station)
				subwayLineMap.A.seen[trim] = true
			case trim == C:
				subwayLineMap.C.Stations = append(subwayLineMap.C.Stations, station)
				subwayLineMap.C.StationsByBorough[station.Borough] = append(subwayLineMap.C.StationsByBorough[station.Borough], station)
				subwayLineMap.C.seen[trim] = true
			case trim == E:
				subwayLineMap.E.Stations = append(subwayLineMap.E.Stations, station)
				subwayLineMap.E.StationsByBorough[station.Borough] = append(subwayLineMap.E.StationsByBorough[station.Borough], station)
				subwayLineMap.E.seen[trim] = true
			case trim == B:
				subwayLineMap.B.Stations = append(subwayLineMap.B.Stations, station)
				subwayLineMap.B.StationsByBorough[station.Borough] = append(subwayLineMap.B.StationsByBorough[station.Borough], station)
				subwayLineMap.B.seen[trim] = true
			case trim == D:
				subwayLineMap.D.Stations = append(subwayLineMap.D.Stations, station)
				subwayLineMap.D.StationsByBorough[station.Borough] = append(subwayLineMap.D.StationsByBorough[station.Borough], station)
				subwayLineMap.D.seen[trim] = true
			case trim == F:
				subwayLineMap.F.Stations = append(subwayLineMap.F.Stations, station)
				subwayLineMap.F.StationsByBorough[station.Borough] = append(subwayLineMap.F.StationsByBorough[station.Borough], station)
				subwayLineMap.F.seen[trim] = true
			case trim == M:
				subwayLineMap.M.Stations = append(subwayLineMap.M.Stations, station)
				subwayLineMap.M.StationsByBorough[station.Borough] = append(subwayLineMap.M.StationsByBorough[station.Borough], station)
				subwayLineMap.M.seen[trim] = true
			case trim == N:
				subwayLineMap.N.Stations = append(subwayLineMap.N.Stations, station)
				subwayLineMap.N.StationsByBorough[station.Borough] = append(subwayLineMap.N.StationsByBorough[station.Borough], station)
				subwayLineMap.N.seen[trim] = true
			case trim == Q:
				subwayLineMap.Q.Stations = append(subwayLineMap.Q.Stations, station)
				subwayLineMap.Q.StationsByBorough[station.Borough] = append(subwayLineMap.Q.StationsByBorough[station.Borough], station)
				subwayLineMap.Q.seen[trim] = true
			case trim == R:
				subwayLineMap.R.Stations = append(subwayLineMap.R.Stations, station)
				subwayLineMap.R.StationsByBorough[station.Borough] = append(subwayLineMap.R.StationsByBorough[station.Borough], station)
				subwayLineMap.R.seen[trim] = true
			case trim == W:
				subwayLineMap.W.Stations = append(subwayLineMap.W.Stations, station)
				subwayLineMap.W.StationsByBorough[station.Borough] = append(subwayLineMap.W.StationsByBorough[station.Borough], station)
				subwayLineMap.W.seen[trim] = true
			case trim == L:
				subwayLineMap.L.Stations = append(subwayLineMap.L.Stations, station)
				subwayLineMap.L.StationsByBorough[station.Borough] = append(subwayLineMap.L.StationsByBorough[station.Borough], station)
				subwayLineMap.L.seen[trim] = true
			case trim == G:
				subwayLineMap.G.Stations = append(subwayLineMap.G.Stations, station)
				subwayLineMap.G.StationsByBorough[station.Borough] = append(subwayLineMap.G.StationsByBorough[station.Borough], station)
				subwayLineMap.G.seen[trim] = true
			case trim == S:
				subwayLineMap.S.Stations = append(subwayLineMap.S.Stations, station)
				subwayLineMap.S.StationsByBorough[station.Borough] = append(subwayLineMap.S.StationsByBorough[station.Borough], station)
				subwayLineMap.S.seen[trim] = true
			case trim == J:
				subwayLineMap.J.Stations = append(subwayLineMap.J.Stations, station)
				subwayLineMap.J.StationsByBorough[station.Borough] = append(subwayLineMap.J.StationsByBorough[station.Borough], station)
				subwayLineMap.J.seen[trim] = true
			case trim == Z:
				subwayLineMap.Z.Stations = append(subwayLineMap.Z.Stations, station)
				subwayLineMap.Z.StationsByBorough[station.Borough] = append(subwayLineMap.Z.StationsByBorough[station.Borough], station)
				subwayLineMap.Z.seen[trim] = true
			}
		}
	}
	return subwayLineMap
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
	stationSubwayLineMap := createStationToSubwayLineMap(stations)
	//log.Println(stationSubwayLineMap.G)
	trips := readCSV("./google_transit/trips.csv")
	subwayTripMap := parseStaticTripsCSV(trips)

	return StaticData{
		StationMap:    stationSubwayLineMap,
		SubwayTripMap: subwayTripMap,
	}
}
