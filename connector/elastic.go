package connector

import (
	"cms-utils-go/util"
	"context"
	"fmt"
	"gopkg.in/olivere/elastic.v6"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

// ============ Constants =============

// elastic config params
const (
	ElasticUrl                 = "url"
)

// =========== Exposed (public) Methods - can be called from external packages ============

// GetElasticConnection provides a client to elastic search which can be used for insertion, search, removal etc. operations
func GetElasticSearchConnection(esCredPath, esConfigKey string, indices []string) (*elastic.Client, error) {
	var err error
	elasticConnectionInit.Do(func() {
		err = initElasticConnectionAndIndexes(esCredPath, esConfigKey, indices)
	})
	return elasticClient, err
}

// GetElasticSearchData get the data from elastic search
func GetElasticSearchData(elasticClient *elastic.Client, index string, query *elastic.BoolQuery) (searchResult *elastic.SearchResult, err error){
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

// PushElasticSearchData push data to elastic search
func PushElasticSearchData(elasticClient *elastic.Client, index, docType string, data interface{}) (err error) {
	_, err = elasticClient.Index().
		Index(index).
		Type(docType).
		BodyJson(data).
		Do(context.TODO())
	if err != nil {
		err = fmt.Errorf("error while uploading data to ES | %s", err)
		return
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

var elasticClient *elastic.Client   // singleton instance of elastic search client
var elasticConnectionInit sync.Once // "do once" construct for initializing elastic search client

/*
	initElasticConnectionAndIndexes initializes a global singleton client for elastic search.
	Additionally, it also checks if the indexes already exists, and creates them otherwise.
*/
func initElasticConnectionAndIndexes(esCredPath, esConfigKey string, indices []string) (err error) {

	// read all database config yaml file
	//DataBaseCredPath := ProjectRootPath() + LocalPathToEnvVars // path to file where common config ENV vars to be used are listed
	var bytes []byte
	bytes, err = ioutil.ReadFile(esCredPath)
	if err != nil {
		err = fmt.Errorf("error while reading elastic config: %s", err)
		return
	}

	// parse read yaml config into map for further processing
	allDatabaseConfigs := make(map[string]map[string]string)
	if err = yaml.Unmarshal(bytes, &allDatabaseConfigs); err != nil {
		err = fmt.Errorf("error while parsing the common database configuration: %s", err)
		return
	}

	// fetch elastic search config
	elasticConfig, found := allDatabaseConfigs[esConfigKey] // getting config ENV vars required for elastic search
	if !found {
		err = fmt.Errorf("elastic search config not found")
		return
	}

	// initialize elastic search client
	elasticConnectionURL := fmt.Sprintf("%s", util.GetConfigValue(elasticConfig[ElasticUrl]))
	elasticClient, err = elastic.NewSimpleClient(elastic.SetURL(elasticConnectionURL)) // connecting to elastic search, NOTE: sniffing is turned off currently
	if elastic.IsConnErr(err) {
		err = fmt.Errorf("initializing elastic search client failed: %s", err)
		return
	}

	// fetching index names from ENV vars
	for _, index := range indices{
		// checking and creating indexes, if required
		err = createIndex(index)
		if err != nil {
			return
		}
	}
	return
}

// createIndex checks if the indexName is already exists, and create it otherwise
func createIndex(indexName string) (err error) {

	// check if the index already exist
	index, err := elasticClient.Search(indexName).Do(context.Background())
	if err != nil { // index doesn't exist
		reason := fmt.Sprintf("error checking if %s index already exist: %s", indexName, err)
		fmt.Println(reason)
	} else if index.Shards.Successful < index.Shards.Total { // index already exist, but some shards are unavailable
		reason := fmt.Sprintf("%s index already exist, but only %d shards are available out of %d", indexName,
			index.Shards.Successful, index.Shards.Total)
		fmt.Println(reason)
		//	TODO: Should we raise a panic here?
	} else {
		reason := fmt.Sprintf("%s index already exist, so skipping creation", indexName)
		fmt.Println(reason)
		return
	}

	// as index is non-existent, so creating it
	var result *elastic.IndicesCreateResult
	result, err = elasticClient.CreateIndex(indexName).Do(context.Background())
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
