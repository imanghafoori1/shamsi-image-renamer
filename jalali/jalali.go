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
	jy, jm, jd, _ := jalaali.ToJalaali(times.Year(), times.Month(), times.Day())
	return Jalali{jy, int(jm), jd}
}
