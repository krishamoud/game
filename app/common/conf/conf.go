// Package conf handles all of the applications configuration management
package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// AppConf will hold all of our configuration information
var AppConf = getConf()

// Configuration is the struct that handles the json file with all config info
type Configuration struct {
	Host                     string
	Port                     int
	FoodMass                 float64
	FireFood                 int
	LimitSplit               int
	DefaultPlayerMass        float64
	Virus                    `json:"virus"`
	GameWidth                float64
	GameHeight               float64
	AdminPass                string
	GameMass                 float64
	MaxFood                  float64
	MaxVirus                 int
	SlowBase                 float64
	LogChat                  int
	NetworkUpdateFactor      int
	MaxHeartBeatInterval     int
	FoodUniformDisposition   bool
	VirusUniformDisposition  bool
	NewPlayerInitialPosition string
	MassLossRate             int
	MinMassLoss              int
	MergeTimer               int
}

// Virus handles all configuration with regards to viruses
type Virus struct {
	Fill        string
	Stroke      string
	StrokeWidth int
	DefaultMass `json:"defaultMass"`
	SplitMass   int
}

// DefaultMass is a range for how large a virus can be by default
type DefaultMass struct {
	From float64
	To   float64
}

func getConf() *Configuration {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
	}
	configuration := &Configuration{}
	if err := json.Unmarshal(file, &configuration); err != nil {
		panic(err)
	}
	return configuration
}
