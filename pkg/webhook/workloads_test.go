package webhook

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func Test_workloadHasResourceProperties(t *testing.T) {
	type args struct {
		spec v1.PodSpec
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := workloadHasResourceProperties(tt.args.spec); got != tt.want {
				t.Errorf("workloadHasResourceProperties() = %v, want %v", got, tt.want)
			}
		})
	}
}
