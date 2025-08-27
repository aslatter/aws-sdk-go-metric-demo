package main

import (
	"testing"
)

func Test_fixName(t *testing.T) {

	tests := []struct {
		testName string
		name     string
		want     string
	}{
		{
			testName: "basic",
			name:     "hello_world",
			want:     "hello_world",
		},
		{
			testName: "adjacent invalid chars",
			name:     "hello, world",
			want:     "hello_world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := fixName(tt.name); got != tt.want {
				t.Errorf("fixName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixLabelName(t *testing.T) {
	tests := []struct {
		testName string
		name     string
		want     string
	}{
		{
			testName: "basic",
			name:     "label_key",
			want:     "label_key",
		}, {
			testName: "add key prefix",
			name:     "_label_key",
			want:     "key_label_key",
		},
		{
			testName: "replace invalid chars",
			name:     "label.key",
			want:     "label_key",
		},
		{
			testName: "retain reserved prefix",
			name:     ">>label key",
			want:     "__label_key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := fixLabelName(tt.name); got != tt.want {
				t.Errorf("fixLabelName() = %v, want %v", got, tt.want)
			}
		})
	}
}
