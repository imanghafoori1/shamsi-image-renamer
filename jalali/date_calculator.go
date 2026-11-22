package jalali

func (j *Jalali) SeasonNameLatin() string {
	m := j.Month
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

func (j *Jalali) SeasonNamePersian() string {
	switch j.Month {
	case 1, 2, 3:
		return "بهار"
	case 4, 5, 6:
		return "تابستان"
	case 7, 8, 9:
		return "پاییز"
	case 10, 11, 12:
		return "زمستان"
	}
	return ""
}

func (j *Jalali) MonthNameLatin() string {
	m := j.Month

	months := []string{
		"farvardin", "ordibehesht", "khordad", "tir", "mordad", "shahrivar",
		"mehr", "aban", "azar", "dey", "bahman", "esfand",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}

func (j *Jalali) MonthNamePersian() string {
	m := j.Month
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if m >= 1 && m <= 12 {
		return months[int(m)-1]
	}
	return ""
}
