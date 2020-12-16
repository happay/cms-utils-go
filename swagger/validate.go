package swagger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/happay/cms-utils-go/util"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"strings"
	"sync"
)

var swaggerApiParseOnce sync.Once
var swaggerApiSpec map[string]interface{}

// ValidateSwagger validate the swagger specs against the request coming from client.
// It takes the following parameters:
// 1. requestBodyMap - Request Body coming in request
// 2. swaggerApiSpecPath - Path of the swagger file
// 3. requestUrlPath - Request URL
// 4. requestMethod - Request method like POST, GET etc
// 5. routerGroups - This will include all the router groups service has
func ValidateSwagger(requestBodyMap map[string]interface{}, swaggerApiSpecPath, requestUrlPath, requestMethod string, routerGroups ...string) (eventName string, err error) {
	requestBodyPresent := true
	if requestBodyMap == nil || len(requestBodyMap) == 0 { // nil or empty request body
		requestBodyPresent = false // as no request body, so no schema validation required
	}

	// leading slash may be missing from the resource URL path as per gin documentation, so adding one explicitly if missing
	if !strings.HasPrefix(requestUrlPath, "/") {
		requestUrlPath = "/" + requestUrlPath
	}
	requestJsonLoader := gojsonschema.NewGoLoader(requestBodyMap)

	var targetSwaggerApiSpec map[string]interface{}
	targetSwaggerApiSpec = getSwaggerApiSpec(swaggerApiSpecPath)

	apiRequestBodySchema, eventName, err := getApiRequestBodySchema(requestUrlPath, requestMethod, requestBodyPresent, targetSwaggerApiSpec, swaggerApiSpecPath, routerGroups...)
	if err != nil {
		reason := fmt.Sprintf("error while fetching API request json schema: %s", err)
		fmt.Println(reason)
		err = errors.New(reason)
		return
	}

	// If no API request body schema is found in the public API swagger specification, search for the internal keyword.
	if apiRequestBodySchema == nil { // as request body is there, a JSON schema need to be present for the same
		//	In case of no schema is found for the request URI, error is not thrown, just a warning log will be added
		//reason := fmt.Sprintf("no json schema found for the request body for URL %s and Method %s",
		//	requestUrlPath, requestMethod)
		//fmt.Println(reason)

		return
	}

	swaggerJsonLoader := gojsonschema.NewGoLoader(apiRequestBodySchema) // fetching schema for the request body using URI
	result, err := gojsonschema.Validate(swaggerJsonLoader, requestJsonLoader)
	if err != nil {
		reason := fmt.Sprintf("error while validating json schema for request: %s", err)
		fmt.Println(reason)
		err = errors.New(reason)
		return
	}
	if !result.Valid() {
		var reasonBuffer bytes.Buffer
		lenErrors := len(result.Errors())
		reasonBuffer.WriteString(fmt.Sprintf("invalid request. %d error(s) found => ", lenErrors))
		for idx, description := range result.Errors() {
			reasonBuffer.WriteString(description.String())
			if idx < lenErrors-1 {
				reasonBuffer.WriteString(" | ")
			}
		}
		reason := reasonBuffer.String()
		fmt.Println(reason)
		err = errors.New(reason)
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

func getSwaggerApiSpec(swaggerApiSpecPath string) map[string]interface{} {
	swaggerApiParseOnce.Do(func() {
		initSwaggerApiSpec(swaggerApiSpecPath)
	})
	return swaggerApiSpec
}

func initSwaggerApiSpec(swaggerApiSpecPath string) {
	swaggerYamlData, err := ioutil.ReadFile(swaggerApiSpecPath)
	if err != nil {
		reason := fmt.Sprintf("error while reading the swagger API specification: %s", err)
		fmt.Println(reason)
		return
	}

	// NOTE: Default unmarshalling of YAML returns a map[interface{}]interface{}, which is little
	// more cumbersome to handle, on the other hand, JSON unmarshal returns the map[string]interface{}
	// So, converting YAML data to JSON data, and then unmarshalling it to get map[string]interface{}
	swaggerJsonData, err := yaml.YAMLToJSON(swaggerYamlData)

	if err = json.Unmarshal(swaggerJsonData, &swaggerApiSpec); err != nil {
		reason := "Error while parsing the swagger API specification."
		fmt.Println(reason)
	}
}

func getApiRequestBodySchema(requestUrlPath, requestMethod string, requestBodyPresent bool, swaggerApiSpec map[string]interface{}, swaggerApiSpecPath string, routerGroups ...string) (requestSchemaMap map[string]interface{},
	eventName string, err error) {
	swaggerApiSpecPaths, found := swaggerApiSpec["paths"]
	if !found {
		reason := fmt.Sprintf("malformed swagger API spec as paths keys is not found")
		err = errors.New(reason)
		fmt.Println(err)
		return
	}
	found = false
	var schemaUrlPathDetail interface{}
	var schemaUrlPath string

	// checking if any URL path exactly matches with request body
	for schemaUrlPath, schemaUrlPathDetail = range swaggerApiSpecPaths.(map[string]interface{}) {
		if isSameURLPath(schemaUrlPath, requestUrlPath) { // TODO: Fix this to identify b/w keyword versus fix words, may be we can give preference to fix words over placeholder values
			found = true
			break
		}
	}

	// if no exact match is found, checking if any parametric path matches with request URL path
	if !found {
		for schemaUrlPath, schemaUrlPathDetail = range swaggerApiSpecPaths.(map[string]interface{}) {
			if isSameParametricURLPath(schemaUrlPath, requestUrlPath, routerGroups...) {
				found = true
				break
			}
		}
	}

	// if still no match is found, add a log and return w/o any error
	if !found {
		reason := fmt.Sprintf("no json schema is found for schemaUrlPath %s and method %s",
			schemaUrlPath, requestMethod)
		fmt.Println(reason)
		return
	}

	// as match found, so check if the request method exists
	methodDetail, found := schemaUrlPathDetail.(map[string]interface{})[strings.ToLower(requestMethod)]
	if !found {
		reason := fmt.Sprintf("request method %s is not defined for URL Path: %s",
			requestMethod, requestUrlPath)
		fmt.Println(reason)
		return
	}

	// if the request method is found, return the request body schema
	var requestBodyDetail interface{}
	methodDetailMap := methodDetail.(map[string]interface{})
	eventNameInterface, found := methodDetailMap["operationId"]
	if !found {
		reason := fmt.Sprintf("unable to find event name")
		fmt.Println(reason)
		err = errors.New(reason)
		return
	}
	eventName = eventNameInterface.(string)

	requestBodyDetail, found = methodDetailMap["requestBody"]
	// requestBody will be absent for some APIs (possibly GET and others), and thus, requestBodySchema
	// will not be defined for such APIs. So, avoid raising error in such case.
	if requestBodyPresent && !found {
		reason := fmt.Sprintf("requestBody field is missing in the API schema for URL Path: %s",
			requestUrlPath)
		fmt.Println(reason)
		return
	}

	if requestBodyPresent { // only if request body is present, do the schema validation
		requestSchemaMap, err = resolveAndGetRequestBodySchema(requestBodyDetail.(map[string]interface{}), swaggerApiSpecPath)
	}
	return
}

//isSameURLPath checks if the schema's path exactly matches with the request URL path
func isSameURLPath(schemaURLPath, requestURLPath string) bool {
	// remove the "/api" from the schemaURKPath before matching with the requestURLPath
	//schemaURLPath = strings.SplitAfterN(schemaURLPath, "api", 2)[1]
	return schemaURLPath == requestURLPath
}

//isSameParametricURLPath checks if the schema's parametric path matches with the request URL path
func isSameParametricURLPath(schemaURLPath, requestURLPath string, routerGroups ...string) bool {
	// remove the "/api" from the schemaURLPath before matching with the requestURLPath
	splitURL := true
	for _, group := range routerGroups{
		if strings.Contains(schemaURLPath, group){
			splitURL = false
		}
	}
	if splitURL == true{
		schemaURLPath = strings.SplitAfterN(schemaURLPath, "api", 2)[1]
	}
	schemaUrlPathParts := strings.Split(schemaURLPath, "/")
	requestUrlPathParts := strings.Split(requestURLPath, "/")
	if len(schemaUrlPathParts) == len(requestUrlPathParts) {
		for idx, schemaPathPart := range schemaUrlPathParts {
			if schemaPathPart != requestUrlPathParts[idx] { // if not exactly matched
				if strings.HasPrefix(schemaPathPart, "{") &&
					strings.HasSuffix(schemaPathPart, "}") { // check if its a path parameter
					// TODO: currently, no schema type check is checked for path params and assumed as string,
					// ideally it should be checked with path param schema
				} else {
					return false // a mismatch in the URL path snippet
				}
			}
		}
		return true // for loop completed successfully
	}
	return false // unequal length of split URL path arrays
}

// resolveAndGetRequestBodySchema resolves if there is any symlink ($ref used in place of responseBody or
func resolveAndGetRequestBodySchema(requestBodyMap map[string]interface{}, swaggerApiSpecPath string) (requestSchemaMap map[string]interface{}, err error) {
	var actualRequestSchema map[string]interface{}
	// get the referenced component (if specified)
	if contentDetails, found := requestBodyMap["$ref"]; found {
		actualRequestSchema, err = getReferencedComponent(contentDetails.(string), swaggerApiSpecPath)
		if err != nil {
			reason := fmt.Sprintf("fetching referenced component failed: %s", err)
			fmt.Println(reason)
			err = errors.New(reason)
			return
		}
	} else {
		actualRequestSchema = requestBodyMap
	}

	// get the schema under content-type from requestSchema
	if contentDetails, found := actualRequestSchema["content"]; found {
		contentDetailMap := contentDetails.(map[string]interface{})
		applJsonDetail, found := contentDetailMap["application/json"]
		if found {
			applJsonMap := applJsonDetail.(map[string]interface{})
			schemaDetail, found := applJsonMap["schema"]
			if found {
				schemaMap := schemaDetail.(map[string]interface{})
				// checking if the schema is actually a defined or a reference is given
				if _, found := schemaMap["type"]; found {
					requestSchemaMap = schemaMap
				} else if refDetail, found := schemaMap["$ref"]; found { // a schema is present
					refDetailKeys := strings.Split(refDetail.(string), "/")[1:]
					requestSchemaInterface, found := util.GetNestedKeyValue(refDetailKeys, getSwaggerApiSpec(swaggerApiSpecPath))
					if !found {
						reason := fmt.Sprintf("request schema not found")
						fmt.Println(reason)
						err = errors.New(reason)
						return
					} else {
						requestSchemaMap = requestSchemaInterface.(map[string]interface{})
					}
				} else {
					reason := fmt.Sprintf("improper schema defined for the request body")
					fmt.Println(reason)
					err = errors.New(reason)
					return
				}
			}
		}
	} else {
		reason := "invalid schema defined for URL"
		fmt.Println(reason)
		err = errors.New(reason)
	}
	return
}

func getReferencedComponent(refString string, swaggerApiSpecPath string) (referredObject map[string]interface{}, err error) {
	refDetailKeys := strings.Split(refString, "/")[1:] // ignoring the '#'
	requestBodyMap, found := util.GetNestedKeyValue(refDetailKeys, getSwaggerApiSpec(swaggerApiSpecPath))
	if !found {
		reason := fmt.Sprintf("request schema not found")
		fmt.Println(reason)
		err = errors.New(reason)
	} else {
		referredObject = requestBodyMap.(map[string]interface{})
	}
	return
}