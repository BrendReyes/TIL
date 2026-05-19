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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "non-existent id returns no error",
			fields:  fields{DB: newTestDB(t)},
			args:    args{id: 99999},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				DB: tt.fields.DB,
			}
			if err := s.EditEntry(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("State.EditEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}