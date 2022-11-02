package search

import (
	"fmt"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

// a subset of information about a Pull Request.
// see https://docs.github.com/en/graphql/reference/objects#pullrequest
type PRInfo struct {
	Number         int
	Title          string
	URL            string
	State          string
	Mergeable      string
	ReviewDecision string
	BaseRepository struct {
		Name string
	}
	Commits   Commits `graphql:"commits(last: 1)"`
	UpdatedAt time.Time
}

type Commits struct {
	Nodes []struct {
		Commit struct {
			StatusCheckRollup struct {
				State string
			}
		}
	}
}

func (p PRInfo) SinceLastUpdate() time.Duration {
	return time.Since(p.UpdatedAt)
}

func SearchPRs(q string, vars map[string]interface{}) ([]PRInfo, error) {
	gqClient, err := gh.GQLClient(nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create GraphQL client: %v", err)
	}
	var query struct {
		Search struct {
			Nodes []struct {
				PullRequest PRInfo `graphql:"... on PullRequest"`
			}
		} `graphql:"search(type: ISSUE, first: $limit, query: $query)"`
	}
	vars["query"] = graphql.String(q) + " is:pr"
	err = gqClient.Query("SearchPRs", &query, vars)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Query results: %#v\n", query)
	prinfos := make([]PRInfo, 0)
	for _, n := range query.Search.Nodes {
		prinfos = append(prinfos, n.PullRequest)
	}
	return prinfos, nil
}
