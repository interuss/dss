package models

import "testing"

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
