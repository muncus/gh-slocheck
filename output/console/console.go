package console

import (
	"fmt"
	"io"
	"time"

	gloss "github.com/charmbracelet/lipgloss"
	"github.com/muncus/gh-slocheck/output"
	"github.com/muncus/gh-slocheck/search"
)

// Status indicators, used in various single-character status displays:
var StatusGood = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#00ff00")).Render("🗸")
var StatusBad = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("✖️")
var StatusPending = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#FF8000")).Render("️•")
var StatusUnknown = gloss.NewStyle().Bold(true).Foreground(gloss.Color("#ff0000")).Render("⁉")

var titleStyle = gloss.NewStyle().Bold(true)
var lastUpdatedStyle = gloss.NewStyle().Faint(true).Align(gloss.Right)
var outOfSLOHeaderStyle = lastUpdatedStyle.Copy().Foreground(gloss.Color("#ff0000")).Bold(true)

func New(t time.Duration) ConsoleFormatter {
	return ConsoleFormatter{
		warnTime: t,
	}
}

type ConsoleFormatter struct {
	warnTime time.Duration
}

// Print outputs the string version of this PRInfo to the provided Writer.
func (c *ConsoleFormatter) Print(w io.Writer, p search.PRInfo) error {
	_, err := w.Write([]byte(c.ToString(p)))
	return err
}

// ToString returns the string representation of the provided PRInfo.
func (c *ConsoleFormatter) ToString(p search.PRInfo) string {
	return fmt.Sprintf("C:%s R:%s (%s) -- %s\n  %s\n  %s\n",
		statusRollupSigil(p), reviewSigil(p), c.lastUpdate(p), getSlug(p), p.Title, p.URL)
}
func getSlug(p search.PRInfo) string {
	return titleStyle.Render(fmt.Sprintf("%s#%d", p.BaseRepository.Name, p.Number))
}

func reviewSigil(p search.PRInfo) string {
	switch p.ReviewDecision {
	case "APPROVED":
		return StatusGood
	case "REVIEW_REQUIRED":
		return StatusPending
	case "CHANGES_REQUESTED":
		return StatusBad
	}
	return StatusUnknown
}

func statusRollupSigil(p search.PRInfo) string {
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

// Ensure we meet the output.Formatter interface.
var _ output.Formatter = &ConsoleFormatter{}
