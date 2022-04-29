package server

type config struct {
	Server struct {
		Node         int64  `json:"node"`
		Version      string `json:"version"`
		RunningModel string `json:"runningModel"`
	}
}
