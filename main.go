package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/joho/godotenv/autoload"
)

const (
	IndexName = "olympic-events"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func main() {
	//configure es client
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("HOST")},
	}

	//instantiate new es client
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("fail create client err: %s", err)
	}

	//show version and info
	log.Println(elasticsearch.Version)
	log.Println(esClient.Info())

	//check if index exist
	info, err := esClient.Indices.Get([]string{IndexName})
	if err != nil {
		log.Fatalf("fail get index err: %s", err)
	}
	defer info.Body.Close()

	//delete when exist
	if info.StatusCode == 200 {
		var res *esapi.Response
		if res, err := esClient.Indices.Delete([]string{IndexName}, esClient.Indices.Delete.WithIgnoreUnavailable(true)); err != nil || res.IsError() {
			log.Fatalf("Cannot delete index: %s", err)
		}
		if res != nil {
			res.Body.Close()
		}
	}

	//create index
	err = createIndex(esClient, IndexName, os.Getenv("SOURCE"))
	if err != nil {
		log.Fatalf("fail create index err: %s", err)
	}

	//simple prefix search

}

// createIndex given the client, indexName and the target absolute filename
// will create the index using default analyzers. This is very basic function, it uses
// the worker thread pool pattern. The total go routines is Total CPU NUM - 1
func createIndex(esClient *elasticsearch.Client, indexName, filename string) error {
	defer timeTrack(time.Now(), "createIndex")

	file, err := os.Open(filename)

	if err != nil {
		return err
	}
	defer file.Close()

	inputChannel := make(chan string)

	//create worker routines
	wg := sync.WaitGroup{}

	// for number of CPU - 1 we have, create worker thread
	//for i := 0; i < runtime.NumCPU()-1; i++ {
	for i := 0; i < 1000; i++ {

		// semaphore
		wg.Add(1)

		// thread id/counter
		wID := i + 1
		log.Printf("[WorkerPool] Worker %d has been spawned", wID)

		go func(workerID int, esClient *elasticsearch.Client, indexName string) {
			var res *esapi.Response
			var err error

			// reduce wg semaphore counter
			defer wg.Done()

			// read data from channel
			for data := range inputChannel {

				log.Println("data :", data)

				req := esapi.IndexRequest{
					Index:   indexName,
					Body:    strings.NewReader(data),
					Refresh: "true",
				}

				// Perform the request with the client.
				res, err = req.Do(context.Background(), esClient)
				if err != nil {
					log.Fatalf("Error getting response: %s", err)
				}

				if res.IsError() {
					log.Printf("[%s] Error indexing document body=%s", res.Status(), res.Body)
				} else {
					// Deserialize the response into a map.
					var r map[string]interface{}
					if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
						log.Printf("Error parsing the response body: %s", err)
					} else {
						// Print the response status and indexed document version.
						log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
					}
				}

				res.Body.Close()
			}
		}(wID, esClient, indexName)
	}

	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				//end of file
				break
			}
			return err
		}

		inputChannel <- string(line)

		log.Printf("%s \n", line)
	}

	close(inputChannel)

	log.Printf("wait until process done")
	wg.Wait()

	return nil
}
