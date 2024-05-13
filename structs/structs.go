package structs

type DsConfig struct {
	Self       DsConfigSelf        `json:"self"`
	Network    DsConfigNetwork     `json:"network"`
	Neighbours []DsConfigNeighbour `json:"neighbours"`
}

type DsConfigSelf struct {
	Name     string                  `json:"name"`
	Protocol string                  `json:"protocol"`
	URL      string                  `json:"url"`
	Port     string                  `json:"port"`
	ApiPort  string                  `json:"apiPort"`
	Identity DsConfigPrivateIdentity `json:"identity"`
}

type DsConfigNetwork struct {
	ConsensusThreshold int    `json:"consensusThreshold"`
	Consensus          string `json:"consensus"`
}

type DsConfigNeighbour struct {
	Name     string           `json:"name"`
	Protocol string           `json:"protocol"`
	URL      string           `json:"url"`
	Port     string           `json:"port"`
	Identity DsConfigIdentity `json:"identity"`
}

type DsConfigIdentity struct {
	Type      string `json:"type"`
	PublicKey string `json:"publicKey"`
}

type DsConfigPrivateIdentity struct {
	Type       string `json:"type"`
	PrivateKey string `json:"privateKey"`
}

type SelfResp struct {
	Name       string        `json:"name"`
	PublicKey  string        `json:"publicKey"`
	Neighbours SelfNeighbour `json:"neighbours"`
}

type SelfNeighbour struct {
	Count int           `json:"count"`
	Nodes []interface{} `json:"nodes"`
}
