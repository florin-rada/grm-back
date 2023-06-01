package users

import "testing"

func TestCheckPasswordStrength(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Valid password",
			args:    args{password: "MyVal!dP@ssword"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Invalid password",
			args:    args{password: "MyPassworda"},
			want:    false,
			wantErr: true,
		},
		{
			name:    "Invalid password",
			args:    args{password: "My1Passworda"},
			want:    false,
			wantErr: true,
		},
		{
			name:    "Invalid password",
			args:    args{password: "ShortPa!s"},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPasswordStrength(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPasswordStrength() = %v, want %v", got, tt.want)
			}
		})
	}
}
