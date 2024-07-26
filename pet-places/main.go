package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)
type Place struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Address  string `json:"address"`
	Description string `json:"description"`
}

// var places = []Place{
// 	{ID: 1, Name: "Paddock For Paws", Address:"HP3 0JS", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
// 	{ID: 2, Name: "Dinosaur Safari Adventure Golf", Address:"EN5 3HW", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
// 	{ID: 3, Name: "Three Horseshoes", Address:"AL4 0HP", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
// }

var nextID = 4

var places = map[int]Place{
	1:{ID: 1, Name: "Paddock For Paws", Address:"HP3 0JS", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
	2:{ID: 2, Name: "Dinosaur Safari Adventure Golf", Address:"EN5 3HW", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
	3:{ID: 3, Name: "Three Horseshoes", Address:"AL4 0HP", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
}

func main() {
	// Postgres connection
	connStr := "postgres://user:password@localhost:5432/petplaces?sslmode=disable"
	db,err:=sql.Open("postgres",connStr)
	if err!=nil {
		log.Fatal(err)
	}
	defer db.Close()

	pingErr:=db.Ping()
	if pingErr!=nil {
		log.Fatal(pingErr)
	}

	createPlaceTable(db)

	mux := http.NewServeMux()
  mux.HandleFunc("GET /places", getPlaces)

	mux.HandleFunc("POST /places", createPlace)

  mux.HandleFunc("GET /places/{id}", getPlace)	

	mux.HandleFunc("PUT /places/{id}", updatePlace)

	mux.HandleFunc("DELETE /places/{id}", deletePlace)
  
	slog.Info("Starting on port 8080")
  http.ListenAndServe("localhost:8080", mux)
	

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

func getPlace(w http.ResponseWriter, r *http.Request) {
	
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Fatal(err)
	}
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

	// Insert place into db
	// place.ID = nextID
	// nextID += 1
	place.ID = insertPlace(db,place)
	places[place.ID] = place
	// places = append(places, place)
	// places[nextID] = place
	resp, err := json.Marshal(place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func updatePlace(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Do we need loop?
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
			err=updatePlaceInDB(db, updatedPlace)
			if err!=nil {
				log.Fatal(err)
			}
			places[index] = updatedPlace

			resp, err := json.Marshal(updatedPlace)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(resp)
		}
	}
	io.WriteString(w, "Place not found")
}

func deletePlace(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Fatal(err)
	}	

	err=deletePlaceFromDB(db,id)
	if err!=nil {
		log.Fatal(err)
	}
			// places = append(places[:index], places[index+1:]...)
	delete(places, id)	
}


func createPlaceTable(db *sql.DB) {
	query:= `CREATE TABLE IF NOT EXISTS place (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	address VARCHAR(100) NOT NULL,
	description TEXT,
	created timestamp DEFAULT NOW()
	)`

	_,err:=db.Exec(query)
	if err!=nil {
		log.Fatal(err)
	}

}

func insertPlace(db *sql.DB, place Place) int {
	query:= `INSERT INTO place (name, address, description)
	VALUES ($1, $2, $3) RETURNING id`
	var id int
	err:=db.QueryRow(query, place.Name, place.Address, place.Description).Scan(&id)
	if err!=nil {
		log.Fatal(err)
	}
	fmt.Print("Inserted ID=",id)
	return id
}

func updatePlaceInDB(db *sql.DB, place Place) error {
	query := `UPDATE place SET name=$1, address=$2, description=$3 WHERE id=$4`
	_, err := db.Exec(query, place.Name, place.Address, place.Description, place.ID)
	fmt.Print("Updated",place)
	return err
}

func deletePlaceFromDB(db *sql.DB, id int) error {
	query :=`DELETE FROM place WHERE id=$1`
	_,err:=db.Exec(query,id)
	fmt.Print("Deleted ID=",id)
	
	return err
}
// Data structures: Map (id), Struc
// Unit testing