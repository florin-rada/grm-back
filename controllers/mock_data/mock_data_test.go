package mock_data

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReturnMockData(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReturnMockData(tt.args.ctx)
		})
	}
}
