package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type place struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
}

var places = []place{
	{ID: "1", Name: "Paddock For Paws", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
	{ID: "2", Name: "Dinosaur Safari Adventure Golf", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
	{ID: "3", Name: "Three Horseshoes", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
}

func main() {
	http.HandleFunc("/places", getPlaces)	

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getPlaces(w http.ResponseWriter, _ *http.Request) {
	// io.WriteString(w, "Hello from a HandleFunc #1!\n")	
	// How to handle errors
	resp, err :=json.Marshal(places)
	if err != nil {
    // log.Fatal(err)
		http.Error(w,err.Error(),500)
		return
} 
w.Write(resp)
	
}