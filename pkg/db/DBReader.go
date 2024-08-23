package db

import (
	"encoding/json"
	"fmt"
	types "main/internal/domain/place"
	db "main/pkg/repository"
)

type Store interface {
	// returns a list of items, a total number of hits and (or) an error in case of one
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
}

type PlaceStore struct{}

func NewPlaceStore() *PlaceStore {
	return &PlaceStore{}
}

func (store *PlaceStore) GetPlaces(limit int, offset int) ([]types.Place, int, error) {
	places := make([]types.Place, 0)
	elastic, err := db.NewElastic()

	if err != nil {
		return places, 0, err
	}

	placesBytes, err := elastic.GetPoolElements("places", limit, offset)

	if err != nil {
		return places, 0, err
	}

	places = store.parsePlaces(placesBytes)
	recordsCount, err := store.GetTotalElements(placesBytes)

	if err != nil {
		return places, 0, err
	}

	total := recordsCount

	return places, total, nil
}

func (placeStore *PlaceStore) GetRecommend(lat float64, lon float64) ([]types.Place, error) {
	elastic, err := db.NewElastic()
	var places []types.Place

	if err != nil {
		return places, err
	}

	r, err := elastic.SearchByLocation("places", lat, lon)
	if err != nil {
		return places, err
	}

	ss := placeStore.parsePlaces(r)

	return ss, nil
}

func (placeStore *PlaceStore) GetTotalElements(data []byte) (int, error) {
	elastic, err := db.NewElastic()
	if err != nil {
		return 0, err
	}
	totalBytes, err := elastic.GetTotalElements("places")

	if err != nil {
		return 0, err
	}

	return placeStore.parseCount(totalBytes), nil
}

func (placeStore *PlaceStore) parsePlaces(data []byte) []types.Place {
	fmt.Println(string(data))
	defer func() {
		if r := recover(); r != nil {
			fmt.Errorf("Faluer parse response")
		}
	}()

	places := make([]types.Place, 0)
	parser := make(map[string]interface{})

	json.Unmarshal(data, &parser)

	hits := parser["hits"].(map[string]interface{})

	for _, source := range hits["hits"].([]interface{}) {
		place := source.(map[string]interface{})["_source"].(map[string]interface{})

		places = append(places, types.Place{
			Address: place["address"].(string),
			ID:      int(place["id"].(float64)),
			Name:    place["name"].(string),
			Phone:   place["phone"].(string),
			Location: types.Location{
				Latitude:  place["location"].(map[string]interface{})["lat"].(float64),
				Longitude: place["location"].(map[string]interface{})["lon"].(float64),
			},
		})
	}

	return places
}

func (store *PlaceStore) parseCount(data []byte) int {
	parser := make(map[string]interface{})
	json.Unmarshal(data, &parser)

	return int(parser["count"].(float64))
}
