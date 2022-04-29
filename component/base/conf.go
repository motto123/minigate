package base

type config struct {
	Log struct {
		Path string `json:"path"`
		Out  string `json:"out"`
	} `json:"log"`
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		Db       uint8  `json:"db"`
	} `json:"redis"`
	Mysql struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Ip       string `json:"ip"`
		Port     int    `json:"port"`
		DbName   string `json:"dbName"`
	} `json:"mysql"`
	Mq struct{} `json:"mq"`
}
