package srs

import (
	"testing"

	"github.com/brendreyes/til/internal/database"
)

func TestState_EditEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		id int64
	}

	tests := []struct {
		name        string
		fields      fields
		args        func(t *testing.T, q *database.Queries) int64
		mockEditor  func(b, t string) (string, string, bool, error)
		wantErr     bool
		wantBody    string
		wantTag     string
	}{
		{
			name:   "non-existent id returns no error",
			fields: fields{DB: newTestDB(t)},
			args: func(t *testing.T, q *database.Queries) int64 {
				return 99999
			},
			wantErr: false,
		},
		{
			name:   "successful edit",
			fields: fields{DB: newTestDB(t)},
			args: func(t *testing.T, q *database.Queries) int64 {
				return seedEntry(t, q, "original body", "original tag")
			},
			mockEditor: func(b, tag string) (string, string, bool, error) {
				return "updated body", "updated tag", true, nil
			},
			wantErr:  false,
			wantBody: "updated body",
			wantTag:  "updated tag",
		},
		{
			name:   "aborted edit returns no error",
			fields: fields{DB: newTestDB(t)},
			args: func(t *testing.T, q *database.Queries) int64 {
				return seedEntry(t, q, "original body", "original tag")
			},
			mockEditor: func(b, tag string) (string, string, bool, error) {
				return "ignored", "ignored", false, nil
			},
			wantErr:  false,
			wantBody: "original body",
			wantTag:  "original tag",
		},
		{
			name:   "no changes detected returns no error",
			fields: fields{DB: newTestDB(t)},
			args: func(t *testing.T, q *database.Queries) int64 {
				return seedEntry(t, q, "original body", "original tag")
			},
			mockEditor: func(b, tag string) (string, string, bool, error) {
				return "original body", "original tag", true, nil
			},
			wantErr:  false,
			wantBody: "original body",
			wantTag:  "original tag",
		},
		{
			name:   "empty body returns error",
			fields: fields{DB: newTestDB(t)},
			args: func(t *testing.T, q *database.Queries) int64 {
				return seedEntry(t, q, "original body", "original tag")
			},
			mockEditor: func(b, tag string) (string, string, bool, error) {
				return "", "some tag", true, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockEditor != nil {
				orig := RunEditorFunc
				RunEditorFunc = tt.mockEditor
				defer func() { RunEditorFunc = orig }()
			}

			id := tt.args(t, tt.fields.DB)
			s := &State{
				DB: tt.fields.DB,
			}
			err := s.EditEntry(id)
			if (err != nil) != tt.wantErr {
				t.Errorf("State.EditEntry() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.wantBody != "" {
				entry, _ := s.DB.GetEntryByID(t.Context(), id)
				if entry.Body != tt.wantBody {
					t.Errorf("expected body %q, got %q", tt.wantBody, entry.Body)
				}
				if entry.Tag != tt.wantTag {
					t.Errorf("expected tag %q, got %q", tt.wantTag, entry.Tag)
				}
			}
		})
	}
}