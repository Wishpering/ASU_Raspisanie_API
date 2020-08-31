package main

type Token struct {
	Key string `json:"key"`
}

type Pool struct {
	Count   int                 `json:"count"`
	Payload map[string][]string `json:"payload"`
}

type Configuration struct {
	Password   []byte
	Port       string
	Compress   bool
	DB_address string
	DB_port    string
}
