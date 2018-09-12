package seth

type ParityClient struct {
	Client
}

func (c *ParityClient) Chain() (out string, err error) {
	err = c.Do("parity_chain", nil, &out)
	return
}

func (c *ParityClient) Mode() (out string, err error) {
	err = c.Do("parity_mode", nil, &out)
	return
}

type ParityNetPeers struct {
	Active    int `json:"active"`
	Connected int `json:"connected"`
	Max       int `json:"max"`
	Peers     []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Network struct {
			LocalAddress  string `json:"localAddress"`
			RemoteAddress string `json:"remoteAddress"`
		} `json:"network"`
		Protocols map[string]struct {
			Difficulty string `json:"difficulty"`
			Head       string `json:"head"`
			Version    int    `json:"version"`
		} `json:"protocols"`
		Caps []string `json:"caps"`
	} `json:"peers"`
}

func (c *ParityClient) NetPeers() (out ParityNetPeers, err error) {
	err = c.Do("parity_netPeers", nil, &out)
	return
}
