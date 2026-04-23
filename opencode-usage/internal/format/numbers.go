package format

import (
	"fmt"
	"math"
	"time"
)

var KST = time.FixedZone("KST", 9*60*60)

func msToKST(ms int64) time.Time {
	return time.UnixMilli(ms).In(KST)
}

func KoreanNumber(n int64) string {
	if n == 0 {
		return "0"
	}
	abs := n
	if abs < 0 {
		abs = -abs
	}

	if abs >= 1_000_000_000_000 {
		jo := n / 1_000_000_000_000
		rem := n % 1_000_000_000_000
		if rem == 0 {
			return fmt.Sprintf("%d조", jo)
		}
		eok := rem / 100_000_000
		if eok > 0 {
			return fmt.Sprintf("%d억%d조", jo, eok)
		}
		return fmt.Sprintf("%d조", jo)
	}
	if abs >= 100_000_000 {
		eok := n / 100_000_000
		rem := n % 100_000_000
		if rem == 0 {
			return fmt.Sprintf("%d억", eok)
		}
		man := rem / 10_000
		if man > 0 {
			remMan := rem % 10_000
			if remMan > 0 {
				return fmt.Sprintf("%d억%d만%d", eok, man, remMan)
			}
			return fmt.Sprintf("%d억%d만", eok, man)
		}
		return fmt.Sprintf("%d억", eok)
	}
	if abs >= 10_000 {
		man := n / 10_000
		rem := n % 10_000
		if rem == 0 {
			return fmt.Sprintf("%d만", man)
		}
		return fmt.Sprintf("%d만%d", man, rem)
	}
	return fmt.Sprintf("%d", n)
}

func KoreanNumberShort(n int64) string {
	if n == 0 {
		return "0"
	}
	abs := n
	if abs < 0 {
		abs = -abs
	}

	if abs >= 1_000_000_000_000 {
		v := float64(n) / 1_000_000_000_000.0
		return fmt.Sprintf("%.1f조", v)
	}
	if abs >= 100_000_000 {
		v := float64(n) / 100_000_000.0
		return fmt.Sprintf("%.1f억", v)
	}
	if abs >= 10_000 {
		v := float64(n) / 10_000.0
		return fmt.Sprintf("%.1f만", v)
	}
	return fmt.Sprintf("%d", n)
}

func FormatCost(cost float64) string {
	if cost == 0 {
		return "무료"
	}
	abs := math.Abs(cost)
	if abs < 1 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

func FormatDate(dateStr string) string {
	if len(dateStr) >= 10 {
		return dateStr[5:7] + "/" + dateStr[8:10]
	}
	return dateStr
}

func FormatDateRange(fromMs, toMs int64) string {
	return fmt.Sprintf("%s ~ %s",
		msToKSTDate(fromMs),
		msToKSTDate(toMs),
	)
}

func msToKSTDate(ms int64) string {
	if ms <= 0 {
		return "시작"
	}
	t := msToKST(ms)
	return fmt.Sprintf("%d.%02d.%02d", t.Year(), t.Month(), t.Day())
}
