package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Me                Me      `json:"me"`
	Network           Network `json:"network"`
	NearestNeighbours []Node  `json:"nearestNeighbours"`
}

type Me struct {
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	URL           string `json:"url"`
	Port          string `json:"port"`
	ApiPort       string `json:"apiPort"`
	DatastorePath string `json:"datastorePath"`
}

type Network struct {
	ConsensusThreshold int    `json:"consensusThreshold"`
	Consensus          string `json:"consensus"`
}

type Node struct {
	Name     string   `json:"name"`
	Protocol string   `json:"protocol"`
	URL      string   `json:"url"`
	Port     string   `json:"port"`
	Identity Identity `json:"identity"`
}

type Identity struct {
	Type      string `json:"type"`
	PublicKey string `json:"publicKey"`
}

func Get() (*Config, error) {
	config := Config{}

	bConf, err := os.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bConf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
