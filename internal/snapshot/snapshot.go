package snapshot

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"snapshot/internal/helpers"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

func NewSnapshot(user string, accessToken string, excludedRepos map[string]struct{}, excludedLangs map[string]struct{}, includeForkedRepos bool, includeExternalRepos bool, includeProfileViews bool) Snapshot {
	client := &http.Client{Transport: &helpers.TransportWithToken{
		Token:     accessToken,
		Transport: http.DefaultTransport,
	}}

	queryClient := graphql.NewClient("https://api.github.com/graphql", http.DefaultClient).
		WithRequestModifier(func(r *http.Request) {
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		})

	return Snapshot{
		user:                 user,
		accessToken:          accessToken,
		client:               client,
		queryClient:          queryClient,
		excludedRepos:        excludedRepos,
		excludedLangs:        excludedLangs,
		includeForkedRepos:   includeForkedRepos,
		includeExternalRepos: includeExternalRepos,
		IncludeProfileViews:  includeProfileViews,
		_name:                nil,
		_stargazers:          nil,
		_forks:               nil,
		_totalContributions:  nil,
		_languages:           nil,
		_repos:               nil,
		_linesChanged:        nil,
		_views:               nil,
		_profileViews:        nil,
	}
}

func getViewerName(q *ReposOverviewQuery) *string {
	if q.Viewer.Name != "" {
		return &q.Viewer.Name
	}
	if q.Viewer.Login != "" {
		return &q.Viewer.Login
	}
	name := "No Name"
	return &name
}

// GetStats collects the user's Github statistics based on the repos they own or have contributed to.
// It fills out the Snapshot object's attributes for later use.
func getStats(self *Snapshot) {
	if self._stargazers == nil {
		tmp := 0
		self._stargazers = &tmp
	}
	if self._forks == nil {
		tmp := 0
		self._forks = &tmp
	}
	self._repos = make(map[string]RepoWithLanguages)

	repoCursor := graphql.String("")
	contribCursor := graphql.String("")
	// runKey := rand.Int()
	for {
		statsQuery, cursors := reposOverview(helpers.StringPtrOrNil(repoCursor), helpers.StringPtrOrNil(contribCursor))
		helpers.RunQuery(self.queryClient, statsQuery, cursors)

		self._name = getViewerName(statsQuery)
		repos := statsQuery.Viewer.Repositories.Nodes

		// Include repos contributed to without access rights if IncludeExternalRepos is set to true (default is false)
		if self.includeExternalRepos {
			repos = append(repos, statsQuery.Viewer.RepositoriesContributedTo.Nodes...)
		}

		for _, repo := range repos {

			// Ignore excluded repos
			_, excluded := self.excludedRepos[repo.NameWithOwner]
			if excluded {
				continue
			}

			// Ignore duplicate repos from RepositoriesContributedTo if already seen in Repositories or the the other way around
			_, seen := self._repos[repo.NameWithOwner]
			if seen {
				continue
			}

			// Dont count stats if the repo is not a fork of another one or includeForkedRepos is set to false (default)
			if repo.IsFork && !self.includeForkedRepos {
				continue
			}

			self._repos[repo.NameWithOwner] = repo
			parseRepoLanguages(self, &repo)

			if repo.Stargazers.TotalCount > 0 {
				*self._stargazers += repo.Stargazers.TotalCount
				// log.Printf("Get Stats[%d] -> Added %d Stargazers from %s\n", runKey, repo.Stargazers.TotalCount, repo.NameWithOwner)
			}
			*self._forks += repo.ForkCount
		}
		// Update cursors
		repoCursor = graphql.String(statsQuery.Viewer.Repositories.PageInfo.EndCursor)
		contribCursor = graphql.String(statsQuery.Viewer.RepositoriesContributedTo.PageInfo.EndCursor)

		// Exit if no more pages
		if !statsQuery.Viewer.Repositories.PageInfo.HasNextPage && !statsQuery.Viewer.RepositoriesContributedTo.PageInfo.HasNextPage {
			break
		}
	}

	// # TODO: Improve languages to scale by number of contributions to
	// #       specific filetypes
	total := 0
	for _, info := range self._languages {
		total += info.Size
	}
	for _, info := range self._languages {
		if total > 0 {
			info.Prop = float64(info.Size) * 100.0 / float64(total)
		}
	}
}

func parseRepoLanguages(self *Snapshot, repo *RepoWithLanguages) {
	// Initialise languages
	if self._languages == nil {
		self._languages = make(map[string]*helpers.LangInfo)
	}

	for _, langEdge := range repo.Languages.Edges {

		// Check if language should be excluded
		langName := langEdge.Node.Name
		if _, excluded := self.excludedLangs[strings.ToLower(langName)]; excluded {
			continue
		}
		// If already exists, add to size and occurances
		// Otherwise make new

		if entry, ok := self._languages[langName]; ok {
			entry.Size += langEdge.Size
			entry.Occurrences += 1
			continue
		}

		colour := langEdge.Node.Color

		if colour == "" {
			colour = "#000000"
		}

		self._languages[langName] = &helpers.LangInfo{
			Size:        langEdge.Size,
			Occurrences: 1,
			Colour:      colour,
		}
	}
}

func reposOverview(ownedCursor, contribCursor *string) (*ReposOverviewQuery, map[string]any) {
	query := &ReposOverviewQuery{}

	vars := map[string]any{
		"repoCursor":    graphql.String(""),
		"contribCursor": graphql.String(""),
	}

	if ownedCursor != nil {
		vars["repoCursor"] = graphql.String(*ownedCursor)
	}

	if contribCursor != nil {
		vars["contribCursor"] = graphql.String(*contribCursor)
	}

	return query, vars
}

// ContribsByYearQuery dynamically makes a graphql query for retrieving contribution counts for a given year.
// Returns the built query as a string.
func contribsByYearQuery(year int) string {
	return fmt.Sprintf(`
    year%d: contributionsCollection(
        from: "%d-01-01T00:00:00Z",
        to: "%d-01-01T00:00:00Z"
    ) {
      contributionCalendar {
        totalContributions
      }
    }`, year, year, year+1)
}

// AllContributionsQuery dynamically builds a graphql query to get all the contribution counts for a given list of years.
// Returns the built query as a string.
func allContributionsQuery(years []int) string {
	fragments := make([]string, len(years))
	for i, year := range years {
		fragments[i] = contribsByYearQuery(year)
	}

	return fmt.Sprintf("query {\n  viewer {\n%s\n  }\n}", strings.Join(fragments, "\n"))
}

// Properties
func GetName(self *Snapshot) string {
	if self._name != nil {
		return *self._name
	}

	getStats(self)
	return *self._name
}

func GetStargazers(self *Snapshot) int {
	if self._stargazers != nil {
		return *self._stargazers
	}

	getStats(self)
	return *self._stargazers
}

func GetForks(self *Snapshot) int {
	if self._forks != nil {
		return *self._forks
	}

	getStats(self)
	return *self._forks
}

func GetViews(self *Snapshot) int {
	if self._views != nil {
		return *self._views
	}

	total := 0

	for repo := range self._repos {
		uri := fmt.Sprintf("https://api.github.com/repos/%s/traffic/views", repo)

		response, err := self.client.Get(uri)

		if err != nil {
			continue
		}

		defer response.Body.Close()

		var res struct {
			Count float64 `json:"count"`
		}

		if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
			log.Printf("Failed to decode: %v", err)
			continue
		}

		total += int(res.Count)

	}

	self._views = &total
	return total
}

func GetRepos(self *Snapshot) map[string]RepoWithLanguages {
	if self._repos != nil {
		return self._repos
	}
	getStats(self)
	return self._repos
}

func GetContributions(self *Snapshot) int {
	if self._totalContributions != nil {
		return *self._totalContributions
	}

	tmp := 0
	self._totalContributions = &tmp

	var yearsQuery ContributionYearsQuery

	err := helpers.RunQuery(self.queryClient, &yearsQuery, nil)
	if err != nil {
		log.Fatalf("Failed to get contribution years: %s", err)
	}
	years := yearsQuery.Viewer.ContributionsCollection.ContributionYears

	var result map[string]any

	query := allContributionsQuery(years)

	result, err = helpers.RunRawQuery(self.client, query)
	if err != nil {
		log.Fatalf("Raw Query Failed: %s", err)
	}

	viewer := result["viewer"].(map[string]any)
	total := 0
	for year, v := range viewer {
		contribCollection := v.(map[string]any)
		calendar := contribCollection["contributionCalendar"].(map[string]any)
		contributions := int(calendar["totalContributions"].(float64))
		total += contributions
		log.Printf("Made %d contributions in [%s]", contributions, year)
	}
	return total
}

func GetLinesChanged(self *Snapshot) int64 {
	if self._linesChanged != nil {
		return int64(self._linesChanged[0] + self._linesChanged[1])
	}

	additions := 0
	deletions := 0

	// Get lines changed via REST API (far slower, around 10 seconds per repo, results in slightly different count)
	// for repo := range self._repos {
	// 	uri := fmt.Sprintf("repos/%s/stats/contributors", repo)

	// 	response, err := helpers.RunRestQuery(self.client, uri, nil)

	// 	if err != nil {
	// 		log.Printf("Failed to fetch %s: %v", uri, err)
	// 		continue
	// 	}

	// 	var contributors []Contributor

	// 	if err := json.Unmarshal(response, &contributors); err != nil {
	// 		log.Printf("Failed to parse contributor JSON for %s: %v", repo, err)
	// 		continue
	// 	}

	// 	for _, authorObj := range contributors {
	// 		// Assume user was set in main function from env
	// 		if authorObj.Author.Login != self.user {
	// 			continue
	// 		}

	// 		for _, week := range authorObj.Weeks {
	// 			additions += week.Additions
	// 			deletions += week.Deletions
	// 		}
	// 	}

	// }

	for _, repo := range self._repos {

		_, excluded := self.excludedRepos[repo.NameWithOwner]
		if excluded {
			continue
		}

		owner, name, err := helpers.SplitOwnerRepo(repo.NameWithOwner)

		if err != nil {
			// No owner and repo split was found
			continue
		}

		var cursor *graphql.String = nil
		page := 1

		// log.Printf("Getting total commit info for %s", repo.NameWithOwner)
		for {
			// log.Printf("Page: %d", page)
			var commitQuery CommitStatsQuery
			vars := map[string]any{
				"owner":        graphql.String(owner),
				"name":         graphql.String(name),
				"commitCursor": cursor,
			}

			helpers.RunQuery(self.queryClient, &commitQuery, vars)

			history := commitQuery.Repository.DefaultBranchRef.Target.Commit.History
			for _, commit := range history.Nodes {
				if commit.Author.User.Login == self.user {
					additions += commit.Additions
					deletions += commit.Deletions
				}
			}

			cursor = &history.PageInfo.EndCursor

			if !history.PageInfo.HasNextPage {
				break
			}

			page++
		}
	}

	self._linesChanged = &[2]int{additions, deletions} // [0]=add, [1]=del
	return int64(self._linesChanged[0] + self._linesChanged[1])
}

func GetLanguages(self *Snapshot) map[string]*helpers.LangInfo {
	if self._languages != nil {
		return self._languages
	}
	getStats(self)
	return self._languages
}

func GetProfileViews(self *Snapshot) int {
	if self._profileViews != nil {
		return *self._profileViews
	}

	svg, err := helpers.RunSVGRestQuery(self.client, fmt.Sprintf("https://komarev.com/ghpvc/?username=%s", self.user), nil)
	if err != nil {
		log.Fatal(err)
	}

	decoder := xml.NewDecoder(strings.NewReader(svg))

	textCount := 0
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == "text" {
				textCount++
			}
		case xml.CharData:
			if textCount == 4 { // Match the 4th <text> node (the numeric profile view count)
				text := strings.TrimSpace(string(elem))
				cleaned := strings.ReplaceAll(text, ",", "")
				value, err := strconv.Atoi(cleaned)
				if err == nil {
					self._profileViews = &value
					return value
				}
			}
		}
	}

	return 0
}
