package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Place struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Address  string `json:"address"`
	Description string `json:"description"`
}

var places = []Place{
	{ID: "1", Name: "Paddock For Paws", Address:"HP3 0JS", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
	{ID: "2", Name: "Dinosaur Safari Adventure Golf", Address:"EN5 3HW", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
	{ID: "3", Name: "Three Horseshoes", Address:"AL4 0HP", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
}

func main() {
	http.HandleFunc("/places", getPlaces)	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getPlaces(w http.ResponseWriter, _ *http.Request) {
	// How to handle errors
	resp, err :=json.Marshal(places)
	if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
		return
} 
w.Write(resp)	
}

func getPlace(w http.ResponseWriter, _ *http.Request, id string) {
	for _, place := range places {
		if place.ID == id {
			resp, err := json.Marshal(place)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(resp)
			return
		}
}
// Where to put this?
	io.WriteString(w, "Place not found")
}

func createPlace(w http.ResponseWriter, r *http.Request) {
	var place Place
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Use pointer &place to directly modify the original place variable 
	err = json.Unmarshal(body, &place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	place.ID = strconv.Itoa(len(places) + 1)
	places = append(places, place)
	resp, err := json.Marshal(place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func updatePlace(w http.ResponseWriter, r *http.Request, id string) {
	for index, item := range places {
		if item.ID == id {
			var updatedPlace Place
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(body, &updatedPlace)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updatedPlace.ID = id // Ensure the ID remains the same
			places[index] = updatedPlace
			resp, err := json.Marshal(updatedPlace)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(resp)
			return
		}
	}
	io.WriteString(w, "Place not found")
}

func deletePlace(w http.ResponseWriter, _ *http.Request, id string) {
	for index, item := range places {
		if item.ID == id {
			places = append(places[:index], places[index+1:]...)
			resp, err := json.Marshal(places)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(resp)			
		}
	}
	io.WriteString(w, "Place not found")
}

func handler(w http.ResponseWriter, r *http.Request) {
	
}