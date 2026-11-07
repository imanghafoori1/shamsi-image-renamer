package main

import (
	"fmt"

	"github.com/jalaali/go-jalaali"
)

// DateFormat describes one selectable suffix style.
type DateFormat struct {
	Label  string
	Suffix func(jy int, jm jalaali.Month, jd int) string
}

func seasonNameLatin(m jalaali.Month) string {
	switch {
	case m >= 1 && m <= 3:
		return "spring"
	case m >= 4 && m <= 6:
		return "summer"
	case m >= 7 && m <= 9:
		return "autumn"
	case m >= 10 && m <= 12:
		return "winter"
	}
	return ""
}

func seasonNamePersian(m jalaali.Month) string {
	switch {
	case m >= 1 && m <= 3:
		return "بهار"
	case m >= 4 && m <= 6:
		return "تابستان"
	case m >= 7 && m <= 9:
		return "پاییز"
	case m >= 10 && m <= 12:
		return "زمستان"
	}
	return ""
}

// Formats offered in the GUI. Index 0 is the default.
var dateFormats = []DateFormat{
	{
		Label: "1398-11-28  (year-month-day)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%04d-%02d-%02d", jy, int(jm), jd)
		},
	},
	{
		Label: "1398-bahman-28  (year-monthname-day)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%d-%s-%d", jy, monthNameLatin(jm), jd)
		},
	},
	{
		Label: "1398-bahman  (year-monthname)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%d-%s", jy, monthNameLatin(jm))
		},
	},
	{
		Label: "1398-winter  (year-season, Latin)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%d-%s", jy, seasonNameLatin(jm))
		},
	},
	{
		Label: "زمستان-1398  (season-year, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s-%d", seasonNamePersian(jm), jy)
		},
	},
	{
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s-%d", monthNamePersian(jm), jy)
		},
	},
}

func monthNameLatin(m jalaali.Month) string {
	months := []string{
		"farvardin", "ordibehesht", "khordad", "tir", "mordad", "shahrivar",
		"mehr", "aban", "azar", "dey", "bahman", "esfand",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}

func monthNamePersian(m jalaali.Month) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}
