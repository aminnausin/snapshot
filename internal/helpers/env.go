package helpers

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func readEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetRequiredEnv(name string) (string, error) {
	readEnvFile()

	value, valueExists := os.LookupEnv(name)

	if !valueExists {
		log.Fatalf("No %s has been configured.", name)
	}

	return value, nil
}

func GetListEnv(name string) (valueList map[string]struct{}) {
	readEnvFile()
	value, valueExists := os.LookupEnv(name)

	if !valueExists {
		return
	}

	for _, v := range strings.Split(strings.ToLower(value), ",") {
		parsed := strings.TrimSpace(v)
		if parsed != "" {
			valueList[parsed] = struct{}{}
		}
	}

	return
}

func GetBooleanEnv(name string) bool {
	readEnvFile()
	value, valueExists := os.LookupEnv(name)
	return valueExists && strings.ToLower(strings.TrimSpace(value)) != "false"
}
