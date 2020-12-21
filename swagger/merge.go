package swagger

func MergeJson(jsonList [][]map[string]interface{}) (resultantJson []map[string]interface{}){
	resultantJson = make([]map[string]interface{},0)
	for _, json := range jsonList{
		for _, value := range json{
			resultantJson = append(resultantJson, value)
		}
	}
	return
}
