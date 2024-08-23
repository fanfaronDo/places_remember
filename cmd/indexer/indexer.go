package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	types "main/internal/domain/place"
	cnf "main/pkg/config"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

var (
	indexName string = "places"
	places    []types.Place
)

func main() {
	conf := cnf.ReadInConfig()

	host := "http://" + conf.ElasticHost + ":" + conf.ElasticPort
	fmt.Println("Connect to elasticsearch:", host)
	// indexName = conf.ElasticIndex

	records := ReadCsvFile("../materials/data.csv")

	elasticConfig := elasticsearch.Config{
		Addresses: []string{
			host,
		},
	}

	client, err := elasticsearch.NewClient(elasticConfig)
	if err != nil {
		panic(err)
	}

	if res, err := client.Indices.Delete([]string{indexName}, client.Indices.Delete.WithIgnoreUnavailable(true)); err != nil || res.IsError() {
		log.Fatalf("Cannot delete index: %s", err)
	}

	code, err := CreateIndex(client)

	if code != 200 || err != nil {
		panic(fmt.Sprintf("Error creating index: %d", code))
	}

	bulk, err := CreateBulk(client)

	if err != nil {
		panic(err)
	}

	bulk, err = AddPlaces(records, bulk)

	if err != nil {
		panic(err)
	}

	defer func(bi esutil.BulkIndexer, ctx context.Context) {
		err := bi.Close(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}(bulk, context.Background())

	fmt.Println("Added record: ", bulk.Stats().NumAdded)
}

func AddPlaces(records [][]string, bulk esutil.BulkIndexer) (esutil.BulkIndexer, error) {
	for id, record := range records {
		if id == 0 {
			continue
		}

		lat, _ := strconv.ParseFloat(record[4], 64)
		lon, _ := strconv.ParseFloat(record[5], 64)

		places = append(places, types.Place{
			ID:      id,
			Name:    record[1],
			Address: record[2],
			Phone:   record[3],
			Location: types.Location{
				Latitude:  lat,
				Longitude: lon,
			},
		})
	}

	for _, place := range places {
		data, err := json.Marshal(&place)
		if err != nil {
			return bulk, err
		}
		err = bulk.Add(context.Background(), esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: strconv.FormatUint(uint64(place.ID), 10),
			Body:       strings.NewReader(string(data)),
		})
		if err != nil {
			return bulk, err
		}
	}

	return bulk, nil
}

func ReadCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = '\t'
	records, err := csvReader.ReadAll()

	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func CreateBulk(client *elasticsearch.Client) (esutil.BulkIndexer, error) {
	bulk, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      indexName,
		Client:     client,
		NumWorkers: 5,
	})
	if err != nil {
		return nil, err
	}

	return bulk, nil
}

func CreateIndex(client *elasticsearch.Client) (int, error) {
	var code int
	var esapi *esapi.Response
	f, err := os.ReadFile("api/schema.json")
	if err != nil {
		fmt.Println("Can`t open file ", err)
		return code, err
	}

	mapping := string(f)
	client.Indices.Create(
		indexName,
		client.Indices.Create.WithContext(context.Background()),
	)
	esapi, err = client.Indices.PutMapping([]string{indexName}, strings.NewReader(mapping))
	if err != nil {
		return code, err
	}

	code = esapi.StatusCode

	return code, nil
}
