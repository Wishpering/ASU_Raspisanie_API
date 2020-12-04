package main

type Pool struct {
	Count   int                 `json:"count"`
	Payload map[string][]string `json:"payload"`
}

type Configuration struct {
	Port       string
	Compress   bool
	DB_address string
	DB_port    string
}
