package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
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

func main() {
	// Postgres connection
	os.Setenv("DATABASE_URL", "postgres://postgres:mysecretpassword@localhost:5432/petplaces?sslmode=disable")
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL can't be found")
}

	srv := &server{}

	// short var declaration := does not work with struct srv.db
	var err error
	srv.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.db.Close()

	// db connection pooling???

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



// Validate input to improve security
func validatePlace(place *Place) error {
	if place.Name == "" {
		return errors.New("Name is required")
}
if len(place.Name) > 100 {
		return errors.New("Name cannot exceed 100 characters")
}
if place.Address == "" {
		return errors.New("Address is required")
}
if len(place.Address) > 200 {
		return errors.New("Address cannot exceed 200 characters")
}
if len(place.Description) > 500 {
		return errors.New("Description cannot exceed 500 characters")
}
return nil
}

func (s *server) getPlaces(w http.ResponseWriter, _ *http.Request) {
	// Using db instead of local memory
	places, err := getPlacesDB(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(places)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (s *server) getPlace(w http.ResponseWriter, r *http.Request, id int) {

	place, err := getPlaceDB(s.db, id)
	if err != nil {
		http.Error(w, "ID not found", http.StatusBadRequest)
	}

	resp, err := json.Marshal(place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
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

	if err := validatePlace(&place); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
}

	// Insert place into db
	place.ID, err = createPlaceDB(s.db, place)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	if err := validatePlace(&updatedPlace); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
}

	err = json.Unmarshal(body, &updatedPlace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = updatePlaceDB(s.db, updatedPlace, id)
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

	err = deletePlaceDB(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func createPlaceTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS places (
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



// create postman requests for create/getPlace
// move sql query into func for getplace and getplaces
// create db.go and move db funcs into that file

