package util

import "os"


func GetConfigValue(key string) string {
	//TODO: include aws store parameter
	return os.Getenv(key)
}