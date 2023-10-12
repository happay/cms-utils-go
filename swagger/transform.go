package swagger

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/happay/cms-utils-go/logger"
)

const (
	ReferenceKey = "$refValues" //key to identify if items are fetched by reference
)

// TransformSwagger modifies the swagger API spec as per needed by the developer portal
// for rendering the API details. This function will help to transform the swagger yaml file to json.
// It takes one parameter swaggerApiSpecPath which is the path of swagger file
func TransformSwagger(swaggerApiSpecPath string) (modifiedSwaggerSpec []map[string]interface{}, err error) {
	// get the modified version of Swagger API spec
	modifiedSwaggerSpec, err = getModifiedSwaggerSpec(swaggerApiSpecPath)
	if err != nil {
		reason := fmt.Sprintf("error while creating modified swagger spec: %s", err)
		logger.GetLoggerV3().Error(reason)
		err = errors.New(reason)
		return
	}

	modifiedSwaggerSpec, err = makeModifiedSwaggerSpecOrdered(modifiedSwaggerSpec)
	if err != nil {
		reason := fmt.Sprintf("error while creating modified swagger spec: %s", err)
		logger.GetLoggerV3().Error(reason)
		err = errors.New(reason)
		return
	}
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============

// modifies the stored swagger API spec (in config) to render it easily on the developer portal UI
func getModifiedSwaggerSpec(swaggerApiSpecPath string) (modifiedSwaggerSpec []map[string]interface{}, err error) {
	originalSwaggerSpec := getSwaggerApiSpec(swaggerApiSpecPath)
	flattendSpec, err := expandSwaggerSpecRefs(originalSwaggerSpec, swaggerApiSpecPath, false)
	if err != nil {
		logger.GetLoggerV3().Error(err.Error())
		return
	}
	moreFlattenSpec, err := flattenNestedRequestBodySchema(flattendSpec)
	if err != nil {
		logger.GetLoggerV3().Error(err.Error())
		return
	}
	modifiedSwaggerSpec = make([]map[string]interface{}, 0)
	pathsMap, found := moreFlattenSpec["paths"].(map[string]interface{})
	if !found {
		err = fmt.Errorf("malformed swagger json, paths key not found")
		logger.GetLoggerV3().Error(err.Error())
		return
	}
	tagsDescription := originalSwaggerSpec["tags"].([]interface{})
	for path, schema := range pathsMap {
		for method, methodSchema := range schema.(map[string]interface{}) {
			if reflect.TypeOf(methodSchema).Kind() == reflect.Map {
				methodSchemaMap := methodSchema.(map[string]interface{})
				tag := methodSchemaMap["tags"].([]interface{})[0] // getting the tag to create outer level fields
				var groupDetail map[string]interface{}
				groupDetail, found = getExistingTagDetail(modifiedSwaggerSpec, tag.(string))
				if !found {
					groupDetail = make(map[string]interface{})
					groupDetail["name"] = tag
					groupDetail["path"] = strings.ToLower(tag.(string))
					groupDetail["description"] = getTagDescription(tag.(string), tagsDescription)
					groupDetail["methods"] = make([]map[string]interface{}, 0)
					modifiedSwaggerSpec = append(modifiedSwaggerSpec, groupDetail) // adding new group detail
				}
				apiSchema := make(map[string]interface{})
				apiSchema["name"] = methodSchemaMap["summary"]
				apiSchema["path"] = transformStringWithHyphen(strings.ToLower(methodSchemaMap["summary"].(string)))
				apiSchema["type"] = method
				apiSchema["url"] = path
				if _, descFound := methodSchemaMap["description"]; descFound {
					apiSchema["description"] = methodSchemaMap["description"].(string)
				} else {
					apiSchema["description"] = ""
				}
				if _, summaryFound := methodSchemaMap["summary"]; summaryFound {
					apiSchema["summary"] = methodSchemaMap["summary"].(string)
				} else {
					apiSchema["summary"] = ""
				}
				apiSchema["operationId"] = methodSchemaMap["operationId"]
				if _, parametersFound := methodSchemaMap["parameters"]; parametersFound {
					apiSchema["attributes"] = parseParameters(methodSchemaMap["parameters"].([]interface{}))
				} else {
					apiSchema["attributes"] = nil
				}
				requestBody, found := methodSchemaMap["requestBody"]
				if found {
					apiSchema["requestBody"] = parseRequestBody(requestBody.(map[string]interface{}))
					if apiSchema["attributes"] == nil {
						apiSchema["attributes"] = make([]interface{}, 0)
					}
					for _, reqField := range apiSchema["requestBody"].(map[string]interface{})["properties"].([]map[string]interface{}) {
						reqField["in"] = "body" // adding in: body for body fields to be rendered on the UI portal
						apiSchema["attributes"] = append(apiSchema["attributes"].([]interface{}), reqField)
					}
				} else {
					apiSchema["requestBody"] = nil
				}
				responses, found := methodSchemaMap["responses"]
				if found {
					apiSchema["responses"] = parseResponses(responses.(map[string]interface{}))
				} else {
					apiSchema["responses"] = nil
				}
				groupDetail["methods"] = append(groupDetail["methods"].([]map[string]interface{}), apiSchema)
			}
		}
	}
	return
}

// order the API using operationId, and the responses for each API on the basis of HTTP Status Code assosciated with the response
func makeModifiedSwaggerSpecOrdered(modifiedSwaggerSpec []map[string]interface{}) ([]map[string]interface{}, error) {
	for _, tagMethods := range modifiedSwaggerSpec {

		// sorting API as per the lexicographic ordering of operationId
		methodsList := tagMethods["methods"].([]map[string]interface{})
		methodOperationIds := make([]string, 0)
		for _, methodDetail := range methodsList {
			methodOperationIds = append(methodOperationIds, methodDetail["operationId"].(string))
		}
		sort.Strings(methodOperationIds)
		updatedMethodList := make([]map[string]interface{}, 0, len(methodsList))
		for _, methodOperationId := range methodOperationIds {
			for _, methodDetail := range methodsList {
				if methodOperationId == methodDetail["operationId"].(string) {
					updatedMethodList = append(updatedMethodList, methodDetail)
				}
			}
		}
		tagMethods["methods"] = updatedMethodList

		// sorting each API responses as per their HTTP status codes
		methodsList = tagMethods["methods"].([]map[string]interface{})
		for _, methodDetail := range methodsList {
			responses := methodDetail["responses"].([]map[string]interface{})
			responseCodes := make([]string, 0, len(responses))
			for _, response := range responses {
				responseCodes = append(responseCodes, response["status"].(string))
			}
			sort.Strings(responseCodes)
			updatedResponses := make([]map[string]interface{}, 0, len(responses))
			for _, responseCode := range responseCodes {
				for _, response := range responses {
					if responseCode == response["status"].(string) {
						updatedResponses = append(updatedResponses, response)
					}
				}
			}
			methodDetail["responses"] = updatedResponses
		}
	}
	return modifiedSwaggerSpec, nil
}

// expand
// taken from https://github.com/go-swagger/go-swagger/blob/master/cmd/swagger/commands/flatten.go
func expandSwaggerSpecRefs(originalSwaggerSpec map[string]interface{}, swaggerApiSpecPath string, objectSchemaReference bool) (flattendSpec map[string]interface{}, err error) {
	flattendSpec = originalSwaggerSpec
	for key, val := range originalSwaggerSpec {
		if reflect.TypeOf(val) == nil {
			continue
		}
		if reflect.TypeOf(val).Kind() == reflect.String &&
			strings.HasPrefix(val.(string), "#/") { // a internal component reference, so expand it
			flattendSpec[key], err = getReferencedComponent(val.(string), swaggerApiSpecPath)
			if err != nil {
				logger.GetLoggerV3().Error(err.Error())
				panic(err)
			}
			flattendSpec[ReferenceKey] = flattendSpec[key]
			delete(flattendSpec, key)
			if objectSchemaReference {
				for keyInner, valInner := range flattendSpec[ReferenceKey].(map[string]interface{}) {
					flattendSpec[keyInner] = valInner
				}
				delete(flattendSpec, ReferenceKey)
			}
		} else if reflect.TypeOf(val).Kind() == reflect.Map {
			if key != "items" {
				flattendSpec[key], err = expandSwaggerSpecRefs(val.(map[string]interface{}), swaggerApiSpecPath, true)
				if err != nil {
					logger.GetLoggerV3().Error(err.Error())
					panic(err)
				}
			} else {
				flattendSpec[key], err = expandSwaggerSpecRefs(val.(map[string]interface{}), swaggerApiSpecPath, false)
				if err != nil {
					logger.GetLoggerV3().Error(err.Error())
					panic(err)
				}
			}
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			if key == "responses" || key == "requestBody" || key == "401" || key == "403" {
				if reflect.TypeOf((val.([]interface{}))[0]).Kind() == reflect.Map {
					for idx, value := range val.([]interface{}) {
						val.([]interface{})[idx], err = expandSwaggerSpecRefs(value.(map[string]interface{}), swaggerApiSpecPath, true)
						if err != nil {
							logger.GetLoggerV3().Error(err.Error())
							panic(err)
						}
					}
				}
			} else {
				if reflect.TypeOf((val.([]interface{}))[0]).Kind() == reflect.Map {
					for idx, value := range val.([]interface{}) {
						val.([]interface{})[idx], err = expandSwaggerSpecRefs(value.(map[string]interface{}), swaggerApiSpecPath, false)
						if err != nil {
							logger.GetLoggerV3().Error(err.Error())
							panic(err)
						}
					}
				}
			}
		}
	}
	return
}

func flattenNestedRequestBodySchema(partiallyFlattendSpec map[string]interface{}) (flattendSpec map[string]interface{}, err error) {
	for path, pathVal := range partiallyFlattendSpec["paths"].(map[string]interface{}) {
		for method, methodVal := range pathVal.(map[string]interface{}) {
			requestBodyInterface, exist := methodVal.(map[string]interface{})["requestBody"]
			if exist {

				contentMap := requestBodyInterface.(map[string]interface{})["content"].(map[string]interface{})
				var schemaInterfaceMap map[string]interface{}
				if schemaInterface, found := contentMap["application/json"]; found {
					schemaInterfaceMap = schemaInterface.(map[string]interface{})
				} else if schemaInterface, found := contentMap["multipart/form-data"]; found { // handling the multipart/form-data
					schemaInterfaceMap = schemaInterface.(map[string]interface{})
				}
				schemaInterfaceMap = schemaInterfaceMap["schema"].(map[string]interface{})
				if _, propExists := schemaInterfaceMap["properties"]; !propExists {
					for _, schemaNestedVal := range schemaInterfaceMap {
						partiallyFlattendSpec["paths"].(map[string]interface{})[path].(map[string]interface{})[method].(map[string]interface{})["requestBody"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"] = schemaNestedVal // TODO: Bad hack for now
						break
					}
				}
			}
		}
	}
	flattendSpec = partiallyFlattendSpec
	return
}

// check if this tag is already used as the name of the group in the spec
func getExistingTagDetail(modifiedSwaggerSpec []map[string]interface{}, tag string) (groupDetail map[string]interface{}, foundResult bool) {
	for _, groupDetail = range modifiedSwaggerSpec {
		name, found := groupDetail["name"]
		if found && name == tag {
			foundResult = true
			return
		}
	}
	return
}

// get description of the tag
func getTagDescription(tag string, tagsDescriptions []interface{}) (description string) {
	for _, tagDescription := range tagsDescriptions {
		tagDescriptionMap := tagDescription.(map[string]interface{})
		if tagDescriptionMap["name"] == tag {
			description = tagDescriptionMap["description"].(string)
			return
		}
	}
	return
}

// modify a given string with all whitespaces replaced with hyphen
func transformStringWithHyphen(description string) (response string) {
	splitStr := strings.Split(description, " ")
	response = strings.Join(splitStr, "-")
	return
}

// parse the parameters from swagger file
func parseParameters(parameters []interface{}) (parametersParsed []interface{}) {
	parametersParsed = make([]interface{}, len(parameters))
	for idx, parameterFieldMap := range parameters {
		modifiedParamMap := make(map[string]interface{})
		//modifiedParamMap["name"] = parameterFieldMap["name"]
		for keyParam, valParam := range parameterFieldMap.(map[string]interface{}) {
			if reflect.TypeOf(valParam).Kind() == reflect.Map {
				for keyInnerParam, valInnerParam := range valParam.(map[string]interface{}) {
					modifiedParamMap[keyInnerParam] = valInnerParam
				}
			} else {
				modifiedParamMap[keyParam] = valParam
			}
		}
		if _, foundSchema := modifiedParamMap["schema"]; foundSchema {
			for schemaKey, schemaVal := range modifiedParamMap["schema"].(map[string]interface{}) {
				modifiedParamMap[schemaKey] = schemaVal
			}
			delete(modifiedParamMap, "schema")
		}
		parametersParsed[idx] = modifiedParamMap
	}
	return
}

// parse request body from swagger file
func parseRequestBody(requestBody map[string]interface{}) (requestBodyParsed map[string]interface{}) {
	requestBodyParsed = make(map[string]interface{})
	requestBodyParsed["description"] = requestBody["description"]
	requestBodyParsed["required"] = requestBody["required"]
	if contentType, found := requestBody["content"]; found {
		contentTypeMap := contentType.(map[string]interface{})
		for key, val := range contentTypeMap {
			requestBodyParsed["contentType"] = key
			for keyInner, valInner := range (val.(map[string]interface{}))["schema"].(map[string]interface{}) {
				requestBodyParsed[keyInner] = valInner
			}
			if _, propertiesFound := requestBodyParsed["properties"]; propertiesFound {
				if _, reqFound := requestBodyParsed["required"]; reqFound &&
					reflect.TypeOf(requestBodyParsed["required"]).Kind() == reflect.Slice {
					requestBodyParsed["properties"] = parseProperties(requestBodyParsed["properties"].(map[string]interface{}),
						requestBodyParsed["required"].([]interface{}))
					delete(requestBodyParsed, "required")
				} else {
					requestBodyParsed["properties"] = parseProperties(requestBodyParsed["properties"].(map[string]interface{}),
						nil)
				}
			}
		}
	}
	return
}

// parse response from swagger file
func parseResponses(responseBody map[string]interface{}) (responses []map[string]interface{}) {
	responses = make([]map[string]interface{}, 0)
	for key, val := range responseBody {
		responseBodyMap := make(map[string]interface{})
		responseBodyMap["status"] = key
		valMap := val.(map[string]interface{})
		responseBodyMap["description"] = valMap["description"]
		for keyContent, valContentMap := range valMap["content"].(map[string]interface{}) {
			responseBodyMap["contentType"] = keyContent
			for keySchema, schemaMapVal := range (valContentMap.(map[string]interface{}))["schema"].(map[string]interface{}) {
				responseBodyMap[keySchema] = schemaMapVal
			}
		}

		var itemsFound bool
		var propertiesInt interface{}
		if _, ok := responseBodyMap["items"]; ok {
			propertiesInt, itemsFound = responseBodyMap["items"].(map[string]interface{})["properties"]
		}

		if _, propertiesFound := responseBodyMap["properties"]; !propertiesFound && !itemsFound {
			for respKey, respVal := range responseBodyMap {
				if reflect.TypeOf(respVal).Kind() == reflect.Map {
					for schemaKey, schemaVal := range respVal.(map[string]interface{}) {
						responseBodyMap[schemaKey] = schemaVal
					}
					delete(responseBodyMap, respKey)
					break
				}
			}
		}

		if _, propertiesFound := responseBodyMap["properties"]; propertiesFound {
			propertiesInt = responseBodyMap["properties"]
		}

		var modifiedProperties []map[string]interface{}
		if propertiesInt != nil {
			var requestSlice interface{}
			if itemsFound {
				requestSlice, _ = responseBodyMap["items"].(map[string]interface{})["required"]
				delete(responseBodyMap["items"].(map[string]interface{}), "required")
			} else {
				requestSlice, _ = responseBodyMap["required"]
				delete(responseBodyMap, "required")
			}

			if requestSlice != nil {
				modifiedProperties = parseProperties(propertiesInt.(map[string]interface{}),
					requestSlice.([]interface{}))
			} else {
				modifiedProperties = parseProperties(propertiesInt.(map[string]interface{}), nil)
			}

			if itemsFound {
				itemValues := make([]map[string]interface{}, 0)
				itemValues = append(itemValues, map[string]interface{}{"properties": modifiedProperties,
					"type": responseBodyMap["items"].(map[string]interface{})["type"]})
				responseBodyMap["items"] = itemValues
			} else {
				responseBodyMap["properties"] = modifiedProperties
			}
		}
		responses = append(responses, responseBodyMap)
	}
	return
}

// parse the properties from swagger file
func parseProperties(properties map[string]interface{}, required []interface{}) (parsedProperty []map[string]interface{}) {
	parsedProperty = make([]map[string]interface{}, 0)
	for keyProp, valProp := range properties {
		newProp := make(map[string]interface{})
		newProp["name"] = keyProp
		if reflect.TypeOf(valProp) != nil && reflect.TypeOf(valProp).Kind() == reflect.Map {
			for propKey, propVal := range valProp.(map[string]interface{}) {
				if reflect.TypeOf(propVal) == nil {
					newProp[propKey] = nil
				} else if reflect.TypeOf(propVal).Kind() == reflect.Map {
					if _, found := propVal.(map[string]interface{})[ReferenceKey]; !found && propKey == "items" {
						propVal = map[string]interface{}{keyProp: propVal}
					}
					newProp[propKey] = parseProperties(propVal.(map[string]interface{}), nil)
				} else { // string
					newProp[propKey] = propVal
				}
			}
		}
		newProp["required"] = false // by default, each field will be non-required, unless explicitly stated otherwise
		parsedProperty = append(parsedProperty, newProp)
	}

	// adding required flag inside properties attributes
	if required != nil {
		for _, reqFieldName := range required {
			for _, propertyFieldMap := range parsedProperty {
				if propertyFieldMap["name"] == reqFieldName.(string) {
					propertyFieldMap["required"] = true
				}
			}
		}

	}
	return
}
