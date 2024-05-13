package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/thejff/go-chucklechain/api"
	"github.com/thejff/go-chucklechain/config"
	"github.com/thejff/go-chucklechain/datastore"
	"github.com/thejff/go-chucklechain/key"
	"github.com/thejff/go-chucklechain/structs"
)

func main() {
	fileCfg, err := config.Get()
	if err != nil {
		log.Fatalln(err)
	}

	ds := datastore.NewDatastore[any](fileCfg.Me.DatastorePath)

	cfg, err := initialise(ds, fileCfg)
	if err != nil {
		log.Fatalln(err)
	}

	pkey, err := key.LoadRSA[any](cfg.Self.Identity.PrivateKey)
	if err != nil {
		log.Fatalln(err)
	}

	a := api.New(cfg, pkey)
	if err := a.Start(); err != nil {
		log.Fatalln(err)
	}

}

func initialise(ds datastore.Datastore[any], cfg *config.Config) (structs.DsConfig, error) {
	dsConf, err := ds.Read("config")
	if err != nil {
		if strings.ToLower(err.Error()) == "key not found" {
			return structs.DsConfig{}, onboardConfig(ds, cfg)
		}

		return structs.DsConfig{}, err
	}

	bCfg, err := json.Marshal(dsConf.Data)
	if err != nil {
		return structs.DsConfig{}, err
	}

	storedCfg := structs.DsConfig{}
	if err := json.Unmarshal(bCfg, &storedCfg); err != nil {
		return structs.DsConfig{}, err
	}

	fmt.Println("Config already exists")

	return storedCfg, nil
}

func onboardConfig(ds datastore.Datastore[any], cfg *config.Config) error {
	// Create a new config object
	// Store it in the datastore

	key, err := key.NewRSA[any]()
	if err != nil {
		return err
	}

	pem := key.GetPrivatePEM()

	neighbours := []structs.DsConfigNeighbour{}
	for _, n := range cfg.NearestNeighbours {
		neighbours = append(neighbours, structs.DsConfigNeighbour{
			Name:     n.Name,
			Protocol: n.Protocol,
			URL:      n.URL,
			Port:     n.Port,
			Identity: structs.DsConfigIdentity{
				Type:      n.Identity.Type,
				PublicKey: n.Identity.PublicKey,
			},
		})
	}

	data := structs.DsConfig{
		Self: structs.DsConfigSelf{
			Name:     cfg.Me.Name,
			Protocol: cfg.Me.Protocol,
			URL:      cfg.Me.URL,
			Port:     cfg.Me.Port,
			ApiPort:  cfg.Me.ApiPort,
			Identity: structs.DsConfigPrivateIdentity{
				Type:       "rsa",
				PrivateKey: pem,
			},
		},
		Network: structs.DsConfigNetwork{
			ConsensusThreshold: cfg.Network.ConsensusThreshold,
			Consensus:          cfg.Network.Consensus,
		},
		Neighbours: neighbours,
	}

	confObj, err := datastore.NewObject[any](data)
	if err != nil {
		return err
	}

	confObj.UUID = "config"
	confObj.Environment = "self"
	confObj.Type = "config"

	if err := ds.Write("config", confObj); err != nil {
		return err
	}

	return nil
}
