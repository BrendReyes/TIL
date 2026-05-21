package srs

import (
	"context"
	"time"

	"github.com/brendreyes/til/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

func timeNow() time.Time { return time.Now().UTC() }

// ---------------------------------------------------------------------------
// Message types returned by fetch commands
// ---------------------------------------------------------------------------

type entriesFetchedMsg struct {
	entries []database.Entry
	err     error
}

type dueEntriesFetchedMsg struct {
	entries []database.Entry
	err     error
}

type entrySavedMsg struct {
	body string
	tag  string
	err  error
}

type statsFetchedMsg struct {
	total      int64
	reviewed   int64
	unreviewed int64
	due        int64
	byTag      []database.CountEntriesByTagRow
	err        error
}

type tagsFetchedMsg struct {
	tags []database.CountEntriesByTagRow
	err  error
}

type tagDeletedMsg struct {
	tag   string
	count int64
	err   error
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

func fetchEntriesCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		entries, err := db.ListAllEntry(context.Background())
		return entriesFetchedMsg{entries: entries, err: err}
	}
}

func fetchDueEntriesCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		entries, err := db.GetDueEntries(context.Background())
		return dueEntriesFetchedMsg{entries: entries, err: err}
	}
}

type reviewResetMsg struct {
	count int64
	err   error
}

func resetReviewsCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		count, err := db.ResetAllReviews(context.Background(), timeNow())
		return reviewResetMsg{count: count, err: err}
	}
}

func saveEditCmd(db *database.Queries, id int64, body, tag string) tea.Cmd {
	return func() tea.Msg {
		err := db.EditEntry(context.Background(), database.EditEntryParams{
			ID:        id,
			Body:      body,
			Tag:       tag,
			UpdatedAt: timeNow(),
		})
		return entrySavedMsg{body: body, tag: tag, err: err}
	}
}

func fetchTagsCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		tags, err := db.CountEntriesByTag(context.Background())
		return tagsFetchedMsg{tags: tags, err: err}
	}
}

func deleteByTagCmd(db *database.Queries, tag string) tea.Cmd {
	return func() tea.Msg {
		count, err := db.DeleteEntriesByTag(context.Background(), tag)
		return tagDeletedMsg{tag: tag, count: count, err: err}
	}
}

func fetchStatsCmd(db *database.Queries) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		total, err := db.CountAllEntries(ctx)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		reviewed, err := db.CountReviewedEntries(ctx)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		unreviewed, err := db.CountUnreviewedEntries(ctx)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		due, err := db.CountDueEntries(ctx)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		byTag, err := db.CountEntriesByTag(ctx)
		if err != nil {
			return statsFetchedMsg{err: err}
		}
		return statsFetchedMsg{
			total:      total,
			reviewed:   reviewed,
			unreviewed: unreviewed,
			due:        due,
			byTag:      byTag,
		}
	}
}