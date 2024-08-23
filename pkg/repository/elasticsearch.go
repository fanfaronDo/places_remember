package repository

import (
	"fmt"
	cnf "main/pkg/config"

	"github.com/elastic/go-elasticsearch/v8"
)

type Elastic struct {
	client *elasticsearch.Client
}

func NewElastic() (*Elastic, error) {
	cnf := cnf.ReadInConfig()

	host := "http://" + cnf.ElasticHost + ":" + cnf.ElasticPort

	elasticConfig := elasticsearch.Config{
		Addresses: []string{
			host,
		},
	}

	client, err := elasticsearch.NewClient(elasticConfig)

	if err != nil {
		return nil, err
	}

	return &Elastic{
		client: client,
	}, nil
}

func (s *Elastic) GetPoolElements(index string, limit int, offset int) ([]byte, error) {
	// Execute the Elasticsearch query
	searchResult, err := s.client.Search(
		s.client.Search.WithIndex(index),
		s.client.Search.WithFrom(offset),
		s.client.Search.WithSize(limit),
		s.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return []byte{}, err
	}
	defer searchResult.Body.Close()

	buf := make([]byte, 4096)
	i, _ := searchResult.Body.Read(buf)

	return buf[:i], nil
}

func (s *Elastic) SearchByLocation(index string, lat float64, lon float64) ([]byte, error) {

	searchResult, err := s.client.Search(
		s.client.Search.WithIndex(index),
		s.client.Search.WithSort(
			`{
				"_geo_distance": {
					"location": {
						"lat": `+fmt.Sprintf("%f", lat)+`,
						"lon": `+fmt.Sprintf("%f", lon)+`
					},
					"order": "asc",
					"unit": "km",
					"mode": "min",
					"distance_type": "arc",
					"ignore_unmapped": true
				}
			}`),
	)
	if err != nil {
		return []byte{}, err
	}

	defer searchResult.Body.Close()
	buf := make([]byte, 4096)
	i, _ := searchResult.Body.Read(buf)

	return buf[:i], nil
}

func (s *Elastic) GetTotalElements(index string) ([]byte, error) {
	searchResult := s.client.Count.WithIndex(index)
	count, err := s.client.Count(searchResult)
	if err != nil {
		return []byte{}, err
	}

	defer count.Body.Close()

	buf := make([]byte, 2048)
	i, _ := count.Body.Read(buf)

	return buf[:i], nil
}
