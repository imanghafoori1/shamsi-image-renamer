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
		Label: "بهمن-1398  (monthname-year, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s-%d", monthNamePersian(jm), jy)
		},
	},
	{
		Label: "بهمن  (monthname only, Persian)",
		Suffix: func(jy int, jm jalaali.Month, jd int) string {
			return fmt.Sprintf("-%s", monthNamePersian(jm))
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