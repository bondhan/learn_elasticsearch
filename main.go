package main

import (
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("HOST")},
	}

	es, _ := elasticsearch.NewClient(cfg)
	log.Println(elasticsearch.Version)
	log.Println(es.Info())
}
