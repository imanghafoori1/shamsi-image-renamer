package main

import (
	"strings"
	"testing"

	"github.com/jalaali/go-jalaali"
)

func TestSeasonNameLatin(t *testing.T) {
	cases := map[jalaali.Month]string{
		1: "spring", 2: "spring", 3: "spring",
		4: "summer", 5: "summer", 6: "summer",
		7: "autumn", 8: "autumn", 9: "autumn",
		10: "winter", 11: "winter", 12: "winter",
	}
	for m, want := range cases {
		if got := seasonNameLatin(m); got != want {
			t.Errorf("seasonNameLatin(%d) = %q, want %q", m, got, want)
		}
	}
	if seasonNameLatin(0) != "" || seasonNameLatin(13) != "" {
		t.Error("out-of-range month must return empty")
	}
}

func TestSeasonNamePersian(t *testing.T) {
	cases := map[jalaali.Month]string{
		1: "بهار", 7: "پاییز", 10: "زمستان", 12: "زمستان",
	}
	for m, want := range cases {
		if got := seasonNamePersian(m); got != want {
			t.Errorf("seasonNamePersian(%d) = %q, want %q", m, got, want)
		}
	}
}

func TestSeasonFormatSuffixes(t *testing.T) {
	// 1398-11-06 Jalali (2020-01-26 Gregorian) is in winter (month 11).
	jy, jm, jd := 1398, jalaali.Month(11), 6

	// Find the two season formats in dateFormats (they replace the old last one).
	var latinSeason, persianSeason *DateFormat
	for i := range dateFormats {
		if strings.Contains(dateFormats[i].Label, "winter") {
			latinSeason = &dateFormats[i]
		}
		if strings.Contains(dateFormats[i].Label, "زمستان") {
			persianSeason = &dateFormats[i]
		}
	}
	if latinSeason == nil || persianSeason == nil {
		t.Fatal("season date formats not found in dateFormats")
	}

	if got := latinSeason.Suffix(jy, jm, jd); got != "-1398-winter" {
		t.Errorf("latin season suffix = %q, want -1398-winter", got)
	}
	if got := persianSeason.Suffix(jy, jm, jd); got != "-زمستان-1398" {
		t.Errorf("persian season suffix = %q, want -زمستان-1398", got)
	}
}

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
	// 2020-01-26 == 1398-11-06 Jalali.
	jy, jm, jd, err := jalaali.ToJalaali(2020, 1, 26)
	if err != nil {
		t.Fatalf("ToJalaali: %v", err)
	}
	if jy != 1398 || jm != 11 || jd != 6 {
		t.Fatalf("sanity: expected 1398-11-06, got %d-%d-%d", jy, int(jm), jd)
	}

	// Each case identifies the format by a distinctive substring of its label,
	// so reordering or replacing formats doesn't silently break the test.
	tests := []struct {
		labelContains string
		want          string
	}{
		{"year-month-day", "-1398-11-06"},
		{"year-monthname-day", "-1398-bahman-6"},
		{"year-monthname)", "-1398-bahman"},
		{"year-season", "-1398-winter"},
		{"season-year", "-زمستان-1398"},
	}

	for _, tt := range tests {
		var found *DateFormat
		for i := range dateFormats {
			if strings.Contains(dateFormats[i].Label, tt.labelContains) {
				found = &dateFormats[i]
				break
			}
		}
		if found == nil {
			t.Errorf("no format with label containing %q", tt.labelContains)
			continue
		}
		if got := found.Suffix(jy, jm, jd); got != tt.want {
			t.Errorf("format %q: got %q, want %q", found.Label, got, tt.want)
		}
	}
}
