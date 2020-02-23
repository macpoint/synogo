package synoclient

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Host     string        `json:"host"`
	Scheme   string        `json:"scheme"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Timeout  time.Duration `json:"timeout"`
}

func LoadJsonConfiguration(file string) (config *Config, err error) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, &GenericError{desc: err.Error()}
	}
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		return nil, &GenericError{desc: fmt.Sprintf("Could not parse JSON config: %v ", err)}
	}
	return config, nil
}
