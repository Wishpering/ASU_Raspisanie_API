package main

type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Faculty struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Groups []Group `json:"groups"`
}

type Para struct {
	Time       string `json:"time"`
	Name       string `json:"name"`
	Prepod     string `json:"prepod"`
	Auditoriya string `json:"auditoriya"`
}

type MongoSchedule struct {
	Date     string `json:"date"`
	Schedule []Para `json:"schedule"`
}
