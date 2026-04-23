package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/config"
	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/db"
	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/format"
	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/models"
)

func Execute() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "❌ 예상치 못한 오류: %v\n", r)
			os.Exit(1)
		}
	}()

	config.RootCmd.RunE = run
	return config.RootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	format.SetColorEnabled(config.IsColorEnabled())

	dbPath := config.GetDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ 데이터베이스 파일을 찾을 수 없습니다: %s\n", dbPath)
		return err
	}

	conn, err := db.Connect(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 데이터베이스 연결 실패: %v\n", err)
		return err
	}
	defer db.Close(conn)

	if err := checkTableExists(conn); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 데이터베이스에 message 테이블이 없습니다.\n")
		return err
	}

	startMs, endMs, displayName, err := config.ParsePeriod(config.Period, config.From, config.To)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	if config.JSON {
		return renderJSON(conn, startMs, endMs, displayName)
	}

	switch {
	case config.ByModel:
		return renderByModel(conn, startMs, endMs)
	case config.ByDay:
		return renderByDay(conn, startMs, endMs)
	case config.ByProject:
		return renderByProject(conn, startMs, endMs)
	case config.ByHour:
		return renderByHour(conn, startMs, endMs)
	case config.ByAgent:
		return renderByAgent(conn, startMs, endMs)
	default:
		return renderDefault(conn, startMs, endMs, displayName)
	}
}

func checkTableExists(conn *sql.DB) error {
	var name string
	err := conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='message'").Scan(&name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("message table not found")
	}
	return err
}

func renderJSON(conn *sql.DB, startMs, endMs int64, dateRange string) error {
	summary, err := db.GetSummary(conn, startMs, endMs)
	if err != nil {
		return err
	}
	userReqs, _ := db.GetUserRequestCount(conn, startMs, endMs)
	aborted, _ := db.GetAbortedCount(conn, startMs, endMs)
	summary.UserRequests = userReqs
	summary.AbortedCount = aborted

	models_, _ := db.GetModelUsage(conn, startMs, endMs)
	daily, _ := db.GetDailyUsage(conn, startMs, endMs)
	projects, _ := db.GetProjectUsage(conn, startMs, endMs)
	hourly, _ := db.GetHourlyUsage(conn, startMs, endMs)
	agents, _ := db.GetAgentUsage(conn, startMs, endMs)

	data := &models.UsageData{
		Summary:   summary,
		Models:    models_,
		Daily:     daily,
		Projects:  projects,
		Hourly:    hourly,
		Agents:    agents,
		DateRange: dateRange,
	}

	return format.RenderJSON(data)
}

func renderByModel(conn *sql.DB, startMs, endMs int64) error {
	models_, err := db.GetModelUsage(conn, startMs, endMs)
	if err != nil {
		return err
	}
	if len(models_) == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}
	fmt.Println(format.RenderModelTable(models_, 0))
	return nil
}

func renderByDay(conn *sql.DB, startMs, endMs int64) error {
	daily, err := db.GetDailyUsage(conn, startMs, endMs)
	if err != nil {
		return err
	}
	if len(daily) == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}
	fmt.Println(format.RenderDailyTable(daily, true, true))
	return nil
}

func renderByProject(conn *sql.DB, startMs, endMs int64) error {
	projects, err := db.GetProjectUsage(conn, startMs, endMs)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}
	fmt.Println(format.RenderProjectTable(projects, 0))
	return nil
}

func renderByHour(conn *sql.DB, startMs, endMs int64) error {
	hourly, err := db.GetHourlyUsage(conn, startMs, endMs)
	if err != nil {
		return err
	}
	if len(hourly) == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}
	fmt.Println(format.RenderHourlyTable(hourly))
	return nil
}

func renderByAgent(conn *sql.DB, startMs, endMs int64) error {
	agents, err := db.GetAgentUsage(conn, startMs, endMs)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}
	fmt.Println(format.RenderAgentTable(agents))
	return nil
}

func renderDefault(conn *sql.DB, startMs, endMs int64, displayName string) error {
	summary, err := db.GetSummary(conn, startMs, endMs)
	if err != nil {
		return err
	}

	if summary.TotalMessages == 0 {
		fmt.Println("해당 기간의 사용량이 없습니다.")
		return nil
	}

	userReqs, _ := db.GetUserRequestCount(conn, startMs, endMs)
	aborted, _ := db.GetAbortedCount(conn, startMs, endMs)
	summary.UserRequests = userReqs
	summary.AbortedCount = aborted

	dateRange := displayName
	if config.From != "" && config.To != "" {
		dateRange = format.FormatDateRange(startMs, endMs)
	}

	format.RenderSummary(summary, dateRange)

	models_, _ := db.GetModelUsage(conn, startMs, endMs)
	if len(models_) > 0 {
		format.RenderSectionHeader("🤖 모델 Top 5")
		fmt.Println(format.RenderModelTable(models_, 5))
	}

	projects, _ := db.GetProjectUsage(conn, startMs, endMs)
	if len(projects) > 0 {
		format.RenderSectionHeader("📁 프로젝트 Top 5")
		fmt.Println(format.RenderProjectTable(projects, 5))
	}

	daily, _ := db.GetDailyUsage(conn, startMs, endMs)
	if len(daily) > 0 {
		format.RenderSectionHeader("📅 일별 추이")
		fmt.Println(format.RenderDailyTable(daily, true, false))

		peakDate, peakMsg := format.FindPeakDaily(daily)
		if peakDate != "" {
			fmt.Printf("  %s %s (%s 응답)\n",
				format.Highlight("피크일:"),
				peakDate,
				format.KoreanNumberShort(peakMsg),
			)
		}
	}

	hourly, _ := db.GetHourlyUsage(conn, startMs, endMs)
	peakHour := format.FindPeakHour(hourly)
	if peakHour != "" {
		fmt.Printf("  %s %s\n", format.Highlight("피크 시간:"), peakHour)
	}

	return nil
}
