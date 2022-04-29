package server

type config struct {
	Server struct {
		Node         int64  `json:"node"`
		Version      string `json:"version"`
		RunningModel string `json:"runningModel"`
	}
	Debug struct {
		IsCheckHeartbeat     bool `json:"isCheckHeartbeat"`
		IsRecordHeartbeatLog bool `json:"isRecordHeartbeatLog"`
	} `json:"debug"`
	Tcp struct {
		Ip   string `json:"ip" `
		Port string `json:"port"`
	}
	Websocket struct {
		Ip    string `json:"ip" `
		Port  string `json:"port"`
		IsWss bool   `json:"isWss"`
	}
	Grpc struct {
		Ip   string `json:"ip" `
		Port string `json:"port"`
	} `json:"grpc"`
}
