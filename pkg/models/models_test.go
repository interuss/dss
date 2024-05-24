package models

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestIDFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    ID
		wantErr bool
	}{
		{
			name: "compliant 1",
			args: args{
				s: "00000179-e36d-40be-838b-eca6ca350000",
			},
			want:    ID("00000179-e36d-40be-838b-eca6ca350000"),
			wantErr: false,
		},
		{
			name: "compliant 2",
			args: args{
				s: "03e5572a-f733-49af-bc14-8a18bd53ee39",
			},
			want:    ID("03e5572a-f733-49af-bc14-8a18bd53ee39"),
			wantErr: false,
		},
		{
			name: "wrong version",
			args: args{
				s: "00000179-e36d-50be-838b-eca6ca350000",
			},
			want:    ID(""),
			wantErr: true,
		},
		{
			name: "wrong variant",
			args: args{
				s: "00000179-e36d-d0be-d38b-eca6ca350000",
			},
			want:    ID(""),
			wantErr: true,
		},
		{
			name: "random",
			args: args{
				s: "sdh2o4mxc912asdvt34123dsfg2123dcsasefvaqyb4bp963r3",
			},
			want:    ID(""),
			wantErr: true,
		},
		{
			name: "missing leading char",
			args: args{
				s: "0000179-e36d-40be-838b-eca6ca350000",
			},
			want:    ID(""),
			wantErr: true,
		},
		{
			name: "missing last char",
			args: args{
				s: "00000179-e36d-40be-838b-eca6ca35000",
			},
			want:    ID(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("IDFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IDFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func scanIntoUUID(t *testing.T, id ID) *pgtype.UUID {
	uuid := pgtype.UUID{}
	err := uuid.Scan(id.String())
	assert.Nil(t, err)
	return &uuid
}

func TestID_PgUUID(t *testing.T) {
	someID := ID("00000179-e36d-40be-838b-eca6ca350000")
	badID := ID("00000179-e36d-40be-838b-eca6ca35")
	tests := []struct {
		name    string
		id      *ID
		want    *pgtype.UUID
		wantErr bool
	}{
		{
			name:    "Ok",
			id:      &someID,
			want:    scanIntoUUID(t, someID),
			wantErr: false,
		},
		{
			name:    "Nil ID",
			id:      nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Bad UUID",
			id:      &badID,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.PgUUID()
			if (err != nil) != tt.wantErr {
				t.Errorf("PgUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PgUUID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
