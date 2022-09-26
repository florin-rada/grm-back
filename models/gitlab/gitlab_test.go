package gitlab

import (
	"testing"

	gl "github.com/xanzy/go-gitlab"
)

func TestGetAllRunners(t *testing.T) {
	type args struct {
		client  *gl.Client
		page    int
		perPage int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		struct {
			name    string
			args    args
			wantErr bool
		}{"Testing valid", args{client, 1, 25}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetAllRunners(tt.args.client, tt.args.page, tt.args.perPage)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllRunners() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("GetAllRunners() error want valid list of runners")
				return
			}
			if got1 == nil {
				t.Errorf("GetAllRunners() error want valid response")
				return
			}
		})
	}
}

func TestGetRunnerDetails(t *testing.T) {
	type args struct {
		client   *gl.Client
		runnerID int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		struct {
			name    string
			args    args
			wantErr bool
		}{"Testing valid", args{client, 17709176}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRunnerDetails(tt.args.client, tt.args.runnerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRunnerDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("GetRunnerDetails() returned empty details")
			}
		})
	}
}
