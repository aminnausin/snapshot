package main

import (
	"fmt"
	"log"
	"os"
	"snapshot/internal/helpers"
	"snapshot/internal/snapshot"
)

func validateOutputDir() error {
	dir := "./generated"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	validateOutputDir()

	accessToken, err1 := helpers.GetRequiredEnv("ACCESS_TOKEN")
	user, err2 := helpers.GetRequiredEnv("GITHUB_ACTOR")

	if err1 != nil || err2 != nil {
		log.Fatal("Failed")
	}

	excludedRepos := helpers.GetListEnv("EXCLUDED_REPOS")
	excludedLangs := helpers.GetListEnv("EXCLUDED_LANGS")

	ignoreForkedRepos := helpers.GetBooleanEnv("EXCLUDE_FORKED_REPOS")

	s := snapshot.NewSnapshot(
		user,
		accessToken,
		excludedRepos,
		excludedLangs,
		ignoreForkedRepos,
	)

	snapshot.GetRepos(&s)
	x := snapshot.GetViews(&s)
	fmt.Println(x)
	// await asyncio.gather(generate_languages(s), generate_overview(s))
}
