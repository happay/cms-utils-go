package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/happay/cms-utils-go/util"
	"github.com/olivere/elastic/v7"
	"gopkg.in/yaml.v2"
)

const (
	OpenSearchUrl = "url"
)

type Configuration struct {
	Settings struct {
		Index struct {
			NumberOfShards   int `json:"number_of_shards"`
			NumberOfReplicas int `json:"number_of_replicas"`
		} `json:"index"`
	} `json:"settings"`
}

// GetOpenSearchConnection provides a client to elastic search which can be used for insertion, search, removal etc. operations
func GetOpenSearchConnection(osCredPath, osConfigKey string, index string, shardsCount int, replicasCount int) (*elastic.Client, error) {
	var err error
	err = initopenSearchConnectionAndIndexes(osCredPath, osConfigKey, index, shardsCount, replicasCount)
	return openSearchClient, err
}

var openSearchClient *elastic.Client // singleton instance of open search client

/*
initOpenSearchConnectionAndIndexes initializes a global singleton client for elastic search.
Additionally, it also checks if the indexes already exists, and creates them otherwise with given shard and replica count.
*/
func initopenSearchConnectionAndIndexes(osCredPath, osConfigKey string, index string, shardsCount int, replicasCount int) (err error) {

	var bytes []byte
	bytes, err = ioutil.ReadFile(osCredPath)
	if err != nil {
		err = fmt.Errorf("error while reading open search config: %s", err)
		return
	}

	// parse read yaml config into map for further processing
	allDatabaseConfigs := make(map[string]map[string]string)
	if err = yaml.Unmarshal(bytes, &allDatabaseConfigs); err != nil {
		err = fmt.Errorf("error while parsing the common database configuration: %s", err)
		return
	}

	// fetch elastic search config
	elasticConfig, found := allDatabaseConfigs[osConfigKey] // getting config ENV vars required for elastic search
	if !found {
		err = fmt.Errorf("open search config not found")
		return
	}

	// initialize elastic search client
	openSearchConnectionURL := fmt.Sprintf("%s", util.GetConfigValue(elasticConfig[OpenSearchUrl]))
	openSearchClient, err = elastic.NewSimpleClient(elastic.SetURL(openSearchConnectionURL)) // connecting to open search, NOTE: sniffing is turned off currently
	if elastic.IsConnErr(err) {
		err = fmt.Errorf("initializing open search client failed: %s", err)
		return
	}

	// checking and creating indexes, if required
	err = CreateIndexWithShardManagement(index, shardsCount, replicasCount)
	if err != nil {
		return
	}
	return
}

// createIndexWithShardManagement checks if the indexName is already exists, and create it otherwise with given count of shards and replicas
func CreateIndexWithShardManagement(indexName string, shardsCount int, replicasCount int) (err error) {

	// check if the index already exist
	if IndexExists(indexName) {
		return
	}
	// as index is non-existent, so creating it
	var result *elastic.IndicesCreateResult
	var config Configuration
	config.Settings.Index.NumberOfShards = shardsCount
	config.Settings.Index.NumberOfReplicas = replicasCount

	requestBody, err := json.Marshal(config)
	if err != nil {
		err = fmt.Errorf("error marshalling shard configuration for index %s | %s", indexName, err)
		fmt.Println(err)
		return
	}

	result, err = openSearchClient.CreateIndex(indexName).Body(string(requestBody)).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("%s index creation fails: %s", indexName, err)
		fmt.Println(err)
		return
	}

	// checking if the index creating request is successfully acknowledged by elastic search
	if result.Acknowledged == false {
		err = fmt.Errorf("%s index creation is not acknowledged by elastic search", indexName)
		fmt.Println(err)
		return
	}
	return
}

func IndexExists(indexName string) bool {
	exists, err := openSearchClient.IndexExists(indexName).Do(context.Background())
	if err != nil {
		reason := fmt.Sprintf("error checking if %s index already exist: %s", indexName, err)
		fmt.Println(reason)
	}
	return exists
}
