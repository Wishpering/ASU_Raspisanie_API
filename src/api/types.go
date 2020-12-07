package main

type GroupsPool struct {
	Count  int      `json:"count"`
	Groups []string `json:"list"`
}
	
type Faculty struct {
	Name   string   `json:"name"`
	Groups GroupsPool `json:"groups"`
}

type FacultiesPool struct {
	Count     int       `json:"count"`
	Faculties []Faculty `json:"faculties"`
}
