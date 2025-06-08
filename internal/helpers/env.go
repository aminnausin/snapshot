package helpers

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func ReadEnvFile() {
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}
}

func GetRequiredEnv(name string) (string, error) {
	value, valueExists := os.LookupEnv(name)

	if !valueExists || value == "" {
		log.Fatalf("No %s has been configured.", name)
	}

	return value, nil
}

func GetListEnv(name string) (valueList map[string]struct{}) {
	value, valueExists := os.LookupEnv(name)

	if !valueExists {
		return
	}

	valueList = make(map[string]struct{})

	for _, v := range strings.Split(strings.ToLower(value), ",") {
		parsed := strings.TrimSpace(v)
		if parsed != "" {
			valueList[parsed] = struct{}{}
		}
	}

	return
}

func GetBooleanEnv(name string, defaultValue bool) bool {
	value, valueExists := os.LookupEnv(name)

	if valueExists {
		return strings.ToLower(strings.TrimSpace(value)) != "false"
	}

	return defaultValue
}
