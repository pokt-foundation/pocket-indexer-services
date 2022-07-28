package environment

import (
	"os"
	"strconv"
)

// GetInt64 gets the env var as an int
func GetInt64(varName string, defaultValue int64) int64 {
	val, ok := os.LookupEnv(varName)
	if !ok {
		return defaultValue
	}

	iVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}

	return iVal
}

// GetString gets the environment var as a string
func GetString(varName string, defaultValue string) string {
	val, _ := os.LookupEnv(varName)
	if val == "" {
		return defaultValue
	}

	return val
}

// GetBool gets the environment var as a bool
func GetBool(varName string, defaultValue bool) bool {
	val, _ := os.LookupEnv(varName)
	if val == "" {
		return defaultValue
	}

	iVal, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}

	return iVal
}
