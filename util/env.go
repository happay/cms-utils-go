package util

import "os"

// GetConfigValue get the environment value using the key.
// if not found, then fetches it from AWS Parameter Store
func GetConfigValue(key string) string {
	//TODO: include AWS Parameter Store
	return os.Getenv(key)
}