package snapshot

import (
	"net/http"
	"snapshot/internal/helpers"

	"github.com/hasura/go-graphql-client"
)

type RepoBase struct {
	NameWithOwner string
	IsFork        bool
	Stargazers    struct {
		TotalCount int
	}
	ForkCount int
}

type RepoWithLanguages struct {
	RepoBase

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

type ReposOverviewQuery struct {
	Viewer struct {
		Login string
		Name  string

		Repositories struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []RepoWithLanguages
		} `graphql:"repositories(first: 100, isFork: false, after: $repoCursor)"`
		RepositoriesContributedTo struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []RepoWithLanguages
		} `graphql:"repositoriesContributedTo(first: 100, includeUserRepositories: false, after: $contribCursor, contributionTypes: [COMMIT, PULL_REQUEST, REPOSITORY, PULL_REQUEST_REVIEW])"`
	} `graphql:"viewer"`
}

type CommitStatsQuery struct {
	Repository struct {
		DefaultBranchRef struct {
			Target struct {
				Commit struct {
					History struct {
						PageInfo struct {
							HasNextPage bool
							EndCursor   graphql.String
						}
						Nodes []struct {
							Additions int
							Deletions int
							Author    struct {
								User struct {
									Login string
								}
							}
						}
					} `graphql:"history(first: 100, after: $commitCursor)"`
				} `graphql:"... on Commit"`
			}
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type ContributionYearsQuery struct {
	Viewer struct {
		ContributionsCollection struct {
			ContributionYears []int
		}
	}
}

type Snapshot struct {
	user                 string
	accessToken          string
	client               *http.Client
	queryClient          *graphql.Client
	excludedRepos        map[string]struct{}
	excludedLangs        map[string]struct{}
	includeForkedRepos   bool
	includeExternalRepos bool
	IncludeProfileViews  bool
	_name                *string
	_stargazers          *int
	_forks               *int
	_totalContributions  *int
	_languages           map[string]*helpers.LangInfo
	_repos               map[string]RepoWithLanguages
	_linesChanged        *[2]int // [0]: Added, [1]: Deleted
	_views               *int
	_profileViews        *int
}

type Contributor struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Weeks []struct {
		Additions int `json:"a"`
		Deletions int `json:"d"`
	} `json:"weeks"`
}
