package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/happay/cms-utils-go/util"
	"github.com/olivere/elastic/v7"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	OpenSearchUrl = "url"
	TypeDoc       = "doc"
)

type Configuration struct {
	Settings struct {
		Index struct {
			NumberOfShards   int `json:"number_of_shards"`
			NumberOfReplicas int `json:"number_of_replicas"`
		} `json:"index"`
	} `json:"settings"`
}

//to get the credentials for Os connection
type CredentialConfiguration struct {
	OsCredPath    string `json:"osCredPath"`
	OsConfigKey   string `json:"osConfigKey"`
	ShardCount    int    `json:"shardCount"`
	ReplicasCount int    `json:"replicasCount"`
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
		err = fmt.Errorf("%s index creation is not acknowledged by open search", indexName)
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

func PostResponseOpenSearch(serviceName, appId, reqId string, logEntry map[string]interface{}, osConfiguration CredentialConfiguration) (err error) {
	index := serviceName + strings.ToLower(time.Now().Month().String()) + "-" + strconv.Itoa(time.Now().Year())
	openSearchClient, err := GetOpenSearchConnection(osConfiguration.OsCredPath, osConfiguration.OsConfigKey, index, osConfiguration.ShardCount, osConfiguration.ReplicasCount)
	if err != nil {
		err = fmt.Errorf("failed to get connection with open search")
		return
	}
	_, err = openSearchClient.Index().
		Index(index).
		Type(TypeDoc).
		BodyJson(logEntry).
		Do(context.TODO())
	if err != nil {
		err = fmt.Errorf("error while uploading Req-Response body to ES for app %s and request %s: %s",
			appId, reqId, err)
		return
	}
	return
}
func GetResponseOpenSearch(serviceName, appId, reqId string, osConfiguration CredentialConfiguration) (searchResult *elastic.SearchResult, err error) {
	index := serviceName + strings.ToLower(time.Now().Month().String()) + "-" + strconv.Itoa(time.Now().Year())
	os, _ := GetOpenSearchConnection(osConfiguration.OsCredPath, osConfiguration.OsConfigKey, index, osConfiguration.ShardCount, osConfiguration.ReplicasCount)
	if os == nil {
		return
	}
	appIdQuery := elastic.NewTermQuery("AppId.keyword", appId)
	reqIdQuery := elastic.NewTermQuery("RequestId.keyword", reqId)
	query := elastic.NewBoolQuery().Must(appIdQuery, reqIdQuery)
	searchResult, err = elasticClient.Search().
		Index(index).
		Query(query).
		Do(context.TODO())
	if err != nil {
		err = fmt.Errorf("failed to query the data | %s", err)
		fmt.Println(err)
		return
	}
	return
}
