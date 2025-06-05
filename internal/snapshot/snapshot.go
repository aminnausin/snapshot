package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type LangInfo struct {
	Size        int
	Occurrences int
	Colour      string
	Prop        float64
}

type ReposOverviewQuery struct {
	Viewer struct {
		Login string
		Name  string

		Repositories struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				NameWithOwner string
				Stargazers    struct {
					TotalCount int
				}
				ForkCount int
				Languages struct {
					Edges []struct {
						Size int
						Node struct {
							Name  string
							Color string
						}
					}
				} `graphql:"languages(first: 10, orderBy: {field: SIZE, direction: DESC})"`
			}
		} `graphql:"repositories(first: 100, isFork: false, after: $repoCursor)"`
		RepositoriesContributedTo struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				NameWithOwner string
				Stargazers    struct {
					TotalCount int
				}
				ForkCount int
				Languages struct {
					Edges []struct {
						Size int
						Node struct {
							Name  string
							Color string
						}
					}
				} `graphql:"languages(first: 10, orderBy: {field: SIZE, direction: DESC})"`
			}
		} `graphql:"repositoriesContributedTo(first: 100, includeUserRepositories: false, after: $contribCursor, contributionTypes: [COMMIT, PULL_REQUEST, REPOSITORY, PULL_REQUEST_REVIEW])"`
	} `graphql:"viewer"`
}

type DebugQuery struct {
	Viewer struct {
		Login string
		Name  string

		Repositories struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				NameWithOwner string
				Stargazers    struct {
					TotalCount int
				}
				ForkCount int
				Languages struct {
					Edges []struct {
						Size int
						Node struct {
							Name  string
							Color string
						}
					}
				} `graphql:"languages(first: 10, orderBy: {field: SIZE, direction: DESC})"`
			}
		} `graphql:"repositories(first: 100, isFork: false, after: $repoCursor)"`
		RepositoriesContributedTo struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				NameWithOwner string
				Stargazers    struct {
					TotalCount int
				}
				ForkCount int
				Languages struct {
					Edges []struct {
						Size int
						Node struct {
							Name  string
							Color string
						}
					}
				} `graphql:"languages(first: 10, orderBy: {field: SIZE, direction: DESC})"`
			}
		} `graphql:"repositoriesContributedTo(first: 100, includeUserRepositories: false, after: $contribCursor, contributionTypes: [COMMIT, PULL_REQUEST, REPOSITORY, PULL_REQUEST_REVIEW])"`
	} `graphql:"viewer"`
}

type Snapshot struct {
	user                string
	accessToken         string
	client              *http.Client
	queryClient         *graphql.Client
	excludedRepos       map[string]struct{}
	excludedLangs       map[string]struct{}
	ignoreForkedRepos   bool
	_name               *string
	_stargazers         *int
	_forks              *int
	_totalContributions *int
	_languages          map[string]*LangInfo //*LangInfo
	_repos              map[string]struct{}
	_linesChanged       *[2]int // [0]: Added, [1]: Deleted
	_views              *int

	queries *Queries
}

type Queries struct {
	user              string
	accessToken       string
	client            *http.Client
	excludedRepos     []string
	excludedLangs     []string
	ignoreForkedRepos bool
}

type transportWithToken struct {
	token     string
	transport http.RoundTripper
}

func (t *transportWithToken) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return t.transport.RoundTrip(req)
}

func NewSnapshot(user string, accessToken string, excludedRepos map[string]struct{}, excludedLangs map[string]struct{}, ignoreForkedRepos bool) Snapshot {
	client := &http.Client{Transport: &transportWithToken{
		token:     accessToken,
		transport: http.DefaultTransport,
	}}

	queryClient := graphql.NewClient("https://api.github.com/graphql", http.DefaultClient).
		WithRequestModifier(func(r *http.Request) {
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		})

	return Snapshot{
		user:              user,
		accessToken:       accessToken,
		client:            client,
		queryClient:       queryClient,
		excludedRepos:     excludedRepos,
		excludedLangs:     excludedLangs,
		ignoreForkedRepos: ignoreForkedRepos,
		// queries: Queries(username, access_token, session),

		_name:               nil,
		_stargazers:         nil,
		_forks:              nil,
		_totalContributions: nil,
		_languages:          nil,
		_repos:              nil,
		_linesChanged:       nil,
		_views:              nil,
	}
}

func runQuery(self Snapshot, query *ReposOverviewQuery, variables map[string]interface{}) {
	// var testQuery DebugQuery
	vars := make(map[string]interface{})
	vars["contribCursor"] = ""
	vars["repoCursor"] = ""
	// err := self.queryClient.Query(context.Background(), &testQuery, vars)
	// if err != nil {
	// log.Fatalf("Failed GraphQL: %s", err)
	// }

	// log.Fatalf("%s", testQuery.Viewer.Repositories.Nodes)
	// log.Fatalf("%s", &query)
	// log.Fatalf("%s | %s", vars, variables)

	err := self.queryClient.Query(context.Background(), &query, variables)
	if err != nil {
		log.Fatalf("Failed GraphQL: %s", err)
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

func stringPtrOrNil(v graphql.String) *string {
	if v == "" {
		return nil
	}
	s := string(v)
	return &s
}

func getStats(self *Snapshot) {
	// """
	// Get lots of summary statistics using one big query. Sets many attributes
	// """
	if self._stargazers == nil {
		tmp := 0
		self._stargazers = &tmp
	}
	if self._forks == nil {
		tmp := 0
		self._forks = &tmp
	}
	self._repos = make(map[string]struct{})

	repoCursor := graphql.String("")
	contribCursor := graphql.String("")

	stillSearching := true

	for stillSearching {
		statsQuery, cursors := reposOverview(stringPtrOrNil(repoCursor), stringPtrOrNil(contribCursor))
		runQuery(*self, &statsQuery, cursors)

		//get("data", {}).get("viewer", {}).get("name", None)
		self._name = getViewerName(&statsQuery)

		repos := statsQuery.Viewer.Repositories.Nodes
		if !self.ignoreForkedRepos {
			repos = append(repos, statsQuery.Viewer.RepositoriesContributedTo.Nodes...)
		}
		for _, repo := range repos {
			_, excluded := self.excludedRepos[repo.NameWithOwner]

			if excluded {
				continue
			}

			self._repos[repo.NameWithOwner] = struct{}{}

			if repo.Stargazers.TotalCount > 0 {
				*self._stargazers += repo.Stargazers.TotalCount
			}
			*self._forks += repo.ForkCount
			for _, langEdge := range repo.Languages.Edges {
				// Initialise languages
				if self._languages == nil {
					self._languages = make(map[string]*LangInfo)
				}

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
				} else {
					colour := langEdge.Node.Color
					if colour == "" {
						colour = "#000000"
					}
					self._languages[langName] = &LangInfo{
						Size:        langEdge.Size,
						Occurrences: 1,
						Colour:      langEdge.Node.Color,
					}
				}
			}
		}

		// contrib_repos = (
		//     raw_results.get("data", {})
		//     .get("viewer", {})
		//     .get("repositoriesContributedTo", {})
		// )
		// owned_repos = (
		//     raw_results.get("data", {}).get("viewer", {}).get("repositories", {})
		// )

		// repos = owned_repos.get("nodes", [])
		// if not self._ignore_forked_repos:
		//     repos += contrib_repos.get("nodes", [])

		// for repo in repos:
		//     if repo is None:
		//         continue
		//     name = repo.get("nameWithOwner")
		//     if name in self._repos or name in self._exclude_repos:
		//         continue
		//     self._repos.add(name)
		//     self._stargazers += repo.get("stargazers").get("totalCount", 0)
		//     self._forks += repo.get("forkCount", 0)

		//     for lang in repo.get("languages", {}).get("edges", []):
		//         name = lang.get("node", {}).get("name", "Other")
		//         languages = await self.languages
		//         if name.lower() in exclude_langs_lower:
		//             continue
		//         if name in languages:
		//             languages[name]["size"] += lang.get("size", 0)
		//             languages[name]["occurrences"] += 1
		//         else:
		//             languages[name] = {
		//                 "size": lang.get("size", 0),
		//                 "occurrences": 1,
		//                 "color": lang.get("node", {}).get("color"),
		//             }

		// Update cursors
		repoCursor = graphql.String(statsQuery.Viewer.Repositories.PageInfo.EndCursor)
		contribCursor = graphql.String(statsQuery.Viewer.RepositoriesContributedTo.PageInfo.EndCursor)

		// Exit if no more pages
		if !statsQuery.Viewer.Repositories.PageInfo.HasNextPage && !statsQuery.Viewer.RepositoriesContributedTo.PageInfo.HasNextPage {
			stillSearching = false
		}

		// if owned_repos.get("pageInfo", {}).get(
		//     "hasNextPage", False
		// ) or contrib_repos.get("pageInfo", {}).get("hasNextPage", False):
		//     next_owned = owned_repos.get("pageInfo", {}).get(
		//         "endCursor", next_owned
		//     )
		//     next_contrib = contrib_repos.get("pageInfo", {}).get(
		//         "endCursor", next_contrib
		//     )
		// else{
		stillSearching = false
		// }

	}

	// # TODO: Improve languages to scale by number of contributions to
	// #       specific filetypes
	// langs_total = sum([v.get("size", 0) for v in self._languages.values()])
	// for k, v in self._languages.items():
	//     v["prop"] = 100 * (v.get("size", 0) / langs_total)

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

func reposOverview(ownedCursor, contribCursor *string) (ReposOverviewQuery, map[string]interface{}) {
	query := ReposOverviewQuery{}

	vars := map[string]interface{}{
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

// Properties
func GetName(self *Snapshot) string {
	if self._name != nil {
		return *self._name
	}

	getStats(self)
	return *self._name
}

func GetStargazers(self *Snapshot) string {
	if self._stargazers != nil {
		return strconv.Itoa(*self._stargazers)
	}

	getStats(self)
	return strconv.Itoa(*self._stargazers)
}

func GetForks(self *Snapshot) string {
	if self._forks != nil {
		return strconv.Itoa(*self._forks)
	}

	getStats(self)
	return strconv.Itoa(*self._forks)
}

func GetViews(self *Snapshot) string {
	if self._views != nil {
		return strconv.Itoa(*self._views)
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
	return strconv.Itoa(total)
}

func GetRepos(self *Snapshot) map[string]struct{} {
	if self._repos != nil {
		return self._repos
	}
	getStats(self)
	return self._repos
}

func GetContributions(self *Snapshot) string {
	if self._totalContributions != nil {
		return strconv.Itoa(*self._totalContributions)
	}

	*self._totalContributions = 0

	getStats(self)
	return strconv.Itoa(*self._totalContributions)
}

func GetLinesChanged(self *Snapshot) string {
	if self._linesChanged != nil {
		return strconv.Itoa(self._linesChanged[0] + self._linesChanged[1])
	}

	additions := 0
	deletions := 0

	for repo := range self._repos {
		uri := fmt.Sprintf("https://api.github.com/repos/%s/stats/contributors", repo)

		response, err := self.client.Get(uri)

		if err != nil {
			continue
		}

		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)

		if err != nil {
			log.Printf("Failed to read response body: %v", err)
			continue
		}

		fmt.Printf("Response for %s:\n%s\n", repo, string(body))

	}

	self._linesChanged = &[2]int{additions, deletions}
	return strconv.Itoa(self._linesChanged[0] + self._linesChanged[1])
}

func GetLanguages(self *Snapshot) map[string]*LangInfo {
	if self._languages != nil {
		return self._languages
	}
	getStats(self)
	return self._languages
}
