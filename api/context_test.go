package api

import (
	"net/http"
	"testing"
)

func Test_status(t *testing.T) {
	type args struct {
		i []int
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"given code", args{[]int{404}}, 404},
		{"not given code", args{[]int{}}, http.StatusInternalServerError},
		{"given nil", args{nil}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotStatus := status(tt.args.i); gotStatus != tt.wantStatus {
				t.Errorf("status() = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
