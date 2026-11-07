package main

import "testing"

func TestStripKnownDecorations(t *testing.T) {
	prefixes := []string{"NIKOND7200_", "NIKON D7200_"}
	suffixes := []string{
		"-1398-11-06", "-1398-bahman-6", "-1398-bahman", "-بهمن-1398", "-بهمن",
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"clean", "DSC_9340", "DSC_9340"},
		{"one suffix", "DSC_9340-1398-11-06", "DSC_9340"},
		{"overlapping suffixes, longest wins",
			"DSC_9340-1398-bahman-6", "DSC_9340"},
		{"two different suffixes",
			"DSC_9340-بهمن-1398-1398-bahman-6", "DSC_9340"},
		{"three suffixes, mixed",
			"DSC_9340-بهمن-1398-1398-bahman-6-1398-11-06", "DSC_9340"},
		{"double prefix", "NIKOND7200_NIKOND7200_DSC_9340", "DSC_9340"},
		{"raw-model prefix", "NIKON D7200_DSC_9340", "DSC_9340"},
		{"prefix + many suffixes",
			"NIKOND7200_NIKOND7200_DSC_9340-بهمن-1398-1398-bahman-6-1398-11-06",
			"DSC_9340"},
		{"unrelated name untouched", "HOLIDAY_photo", "HOLIDAY_photo"},
		{"middle occurrences untouched",
			"DSC-بهمن-1398_more", "DSC-بهمن-1398_more"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripKnownDecorations(tt.in, prefixes, suffixes); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripKnownDecorationsNilSets(t *testing.T) {
	if got := stripKnownDecorations("IMG_0003", nil, nil); got != "IMG_0003" {
		t.Errorf("got %q, want IMG_0003", got)
	}
}