package main

import (
	"database/sql"
	"fmt"
)

func getPlacesDB(db *sql.DB) ([]Place, error) {
	rows, err := db.Query("SELECT id, name, address, description FROM places")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := []Place{}
	for rows.Next() {
		var place Place
		if err := rows.Scan(&place.ID, &place.Name, &place.Address, &place.Description); err != nil {
			return nil, err
		}
		places = append(places, place)
	}
	return places, nil
}

func getPlaceDB(db *sql.DB, id int) (Place, error) {
	var place Place
	err := db.QueryRow("SELECT id, name, address, description FROM places WHERE id = $1", id).Scan(&place.ID, &place.Name, &place.Address, &place.Description)
	if err == sql.ErrNoRows {
		return Place{}, fmt.Errorf("place not found")
	}
	if err != nil {
		return Place{}, err
	}
	return place, nil
}


func createPlaceDB(db *sql.DB, place Place) (int, error) {
	query := `INSERT INTO places (name, address, description)
	VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := db.QueryRow(query, place.Name, place.Address, place.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting place: %w", err)
	}
	fmt.Print("Inserted ID=", id)
	return id, nil
}

func updatePlaceDB(db *sql.DB, place Place, id int) error {
	query := `UPDATE place SET name=$1, address=$2, description=$3 WHERE id=$4`
	_, err := db.Exec(query, place.Name, place.Address, place.Description, id)
	fmt.Print("Updated", place)
	return err
}

func deletePlaceDB(db *sql.DB, id int) error {
	query := `DELETE FROM places WHERE id=$1`
	_, err := db.Exec(query, id)
	fmt.Print("Deleted ID=", id)

	return err
}