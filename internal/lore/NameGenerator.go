package lore

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/mechanical-lich/mlge/utility"
)

const Fallback_Name = "Fallback Name"

type NameData struct {
	FirstNames     []string `json:"first_names"`
	LastNames      []string `json:"last_names"`
	SettlementName []string `json:"settlement_names"`
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

func RandomSettlementName(race string) string {
	if _, ok := nameData[race]; ok {
		names := nameData[race]
		return names.SettlementName[utility.GetRandom(0, len(names.SettlementName))]
	}

	log.Print("Using fallback settlement name")
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
