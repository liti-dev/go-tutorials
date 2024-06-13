package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()
	router.GET("/places", getPlaces)

	router.Run("localhost:8080")
}

// getPlaces responds with the list of all places as JSON.
func getPlaces(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, places)
}