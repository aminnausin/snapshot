package main

import (
	"fmt"
	"log"
	"os"
	"snapshot/internal/helpers"
	"snapshot/internal/snapshot"
	"strconv"
	"strings"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func generateOverview(s *snapshot.Snapshot) {
	dat, err := os.ReadFile("templates/overview.svg")
	check(err)
	output := strings.Replace(string(dat), "{{ name }}", snapshot.GetName(s), 1)
	output = strings.Replace(output, "{{ stars }}", snapshot.GetStargazers(s), 1)
	output = strings.Replace(output, "{{ forks }}", snapshot.GetForks(s), 1)
	output = strings.Replace(output, "{{ contributions }}", snapshot.GetViews(s), 1)
	output = strings.Replace(output, "{{ lines_changed }}", snapshot.GetViews(s), 1)
	output = strings.Replace(output, "{{ repos }}", strconv.Itoa(len(snapshot.GetRepos(s))), 1)
	output = strings.Replace(output, "{{ views }}", snapshot.GetViews(s), 1)

	overview := []byte(output)
	werr := os.WriteFile("generated/overview.svg", overview, 0644)
	check(werr)
}

func generateLanguages(s *snapshot.Snapshot) {
	const templatePath = "templates/languages.svg"
	const outputPath = "generated/languages.svg"

	dat, err := os.ReadFile(templatePath)
	check(err)

	progress := ""
	langList := ""
	sortedLanguages := helpers.SortLanguages(snapshot.GetLanguages(s))
	delay := 50
	for _, entry := range sortedLanguages {
		progress += helpers.BuildProgressHTML(entry)
		langList += helpers.BuildLangListHTML(entry, delay)
		delay += 50
	}

	output := strings.Replace(string(dat), "{{ progress }}", progress, 1)
	output = strings.Replace(output, "{{ lang_list }}", langList, 1)

	overview := []byte(output)
	werr := os.WriteFile(outputPath, overview, 0644)
	check(werr)
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
	fmt.Printf("Repo Views in last 2 weeks: %s\n", x)
	generateOverview(&s)
	generateLanguages(&s)
	// await asyncio.gather(generate_languages(s), generate_overview(s))
}
