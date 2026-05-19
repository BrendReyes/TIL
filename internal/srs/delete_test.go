package srs

import (
	"os"
	"testing"

	"github.com/brendreyes/til/internal/database"
)

func fakeStdin(t *testing.T, input string) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("fakeStdin: os.Pipe: %v", err)
	}

	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("fakeStdin: write: %v", err)
	}
	w.Close()

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		r.Close()
	})
}

func TestState_DeleteEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		stdinText string
		wantErr   bool
	}{
		{
			name: "confirms with 'y' deletes existing entry",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				return q
			}()},
			args:      args{id: 1},
			stdinText: "y\n",
			wantErr:   false,
		},
		{
			name: "confirms with 'Y' deletes existing entry",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "BFS uses a queue", "algorithms")
				return q
			}()},
			args:      args{id: 1},
			stdinText: "Y\n",
			wantErr:   false,
		},
		{
			name: "aborts on 'n' leaves entry intact",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "goroutines are cheap", "go")
				return q
			}()},
			args:      args{id: 1},
			stdinText: "n\n",
			wantErr:   false,
		},
		{
			name: "aborts on empty input leaves entry intact",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "goroutines are cheap", "go")
				return q
			}()},
			args:      args{id: 1},
			stdinText: "\n",
			wantErr:   false,
		},
		{
			name:      "confirms on non-existent id returns no error",
			fields:    fields{DB: newTestDB(t)},
			args:      args{id: 99999},
			stdinText: "y\n",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeStdin(t, tt.stdinText)
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.DeleteEntry(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("State.DeleteEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.stdinText == "y\n" || tt.stdinText == "Y\n" {
				_, err := s.DB.GetEntryByID(t.Context(), tt.args.id)
				if err == nil {
					t.Errorf("expected entry %d to be deleted, but it still exists", tt.args.id)
				}
			} else if tt.stdinText == "n\n" || tt.stdinText == "\n" {
				_, err := s.DB.GetEntryByID(t.Context(), tt.args.id)
				if err != nil {
					t.Errorf("expected entry %d to remain, but got error: %v", tt.args.id, err)
				}
			}
		})
	}
}

func TestState_RemoveAllEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	tests := []struct {
		name      string
		fields    fields
		stdinText string
		wantErr   bool
	}{
		{
			name: "confirms with 'y' on populated table returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				seedEntry(t, q, "BFS uses a queue", "algorithms")
				return q
			}()},
			stdinText: "y\n",
			wantErr:   false,
		},
		{
			name:      "confirms with 'y' on empty table returns no error",
			fields:    fields{DB: newTestDB(t)},
			stdinText: "y\n",
			wantErr:   false,
		},
		{
			name: "aborts on 'n' returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				return q
			}()},
			stdinText: "n\n",
			wantErr:   false,
		},
		{
			name: "aborts on empty input returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				return q
			}()},
			stdinText: "\n",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeStdin(t, tt.stdinText)
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.RemoveAllEntry(); (err != nil) != tt.wantErr {
				t.Errorf("State.RemoveAllEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.stdinText == "y\n" {
				count, _ := s.DB.CountAllEntries(t.Context())
				if count != 0 {
					t.Errorf("expected 0 entries, got %d", count)
				}
			}
		})
	}
}

func TestState_RemoveEntryByTag(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		tag string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		stdinText string
		wantErr   bool
	}{
		{
			name: "confirms with 'y' deletes all entries with matching tag",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				seedEntry(t, q, "goroutines are cheap", "go")
				seedEntry(t, q, "BFS uses a queue", "algorithms")
				return q
			}()},
			args:      args{tag: "go"},
			stdinText: "y\n",
			wantErr:   false,
		},
		{
			name: "confirms with 'Y' deletes matching entries",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				return q
			}()},
			args:      args{tag: "go"},
			stdinText: "Y\n",
			wantErr:   false,
		},
		{
			name: "aborts on 'n' leaves entries intact",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				return q
			}()},
			args:      args{tag: "go"},
			stdinText: "n\n",
			wantErr:   false,
		},
		{
			name:      "tag with no matching entries returns no error",
			fields:    fields{DB: newTestDB(t)},
			args:      args{tag: "rust"},
			stdinText: "y\n",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeStdin(t, tt.stdinText)
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.RemoveEntryByTag(tt.args.tag); (err != nil) != tt.wantErr {
				t.Errorf("State.RemoveEntryByTag() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.stdinText == "y\n" || tt.stdinText == "Y\n" {
				entries, _ := s.DB.GetEntriesByTag(t.Context(), tt.args.tag)
				if len(entries) != 0 {
					t.Errorf("expected 0 entries with tag %q, got %d", tt.args.tag, len(entries))
				}
			}
		})
	}
}