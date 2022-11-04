package search

import (
	"fmt"
	"sort"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

// PRSet de-dupes PRInfos by URL, allowing for the merging of multiple lists
type PRSet map[string]PRInfo

func (ps PRSet) Add(pr PRInfo) {
	ps[pr.URL] = pr
}
func (ps PRSet) Remove(pr PRInfo) {
	if _, ok := ps[pr.URL]; ok {
		delete(ps, pr.URL)
	}
}
func (ps PRSet) Merge(a []PRInfo) {
	for _, p := range a {
		ps.Add(p)
	}
}
func (ps PRSet) ByLastUpdate() []PRInfo {
	s := make([]PRInfo, 0)
	for _, pr := range ps {
		s = append(s, pr)
	}
	sort.SliceStable(s, func(i, j int) bool { return s[i].UpdatedAt.Before(s[j].UpdatedAt) })
	return s
}

// a subset of information about a Pull Request.
// see https://docs.github.com/en/graphql/reference/objects#pullrequest
type PRInfo struct {
	Number         int
	Title          string
	URL            string
	State          string
	Mergeable      string
	Merged         bool
	ReviewDecision string
	BaseRepository struct {
		Name          string
		NameWithOwner string
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

func SearchPRs(q string, vars map[string]interface{}) (PRSet, error) {
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
	prlist := PRSet{}
	for _, n := range query.Search.Nodes {
		prlist.Add(n.PullRequest)
	}
	return prlist, nil
}
