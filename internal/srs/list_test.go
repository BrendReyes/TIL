package srs

import (
	"testing"

	"github.com/brendreyes/til/internal/database"
)

func TestState_ListEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "empty table returns no error",
			fields:  fields{DB: newTestDB(t)},
			wantErr: false,
		},
		{
			name: "table with entries returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				seedEntry(t, q, "BFS uses a queue", "algorithms")
				return q
			}()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.ListEntry(); (err != nil) != tt.wantErr {
				t.Errorf("State.ListEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_GetSpecificEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		id int64
	}

	dbWithOne := newTestDB(t)
	existingID := seedEntry(t, dbWithOne, "Postgres EXPLAIN ANALYZE", "postgres")

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "existing id returns no error",
			fields:  fields{DB: dbWithOne},
			args:    args{id: existingID},
			wantErr: false,
		},
		{
			name:    "non-existent id returns no error",
			fields:  fields{DB: dbWithOne},
			args:    args{id: 99999},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.GetSpecificEntry(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("State.GetSpecificEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ListEntriesByTag(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		tag string
	}

	dbWithEntries := newTestDB(t)
	seedEntry(t, dbWithEntries, "defer runs LIFO", "go")
	seedEntry(t, dbWithEntries, "goroutines are cheap", "go")
	seedEntry(t, dbWithEntries, "BFS uses a queue", "algorithms")

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "tag with matching entries returns no error",
			fields:  fields{DB: dbWithEntries},
			args:    args{tag: "go"},
			wantErr: false,
		},
		{
			name:    "tag with no matches returns no error",
			fields:  fields{DB: dbWithEntries},
			args:    args{tag: "rust"},
			wantErr: false,
		},
		{
			name:    "empty tag returns no error",
			fields:  fields{DB: dbWithEntries},
			args:    args{tag: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.ListEntriesByTag(tt.args.tag); (err != nil) != tt.wantErr {
				t.Errorf("State.ListEntriesByTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_CountEntries(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "empty table returns no error",
			fields:  fields{DB: newTestDB(t)},
			wantErr: false,
		},
		{
			name: "table with entries returns no error",
			fields: fields{DB: func() *database.Queries {
				q := newTestDB(t)
				seedEntry(t, q, "defer runs LIFO", "go")
				seedEntry(t, q, "BFS uses a queue", "algorithms")
				return q
			}()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.CountEntries(); (err != nil) != tt.wantErr {
				t.Errorf("State.CountEntries() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}