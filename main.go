package main

import (
	"flag"
	"fmt"
	"time"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

var searchflag = flag.String("s", "review-requested:@me is:open archived:false", "a github search string")
var warntime = flag.Duration("w", 4*24*time.Hour, "time duration after which to warn that a PR needs attention")
var limitflag = flag.Int("limit", 30, "maximum number of results to return")

// Status indicators, used in various single-character status displays:
var StatusGood = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#00ff00")).Render("🗸")
var StatusBad = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("✖️")
var StatusPending = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#FF8000")).Render("️•")
var StatusUnknown = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("⁉")

var titleStyle = gloss.NewStyle().Bold(true).PaddingLeft(4)
var urlStyle = gloss.NewStyle().PaddingLeft(4)
var lastUpdatedStyle = gloss.NewStyle().Foreground(gloss.AdaptiveColor{Dark: "#222222", Light: "#aaaaaa"})
var headerStyle = gloss.NewStyle().PaddingLeft(1).Width(40)
var outOfSLOHeaderStyle = headerStyle.Copy().Foreground(gloss.Color("#ff0000")).Bold(true)
var reviewIndicatorStyle = gloss.NewStyle().Width(20).Align(gloss.Left)

func main() {
	flag.Parse()
	prs, err := SearchPRs(*searchflag)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	for _, p := range prs {
		// fmt.Println(p)
		printPRInfo(p)
	}

}

func printPRInfo(p PRInfo) {
	var style = headerStyle
	if p.SinceLastUpdate() >= *warntime {
		style = outOfSLOHeaderStyle
	}
	slug := fmt.Sprintf("%s/%d", p.BaseRepository.Name, p.Number)
	fmt.Print(p.StatusRollupRune())
	fmt.Print(style.Render(slug) + reviewIndicatorStyle.Render("R:"+p.ReviewStatus()) + lastUpdatedStyle.Render(p.LastUpdatedStatus()) + "\n")
	fmt.Println(titleStyle.Render(p.Title))
	fmt.Println(urlStyle.Render(p.URL))
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

func (p PRInfo) ReviewStatus() string {
	switch p.ReviewDecision {
	case "APPROVED":
		return StatusGood
	case "REVIEW_REQUIRED":
		return StatusPending
	case "CHANGES_REQUESTED":
		return StatusBad
	default:
		return StatusUnknown
	}
}

func (p PRInfo) StatusRollupRune() string {
	// https://docs.github.com/en/graphql/reference/enums#statusstate
	switch p.Commits.Nodes[0].Commit.StatusCheckRollup.State {
	case "SUCCESS":
		return StatusGood
	case "FAILURE", "ERROR":
		return StatusBad
	case "PENDING", "EXPECTED":
		return StatusPending
	}
	return StatusUnknown
}

func (p PRInfo) SinceLastUpdate() time.Duration {
	return time.Since(p.UpdatedAt)
}

func (p PRInfo) LastUpdatedStatus() string {
	displaystring := fmt.Sprintf("%0.1fd", p.SinceLastUpdate().Hours()/24)
	switch t := p.SinceLastUpdate(); {
	case t > *warntime:
		return fmt.Sprintf("(%s)🚨", displaystring)
	default:
		return fmt.Sprintf("(%s)", displaystring)

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
