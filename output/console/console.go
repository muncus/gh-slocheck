package console

import (
	"fmt"
	"io"
	"time"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/muncus/gh-slocheck/search"
)

// Status indicators, used in various single-character status displays:
var StatusGood = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#00ff00")).Render("🗸")
var StatusBad = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("✖️")
var StatusPending = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#FF8000")).Render("️•")
var StatusUnknown = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("⁉")

var titleStyle = gloss.NewStyle().Bold(true).PaddingLeft(4)
var urlStyle = gloss.NewStyle().PaddingLeft(4)
var lastUpdatedStyle = gloss.NewStyle().Faint(true).Align(gloss.Right)
var headerStyle = gloss.NewStyle().PaddingLeft(1).Width(50).Align(gloss.Left)
var outOfSLOHeaderStyle = lastUpdatedStyle.Copy().Foreground(gloss.Color("#ff0000")).Bold(true)
var reviewIndicatorStyle = gloss.NewStyle().Width(5).Align(gloss.Left)

func New(t time.Duration) ConsoleFormatter {
	return ConsoleFormatter{
		warnTime: t,
	}
}

type ConsoleFormatter struct {
	warnTime time.Duration
}

func (c *ConsoleFormatter) Print(w io.Writer, p search.PRInfo) error {
	return nil
}

func (c *ConsoleFormatter) ToString(p search.PRInfo) string {
	return fmt.Sprintf("C:%s R:%s (%s) -- %s\n  %s\n",
		StatusRollupSigil(p), reviewSigil(p), c.lastUpdate(p), p.Title, p.URL)
}

func reviewSigil(p search.PRInfo) string {
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

func StatusRollupSigil(p search.PRInfo) string {
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

func (c ConsoleFormatter) lastUpdate(p search.PRInfo) string {
	displaystring := fmt.Sprintf("%4.1fd", p.SinceLastUpdate().Hours()/24)
	if p.SinceLastUpdate() > c.warnTime {
		return outOfSLOHeaderStyle.Render(displaystring)
	}
	return lastUpdatedStyle.Render(displaystring)
}
