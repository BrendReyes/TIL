package srs

import (
	"testing"

	"github.com/brendreyes/til/internal/database"
)

func TestState_AddEntry(t *testing.T) {
	type fields struct {
		DB *database.Queries
	}
	type args struct {
		entry string
		tag   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "valid entry and tag",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "defer runs LIFO in Go", tag: "go"},
			wantErr: false,
		},
		{
			name:    "empty body returns error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "", tag: "go"},
			wantErr: true,
		},
		{
			name:    "empty tag returns error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "BFS uses a queue", tag: ""},
			wantErr: true,
		},
		{
			name:    "both empty returns error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "", tag: ""},
			wantErr: true,
		},
		{
			name:    "body exactly at 1000 chars is accepted",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: string(make([]byte, 1000)), tag: "go"},
			wantErr: false,
		},
		{
			name:    "body over 1000 chars returns error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: string(make([]byte, 1001)), tag: "go"},
			wantErr: true,
		},
		{
			name:    "tag exactly at 100 chars is accepted",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "some fact", tag: string(make([]byte, 100))},
			wantErr: false,
		},
		{
			name:    "tag over 100 chars returns error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{entry: "some fact", tag: string(make([]byte, 101))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.AddEntry(tt.args.entry, tt.args.tag); (err != nil) != tt.wantErr {
				t.Errorf("State.AddEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}