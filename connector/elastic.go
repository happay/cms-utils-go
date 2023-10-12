package connector

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/happay/cms-utils-go/logger"
	"github.com/happay/cms-utils-go/util"
	"github.com/olivere/elastic/v7"
	"gopkg.in/yaml.v2"
)

// ============ Constants =============

// elastic config params
const (
	ElasticUrl = "url"
)

// =========== Exposed (public) Methods - can be called from external packages ============

// GetElasticConnection provides a client to elastic search which can be used for insertion, search, removal etc. operations
func GetElasticSearchConnection(esCredPath, esConfigKey string, index string) (*elastic.Client, error) {
	var err error
	err = initElasticConnectionAndIndexes(esCredPath, esConfigKey, index)
	return elasticClient, err
}

// GetElasticSearchData get the data from elastic search
func GetElasticSearchData(elasticClient *elastic.Client, index string, query *elastic.BoolQuery) (searchResult *elastic.SearchResult, err error) {
	searchResult, err = elasticClient.Search().
		Index(index).
		Query(query).
		Do(context.TODO())
	if err != nil {
		err = fmt.Errorf("failed to query the data | %s", err)
		logger.GetLoggerV3().Error(err.Error())
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

// GetTopElasticSearchData will return the top hit search result
func GetTopElasticSearchData(elasticClient *elastic.Client, index string, query *elastic.BoolQuery) (topHitSearchResult *elastic.SearchHit, err error) {
	searchResult, err := GetElasticSearchData(elasticClient, index, query)
	if err != nil {
		err = fmt.Errorf("failed to query the data | %s", err)
		return
	}
	if searchResult.TotalHits() != 1 {
		err = fmt.Errorf("unique data not found | %s", err)
		return
	}
	topHitSearchResult = searchResult.Hits.Hits[0]
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

var elasticClient *elastic.Client // singleton instance of elastic search client

/*
initElasticConnectionAndIndexes initializes a global singleton client for elastic search.
Additionally, it also checks if the indexes already exists, and creates them otherwise.
*/
func initElasticConnectionAndIndexes(esCredPath, esConfigKey string, index string) (err error) {

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

	// checking and creating indexes, if required
	err = createIndex(index)
	if err != nil {
		return
	}
	return
}

// createIndex checks if the indexName is already exists, and create it otherwise
func createIndex(indexName string) (err error) {

	// check if the index already exist
	index, err := elasticClient.Search(indexName).Do(context.Background())
	if err != nil { // index doesn't exist
		reason := fmt.Sprintf("error checking if %s index already exist: %s", indexName, err)
		logger.GetLoggerV3().Info(reason)
	} else if index.Shards.Successful < index.Shards.Total { // index already exist, but some shards are unavailable
		reason := fmt.Sprintf("%s index already exist, but only %d shards are available out of %d", indexName,
			index.Shards.Successful, index.Shards.Total)
		logger.GetLoggerV3().Info(reason)
		//	TODO: Should we raise a panic here?
	} else {
		reason := fmt.Sprintf("%s index already exist, so skipping creation", indexName)
		logger.GetLoggerV3().Info(reason)
		return
	}

	// as index is non-existent, so creating it
	var result *elastic.IndicesCreateResult
	result, err = elasticClient.CreateIndex(indexName).Do(context.Background())
	if err != nil {
		err = fmt.Errorf("%s index creation fails: %s", indexName, err)
		logger.GetLoggerV3().Error(err.Error())
		return
	}

	// checking if the index creating request is successfully acknowledged by elastic search
	if !result.Acknowledged {
		err = fmt.Errorf("%s index creation is not acknowledged by elastic search", indexName)
		logger.GetLoggerV3().Error(err.Error())
		return
	}
	return
}
