package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)
type Place struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

// server struct holds the db config and router
type server struct {
	db     *DB
	router *http.ServeMux
}

func main() {
	// Postgres connection
	// errEnv := godotenv.Load()
	// if errEnv != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL can't be found")
}

	// short var declaration := does not work with struct srv.db
	var err error
	db, err := createDB(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.createPlaceTable()
	if err != nil {
		log.Fatal(err)
	}

	// Set up router
	srv := &server{
		db:     db,           
		router: http.NewServeMux(),
	}

	srv.routes()

	slog.Info("Starting on port 8080")
	err = http.ListenAndServe("0.0.0.0:8080", srv.router)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) routes() {
	// Set up versioned routes for v1
	
	s.router.HandleFunc("GET /v1/places", s.getPlaces)
	s.router.HandleFunc("POST /v1/places", s.createPlace)
	s.router.HandleFunc("GET /v1/places/{id}", s.getPlace)
	s.router.HandleFunc("PUT /v1/places/{id}", s.updatePlace)
	s.router.HandleFunc("DELETE /v1/places/{id}", s.deletePlace)
}
// Look into grouping gorilla mux/ gin

// Validate input to improve security
func validatePlace(place *Place) error {
	if place.Name == "" {
		return errors.New("name is required")
}
if len(place.Name) > 100 {
		return errors.New("name cannot exceed 100 characters")
}
if place.Address == "" {
		return errors.New("address is required")
}
if len(place.Address) > 200 {
		return errors.New("address cannot exceed 200 characters")
}
if len(place.Description) > 500 {
		return errors.New("description cannot exceed 500 characters")
}
return nil
}

func (s *server) getPlaces(w http.ResponseWriter, _ *http.Request) {
	// Using db instead of local memory
	places, err := s.db.getPlacesDB()
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

func (s *server) getPlace(w http.ResponseWriter, r *http.Request) {
id, err := strconv.Atoi(r.PathValue("id"))
if err != nil {
	http.Error(w, "ID not found", http.StatusBadRequest)
}

place, err := s.db.getPlaceDB(id)
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
	place.ID, err = s.db.createPlaceDB(place)
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

	err = s.db.updatePlaceDB(updatedPlace, id)
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

	err = s.db.deletePlaceDB(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

