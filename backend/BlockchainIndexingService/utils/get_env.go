package utils

import "os"

func GetEnv(varName string, defaultValue string) string {
	if value, exists := os.LookupEnv(varName); exists {
		return value
	}

	return defaultValue
}
