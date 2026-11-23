package jalali

import (
	"time"

	"github.com/jalaali/go-jalaali"
)

type Jalali struct {
	Year  int
	Month int
	Day   int
}

func TimeToJalali(times time.Time) Jalali {
	y, m, d, _ := jalaali.ToJalaali(times.Year(), times.Month(), times.Day())
	return Jalali{Year: y, Month: int(m), Day: d}
}
