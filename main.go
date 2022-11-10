package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	graphql "github.com/cli/shurcooL-graphql"
	"github.com/muncus/gh-slocheck/output/console"
	"github.com/muncus/gh-slocheck/search"
)

type StringList []string

func (v *StringList) Set(s string) error {
	*v = append(*v, s)
	return nil
}
func (v *StringList) String() string {
	return fmt.Sprintf("%#v", v)
}

var defaultquery = "review-requested:@me is:open archived:false"

func main() {
	var searchflag = &StringList{}
	flag.Var(searchflag, "s", "a github search string")
	var warntime = flag.Duration("w", 4*24*time.Hour, "time duration after which to warn that a PR needs attention")
	var limitflag = flag.Int("limit", 30, "maximum number of results to return")
	flag.Parse()
	if len(*searchflag) == 0 {
		searchflag = &StringList{defaultquery}
	}
	extra_search_args := os.Getenv("GH_SLOCHECK_SEARCH_EXTRAS")
	results := search.PRSet{}
	for _, query := range *searchflag {
		searchstr := strings.Join([]string{query, extra_search_args}, " ")
		prs, err := search.SearchPRs(searchstr, map[string]interface{}{
			"limit": graphql.Int(*limitflag),
		})
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		for _, pr := range prs {
			results.Add(pr)
		}
	}
	f := console.New(*warntime)
	for _, p := range results.ByLastUpdate() {
		fmt.Print(f.ToString(p))
	}

}
