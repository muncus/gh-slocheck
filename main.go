package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

var searchflag = flag.String("s", "review-requested:@me is:open archived:false", "a github search string")
var warntime = flag.Duration("w", 4*24*time.Hour, "time duration after which to warn that a PR needs attention")
var limitflag = flag.Int("limit", 30, "maximum number of results to return")

func main() {
	flag.Parse()
	prs, err := SearchPRs(*searchflag)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	for _, p := range prs {
		fmt.Println(p)
	}

}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
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
type CheckRun struct {
	Name   graphql.String
	Status graphql.String
}
type StatusContext struct {
	Context graphql.String
	State   graphql.String
	// isRequired needs the PR number.
	// IsRequired bool
}

type Commits struct {
	Nodes []struct {
		Commit struct {
			// Status struct {
			// 	State            string
			// 	CombinedContexts struct {
			// 		Nodes []struct {
			// 			CheckRun      CheckRun      `graphql:"... on CheckRun"`
			// 			StatusContext StatusContext `graphql:"... on StatusContext"`
			// 		}
			// 	} `graphql:"combinedContexts(first: 20)"`
			// }
			StatusCheckRollup struct {
				State string
			}
			// CheckSuites struct {
			// 	Nodes []struct {
			// 		CheckRuns struct {
			// 			Nodes []struct {
			// 				Name       string
			// 				Status     string
			// 				Conclusion string
			// 			}
			// 		} `graphql:"checkRuns(last: 1)"`
			// 	}
			// } `graphql:"checkSuites(last: 30)"`
		}
	}
}

// format a PRInfo object for printing.
func (p PRInfo) String() string {
	return fmt.Sprintf(
		"%s %s/%d: R:%c M:%s %s\n\t%s\n\t%s\n",
		p.StatusRollupRune(),
		p.BaseRepository.Name, p.Number,
		p.ReviewRune(), p.CanMergeRune(), p.LastUpdatedStatus(),
		p.Title, p.URL)
}
func (p PRInfo) ReviewRune() rune {
	switch p.ReviewDecision {
	case "APPROVED":
		return '✅'
	case "REVIEW_REQUIRED":
		return '⏱'
	case "CHANGES_REQUESTED":
		return '🛑'
	default:
		return '❓'
	}
}
func (p PRInfo) StateRune() string {
	switch p.State {
	case "OPEN":
		return "⛏"
	case "MERGED":
		return "✅"
	default:
		return "❓"
	}
}
func (p PRInfo) CanMergeRune() string {
	switch p.Mergeable {
	case "MERGEABLE":
		return "✅"
	default:
		return "❓"
	}
}
func (p PRInfo) StatusRollupRune() string {
	// https://docs.github.com/en/graphql/reference/enums#statusstate
	switch p.Commits.Nodes[0].Commit.StatusCheckRollup.State {
	case "SUCCESS":
		return "✅"
	case "FAILURE", "ERROR":
		return "❌"
	case "PENDING", "EXPECTED":
		return "⏱"
	}
	return "❓"
}

func (p PRInfo) SinceLastUpdate() time.Duration {
	return time.Since(p.UpdatedAt)
}

func (p PRInfo) LastUpdatedStatus() string {
	displaystring := fmt.Sprintf("%0.1fd", p.SinceLastUpdate().Hours()/24)
	switch t := p.SinceLastUpdate(); {
	case t > *warntime:
		return fmt.Sprintf("🚨(%s)", displaystring)
	default:
		return fmt.Sprintf("(%s)", t.Round(time.Hour).String())

	}
}

func SearchPRs(q string) ([]PRInfo, error) {
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
	queryvars := map[string]interface{}{
		"limit": graphql.Int(*limitflag),
		// Note: 'is:pr' is necessary, as the search api returns an empty set without it.
		"query": graphql.String(q) + " is:pr sort:updated-asc",
	}
	err = gqClient.Query("SearchPRs", &query, queryvars)
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
