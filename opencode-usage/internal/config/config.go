package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/models"
	"github.com/spf13/cobra"
)

var (
	Period   string
	From     string
	To       string
	ByModel  bool
	ByDay    bool
	ByProject bool
	ByHour   bool
	ByAgent  bool
	JSON     bool
	DBPath   string
	Color    string
)

var KST = time.FixedZone("KST", 9*60*60)

var RootCmd = &cobra.Command{
	Use:     "oc-usage",
	Short:   "OpenCode usage statistics CLI",
	Version: models.Version,
}

func init() {
	RootCmd.Flags().StringVarP(&Period, "period", "p", "month", "Period: today|week|month|all")
	RootCmd.Flags().StringVar(&From, "from", "", "Start date (YYYY-MM-DD)")
	RootCmd.Flags().StringVar(&To, "to", "", "End date (YYYY-MM-DD)")
	RootCmd.Flags().BoolVar(&ByModel, "by-model", false, "Show model breakdown")
	RootCmd.Flags().BoolVar(&ByDay, "by-day", false, "Show daily trend")
	RootCmd.Flags().BoolVar(&ByProject, "by-project", false, "Show project breakdown")
	RootCmd.Flags().BoolVar(&ByHour, "by-hour", false, "Show hourly distribution")
	RootCmd.Flags().BoolVar(&ByAgent, "by-agent", false, "Show agent breakdown")
	RootCmd.Flags().BoolVarP(&JSON, "json", "j", false, "Output as JSON")
	RootCmd.Flags().StringVar(&DBPath, "db-path", "", "Database path")
	RootCmd.Flags().StringVar(&Color, "color", "auto", "Color: auto|always|never")
	RootCmd.SetVersionTemplate("oc-usage v" + models.Version + "\n")
}

func ParsePeriod(period, from, to string) (int64, int64, string, error) {
	now := time.Now().In(KST)

	if from != "" || to != "" {
		if from == "" || to == "" {
			return 0, 0, "", fmt.Errorf("--from과 --to는 함께 사용해야 합니다")
		}
		startDate, err := time.ParseInLocation("2006-01-02", from, KST)
		if err != nil {
			return 0, 0, "", fmt.Errorf("❌ 잘못된 날짜 형식: YYYY-MM-DD 형식을 사용하세요")
		}
		endDate, err := time.ParseInLocation("2006-01-02", to, KST)
		if err != nil {
			return 0, 0, "", fmt.Errorf("❌ 잘못된 날짜 형식: YYYY-MM-DD 형식을 사용하세요")
		}
		startMs := startDate.UnixMilli()
		endMs := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond).UnixMilli()
		displayName := fmt.Sprintf("%s ~ %s", from, to)
		return startMs, endMs, displayName, nil
	}

	switch period {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, KST).UnixMilli()
		end := now.UnixMilli()
		return start, end, "오늘", nil
	case "week":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, KST).AddDate(0, 0, -7).UnixMilli()
		end := now.UnixMilli()
		return start, end, "최근 7일", nil
	case "month":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, KST).AddDate(0, 0, -30).UnixMilli()
		end := now.UnixMilli()
		return start, end, "최근 30일", nil
	case "all":
		return 0, now.UnixMilli(), "전체 기간", nil
	default:
		return 0, 0, "", fmt.Errorf("❌ 알 수 없는 기간: %s (today|week|month|all)", period)
	}
}

func GetDBPath() string {
	path := DBPath
	if path == "" {
		path = models.DefaultDBPath
	}
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			home := os.Getenv("HOME")
			path = filepath.Join(home, path[2:])
		} else {
			path = filepath.Join(usr.HomeDir, path[2:])
		}
	}
	return path
}

func IsColorEnabled() bool {
	switch Color {
	case "always":
		return true
	case "never":
		return false
	default:
		if os.Getenv("TERM") == "dumb" {
			return false
		}
		fi, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
}
