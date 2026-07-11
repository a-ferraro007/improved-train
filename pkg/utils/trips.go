package utils

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// CreateHeadSignsMap Function
func CreateHeadSignsMap() map[string]string {
	m := make(map[string]string)
	lines, requestA := 0, 0
	f, err := os.Open("../../static_transit/trips.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
		// filter request a
		line := scanner.Bytes()
		// if len(line) <= 30 || line[30] != 'A' {
		// 	continue
		// }
		// if !bytes.Equal(line[22:], []byte("REQUEST-A")) {
		// 	continue
		// }
		requestA++
		// request := string(line)
		chunks := strings.Split(string(line), ",")
		split := strings.Split(string(chunks[len(chunks)-1]), "..")
		if len(split) == 2 {
			trip := string(split[0]) + ".." + string(split[1][0])
			_, ok := m[trip]
			if !ok {
				m[trip] = chunks[len(chunks)-3]
			}
		} else {
			split := strings.Split(string(chunks[len(chunks)-1]), ".")
			if len(split) == 2 {
				trip := string(split[0]) + ".." + string(split[1][0])
				_, ok := m[trip]
				if !ok {
					m[trip] = chunks[len(chunks)-3]
				}

			} else {
				log.Default().Println(chunks)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Default().Println(err)
	}
	log.Default().Println(m["L..N"])
	return m
}
