package main

import (
	"testing"

	"github.com/jalaali/go-jalaali"
)

func TestMonthNameLatin(t *testing.T) {
	cases := map[jalaali.Month]string{
		1: "farvardin", 2: "ordibehesht", 3: "khordad", 4: "tir",
		5: "mordad", 6: "shahrivar", 7: "mehr", 8: "aban",
		9: "azar", 10: "dey", 11: "bahman", 12: "esfand",
	}
	for m, want := range cases {
		if got := monthNameLatin(m); got != want {
			t.Errorf("monthNameLatin(%d) = %q, want %q", m, got, want)
		}
	}
	if monthNameLatin(0) != "" || monthNameLatin(13) != "" {
		t.Error("out-of-range month must return empty")
	}
}

func TestMonthNamePersian(t *testing.T) {
	cases := map[jalaali.Month]string{
		1: "فروردین", 7: "مهر", 11: "بهمن", 12: "اسفند",
	}
	for m, want := range cases {
		if got := monthNamePersian(m); got != want {
			t.Errorf("monthNamePersian(%d) = %q, want %q", m, got, want)
		}
	}
	if monthNamePersian(0) != "" || monthNamePersian(13) != "" {
		t.Error("out-of-range month must return empty")
	}
}

func TestDateFormatSuffixes(t *testing.T) {
	// 2020-01-26 == 1398-11-06 Jalali; verify via go-jalaali first.
	jy, jm, jd, err := jalaali.ToJalaali(2020, 1, 26)
    if err != nil {
        t.Fatalf("ToJalaali: %v", err)
    }

	if jy != 1398 || jm != 11 || jd != 6 {
		t.Fatalf("sanity: expected 1398-11-06, got %d-%d-%d", jy, int(jm), jd)
	}

	tests := []struct {
		idx  int
		want string
	}{
		{0, "-1398-11-06"},
		{1, "-1398-bahman-6"},
		{2, "-1398-bahman"},
		{3, "-بهمن-1398"},
	}
	for _, tt := range tests {
		got := dateFormats[tt.idx].Suffix(jy, jm, jd)
		if got != tt.want {
			t.Errorf("dateFormats[%d] = %q, want %q", tt.idx, got, tt.want)
		}
	}
}