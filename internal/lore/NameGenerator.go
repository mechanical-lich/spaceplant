package lore

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mechanical-lich/mlge/utility"
)

const Fallback_Name = "Fallback Name"

type NameData struct {
	FirstNames   []string `json:"first_names"`
	LastNames    []string `json:"last_names"`
	StationWords []string `json:"station_words"`
}

var nameData map[string]NameData

func init() {
	log.Print("Loading name data...")
	nameData = make(map[string]NameData)

	if nameJson, err := os.OpenFile("data/name_data.json", os.O_RDONLY, 0644); err == nil {
		defer nameJson.Close()
		byteValue, _ := io.ReadAll(nameJson)
		err = json.Unmarshal(byteValue, &nameData)
		if err != nil {
			log.Print("Failed to unmarshal name_data.json:", err)
		}
	} else {
		log.Print("Failed to open name_data.json file:", err)
	}

}

func RandomName(race string) string {
	return RandomFirstName(race) + " " + RandomLastName(race)
}

func RandomStationName() string {
	if data, ok := nameData["stations"]; ok && len(data.StationWords) > 0 {
		wordCount := utility.GetRandom(1, 5) // 1 to 4 words
		if wordCount > len(data.StationWords) {
			wordCount = len(data.StationWords)
		}
		used := make(map[int]bool)
		words := make([]string, 0, wordCount)
		for len(words) < wordCount {
			idx := utility.GetRandom(0, len(data.StationWords))
			if !used[idx] {
				used[idx] = true
				words = append(words, data.StationWords[idx])
			}
		}
		return strings.Join(words, " ")
	}
	log.Print("Using fallback station name")
	return Fallback_Name
}

func RandomFirstName(race string) string {
	if _, ok := nameData[race]; ok {
		names := nameData[race]
		return names.FirstNames[utility.GetRandom(0, len(names.FirstNames))]
	}
	log.Print("Using fallback first name")
	return Fallback_Name
}

func RandomLastName(race string) string {
	if _, ok := nameData[race]; ok {
		names := nameData[race]
		return names.LastNames[utility.GetRandom(0, len(names.LastNames))]
	}
	log.Print("Using fallback last name")
	return Fallback_Name
}
