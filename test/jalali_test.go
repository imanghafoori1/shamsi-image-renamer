package test

import (
	"testing"
	"time"

	"renamer/jalali"
)

func TestTimeToJalali(t *testing.T) {
	tm := time.Date(2020, 1, 26, 14, 30, 0, 0, time.UTC)
	j := jalali.TimeToJalali(tm)
	if j.Year != 1398 || j.Month != 11 || j.Day != 6 {
		t.Errorf("TimeToJalali = %+v, want 1398-11-06", j)
	}
}

func TestJalaliMonthNames(t *testing.T) {
	cases := []struct {
		month   int
		latin   string
		persian string
	}{
		{1, "farvardin", "فروردین"},
		{7, "mehr", "مهر"},
		{11, "bahman", "بهمن"},
		{12, "esfand", "اسفند"},
	}
	for _, c := range cases {
		j := jalali.Jalali{Month: c.month}
		if got := j.MonthNameLatin(); got != c.latin {
			t.Errorf("month %d latin = %q, want %q", c.month, got, c.latin)
		}
		if got := j.MonthNamePersian(); got != c.persian {
			t.Errorf("month %d persian = %q, want %q", c.month, got, c.persian)
		}
	}
	j := jalali.Jalali{Month: 0}
	if got := j.MonthNameLatin(); got != "" {
		t.Errorf("month 0 = %q, want empty", got)
	}
}

func TestJalaliSeasonNames(t *testing.T) {
	cases := []struct {
		month   int
		latin   string
		persian string
	}{
		{1, "spring", "بهار"},
		{4, "summer", "تابستان"},
		{7, "autumn", "پاییز"},
		{10, "winter", "زمستان"},
	}
	for _, c := range cases {
		j := jalali.Jalali{Month: c.month}
		if got := j.SeasonNameLatin(); got != c.latin {
			t.Errorf("month %d latin = %q, want %q", c.month, got, c.latin)
		}
		if got := j.SeasonNamePersian(); got != c.persian {
			t.Errorf("month %d persian = %q, want %q", c.month, got, c.persian)
		}
	}
}
