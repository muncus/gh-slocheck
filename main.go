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

var searchflag = flag.String("s", "review-requested:@me is:open archived:false", "a github search string")
var warntime = flag.Duration("w", 4*24*time.Hour, "time duration after which to warn that a PR needs attention")
var limitflag = flag.Int("limit", 30, "maximum number of results to return")

func main() {
	flag.Parse()
	extra_search_args := os.Getenv("GH_SLOCHECK_SEARCH_EXTRAS")
	searchstr := strings.Join([]string{*searchflag, extra_search_args}, " ")
	prs, err := search.SearchPRs(searchstr, map[string]interface{}{
		"limit": graphql.Int(*limitflag),
	})
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	f := console.New(*warntime)
	for _, p := range prs {
		fmt.Print(f.ToString(p))
	}

}
