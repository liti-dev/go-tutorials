package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := "postgres://postgres:mysecretpassword@localhost:5432/testdb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	createPlaceTable(db)

	return db
}

func TestGetPlaceAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test place into the database
	testPlace := Place{
		Name:        "Test Place",
		Address:     "123 Avenue",
		Description: "abc",
	}
	testPlace.ID = insertPlace(db, testPlace)

	// Setup HTTP Server
	mux := http.NewServeMux()
	

	// Create a request to the /places/{id} endpoint
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/places/%d", testPlace.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

