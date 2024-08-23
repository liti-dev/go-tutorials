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
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// server struct holds the db config and router
type server struct {
	db     *sql.DB
	router *http.ServeMux
}

// var places = []Place{
// 	{ID: 1, Name: "Paddock For Paws", Address:"HP3 0JS", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
// 	{ID: 2, Name: "Dinosaur Safari Adventure Golf", Address:"EN5 3HW", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
// 	{ID: 3, Name: "Three Horseshoes", Address:"AL4 0HP", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
// }

// var nextID = 4

// no need for this when using db
// var places = map[int]Place{
// 	1:{ID: 1, Name: "Paddock For Paws", Address:"HP3 0JS", Description:"A lovely paddock for the dog's. The field is a good size and fully secure agility equipment on the field selection of toys, tennis balls, paddling pools in the summer. "},
// 	2:{ID: 2, Name: "Dinosaur Safari Adventure Golf", Address:"EN5 3HW", Description:"Bring the whole family, as everyone is welcome, even your dog!"},
// 	3:{ID: 3, Name: "Three Horseshoes", Address:"AL4 0HP", Description:"Our country pub in St Albans features seasonal food, cask ale and is dog friendly"},
// }

func main() {
	// Postgres connection
	connStr := "postgres://postgres:mysecretpassword@localhost:5432/petplaces?sslmode=disable"

	srv := &server{}

	// short var declaration := does not work with struct srv.db
	var err error
	srv.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.db.Close()

	pingErr := srv.db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	createPlaceTable(srv.db)

	// Set up router
	// mux := http.NewServeMux()
	srv.router = http.NewServeMux()
	srv.routes()

	slog.Info("Starting on port 8080")
	http.ListenAndServe("localhost:8080", srv.router)
}

func (s *server) routes() {
	s.router.HandleFunc("GET /places", s.getPlaces)
	s.router.HandleFunc("POST /places", s.createPlace)
	s.router.HandleFunc("GET /places/{id}", s.getPlace)
	s.router.HandleFunc("PUT /places/{id}", s.updatePlace)
	s.router.HandleFunc("DELETE /places/{id}", s.deletePlace)
}

func (s *server) getPlaces(w http.ResponseWriter, _ *http.Request) {
	// How to handle errors
	// Using db instead of local memory
	rows, err := s.db.Query("SELECT id, name, address, description FROM places")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var places []Place
	for rows.Next() {
		var place Place
		if err := rows.Scan(&place.ID, &place.Name, &place.Address, &place.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		places = append(places, place)
	}
	resp, err := json.Marshal(places)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (s *server) getPlace(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
	}

	var place Place
	err = s.db.QueryRow("SELECT id, name, address, description FROM places WHERE id = $1", id).Scan(&place.ID, &place.Name, &place.Address, &place.Description)
	if err == sql.ErrNoRows {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
	// 	for _, place := range places {
	// 		if place.ID == id {
	// 			resp, err := json.Marshal(place)
	// 			if err != nil {
	// 				http.Error(w, err.Error(), http.StatusInternalServerError)
	// 				return
	// 			}
	// 			w.Write(resp)
	// 			return
	// 		}
	// }

}

func (s *server) createPlace(w http.ResponseWriter, r *http.Request) {
	var place Place
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Use pointer &place to directly modify the original place variable
	err = json.Unmarshal(body, &place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert place into db
	place.ID = insertPlace(s.db, place)
	// place.ID = nextID
	// nextID += 1

	// places[place.ID] = place
	// places = append(places, place)
	// places[nextID] = place

	resp, err := json.Marshal(place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (s *server) updatePlace(w http.ResponseWriter, r *http.Request) {
	// id, err := strconv.Atoi(r.PathValue("id"))
	idStr := r.URL.Path[len("/places/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	var updatedPlace Place
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &updatedPlace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = updatePlaceInDB(s.db, updatedPlace, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) deletePlace(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	err = deletePlaceFromDB(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	// places = append(places[:index], places[index+1:]...)
	// delete(places, id)
}

func createPlaceTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS place (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	address VARCHAR(100) NOT NULL,
	description TEXT,
	created timestamp DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

}

func insertPlace(db *sql.DB, place Place) int {
	query := `INSERT INTO places (name, address, description)
	VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := db.QueryRow(query, place.Name, place.Address, place.Description).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Inserted ID=", id)
	return id
}

func updatePlaceInDB(db *sql.DB, place Place, id int) error {
	query := `UPDATE place SET name=$1, address=$2, description=$3 WHERE id=$4`
	_, err := db.Exec(query, place.Name, place.Address, place.Description, id)
	fmt.Print("Updated", place)
	return err
}

func deletePlaceFromDB(db *sql.DB, id int) error {
	query := `DELETE FROM place WHERE id=$1`
	_, err := db.Exec(query, id)
	fmt.Print("Deleted ID=", id)

	return err
}

// Data structures: Map (id), Struc
// Unit testing
