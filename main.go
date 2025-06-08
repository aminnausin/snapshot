package main

import (
	"log"
	"os"
	"snapshot/internal/helpers"
	"snapshot/internal/snapshot"
	"strings"

	"github.com/dustin/go-humanize"
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

	output = strings.Replace(output, "{{ stars }}", humanize.Comma(int64(snapshot.GetStargazers(s))), 1)
	output = strings.Replace(output, "{{ forks }}", humanize.Comma(int64(snapshot.GetForks(s))), 1)
	output = strings.Replace(output, "{{ contributions }}", humanize.Comma(int64(snapshot.GetContributions(s))), 1)
	output = strings.Replace(output, "{{ lines_changed }}", humanize.Comma(snapshot.GetLinesChanged(s)), 1)
	output = strings.Replace(output, "{{ repos }}", humanize.Comma(int64(len(snapshot.GetRepos(s)))), 1)
	output = strings.Replace(output, "{{ views }}", humanize.Comma(int64(snapshot.GetViews(s))), 1)

	if s.IncludeProfileViews {
		output = strings.Replace(output, "{{ profile_views }}", humanize.Comma(int64(snapshot.GetProfileViews(s))), 1)
		output = strings.Replace(output, ` class="hide-profile-views"`, "", 1)
		output = strings.Replace(output, ` height="210"`, ` height="234"`, 1)
	}

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

	if s.IncludeProfileViews {
		output = strings.Replace(output, ` height="210"`, ` height="234"`, 1)
	}

	overview := []byte(output)
	werr := os.WriteFile(outputPath, overview, 0644)
	check(werr)
}

func main() {
	validateOutputDir()
	helpers.ReadEnvFile()

	accessToken, err1 := helpers.GetRequiredEnv("ACCESS_TOKEN")
	user, err2 := helpers.GetRequiredEnv("GITHUB_ACTOR")

	if err1 != nil || err2 != nil {
		log.Fatal("Failed loading required ENV")
	}

	excludedRepos := helpers.GetListEnv("EXCLUDED_REPOS")
	excludedLangs := helpers.GetListEnv("EXCLUDED_LANGS")

	includeForkedRepos := helpers.GetBooleanEnv("INCLUDE_FORKED_REPOS", false)
	includeExternalRepos := helpers.GetBooleanEnv("INCLUDE_EXTERNAL_REPOS", false)
	includeProfileViews := helpers.GetBooleanEnv("INCLUDE_PROFILE_VIEWS", false)

	s := snapshot.NewSnapshot(
		user,
		accessToken,
		excludedRepos,
		excludedLangs,
		includeForkedRepos,
		includeExternalRepos,
		includeProfileViews,
	)

	snapshot.GetRepos(&s)
	snapshot.GetProfileViews(&s)
	generateOverview(&s)
	generateLanguages(&s)
}
