package format

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/models"
)

func RenderSummary(summary *models.Summary, dateRange string) {
	fmt.Println()
	fmt.Printf("  %s %s\n", Header(Emoji("📊", "[SUMMARY]")+" OpenCode 사용량 통계"), Highlight("["+dateRange+"]"))
	fmt.Println("  " + Separator())
	fmt.Println()

	fmt.Printf("  %-16s %s\n", "총 응답 수:", KoreanNumber(summary.TotalMessages)+"개")
	fmt.Printf("  %-16s %s\n", "사용자 요청:", KoreanNumber(summary.UserRequests)+"개")
	fmt.Printf("  %-16s %s\n", "Input 토큰:", KoreanNumber(summary.InputTokens))
	fmt.Printf("  %-16s %s\n", "Output 토큰:", KoreanNumber(summary.OutputTokens))
	fmt.Printf("  %-16s %s\n", "Reasoning:", KoreanNumber(summary.ReasoningTokens))
	fmt.Printf("  %-16s %s\n", "Cache Read:", KoreanNumber(summary.CacheReadTokens))
	fmt.Printf("  %-16s %s\n", "Cache Write:", KoreanNumber(summary.CacheWriteTokens))
	fmt.Printf("  %-16s %s\n", "총 비용:", FormatCost(summary.TotalCost))
	if summary.AbortedCount > 0 {
		fmt.Printf("  %-16s %s\n", "중단된 응답:", fmt.Sprintf("%d개", summary.AbortedCount))
	}
	fmt.Println()
}

func newTableWriter() table.Writer {
	t := table.NewWriter()
	if colorEnabled {
		t.SetStyle(table.StyleRounded)
	} else {
		t.SetStyle(table.StyleDefault)
		t.Style().Box = StyleASCII()
	}
	t.Style().Options.SeparateRows = false
	return t
}

func StyleASCII() table.BoxStyle {
	return table.BoxStyle{
		MiddleHorizontal: "-",
		MiddleVertical:   "|",
		MiddleSeparator:  "+",
		BottomLeft:       "+",
		BottomRight:      "+",
		BottomSeparator:  "+",
		Left:             "|",
		LeftSeparator:    "+",
		Right:            "|",
		RightSeparator:   "+",
		TopLeft:          "+",
		TopRight:         "+",
		TopSeparator:     "+",
		UnfinishedRow:    " ",
	}
}

func RenderModelTable(modelUsages []models.ModelUsage, limit int) string {
	if len(modelUsages) == 0 {
		return "  모델 사용량이 없습니다."
	}

	t := newTableWriter()

	t.AppendHeader(table.Row{"#", "모델", "Provider", "응답수", "Input", "Output", "비용"})

	data := modelUsages
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}

	for i, m := range data {
		t.AppendRow(table.Row{
			i + 1,
			m.Model,
			m.Provider,
			KoreanNumberShort(m.Messages),
			KoreanNumberShort(m.InputTokens),
			KoreanNumberShort(m.OutputTokens),
			FormatCost(m.Cost),
		})
	}

	if limit > 0 && len(modelUsages) > limit {
		t.AppendFooter(table.Row{"", fmt.Sprintf("... 외 %d개", len(modelUsages)-limit), "", "", "", "", ""})
	}

	return t.Render()
}

func RenderProjectTable(projects []models.ProjectUsage, limit int) string {
	if len(projects) == 0 {
		return "  프로젝트 사용량이 없습니다."
	}

	t := newTableWriter()

	t.AppendHeader(table.Row{"#", "프로젝트", "세션", "최초 사용", "마지막 사용"})

	data := projects
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}

	for i, p := range data {
		name := p.Project
		if len(name) > 20 {
			name = "..." + name[len(name)-17:]
		}
		t.AppendRow(table.Row{
			i + 1,
			name,
			p.Sessions,
			FormatDate(p.FirstUsed),
			FormatDate(p.LastUsed),
		})
	}

	if limit > 0 && len(projects) > limit {
		t.AppendFooter(table.Row{"", fmt.Sprintf("... 외 %d개", len(projects)-limit), "", "", ""})
	}

	return t.Render()
}

func RenderDailyTable(daily []models.DailyUsage, showPeak bool, showCumulative bool) string {
	if len(daily) == 0 {
		return "  일별 사용량이 없습니다."
	}

	var peakIdx int
	var peakMsg int64
	for i, d := range daily {
		if d.Messages > peakMsg {
			peakMsg = d.Messages
			peakIdx = i
		}
	}

	t := newTableWriter()

	if showCumulative {
		t.AppendHeader(table.Row{"날짜", "응답", "Input", "Output", "누적 응답", "누적 Input", "누적 Output"})
	} else {
		t.AppendHeader(table.Row{"날짜", "응답", "Input", "Output", "비용"})
	}

	for i, d := range daily {
		dateStr := FormatDate(d.Date)
		if showPeak && i == peakIdx {
			dateStr = Peak(dateStr + " ⬆")
		}

		if showCumulative {
			t.AppendRow(table.Row{
				dateStr,
				KoreanNumberShort(d.Messages),
				KoreanNumberShort(d.InputTokens),
				KoreanNumberShort(d.OutputTokens),
				KoreanNumberShort(d.CumMessages),
				KoreanNumberShort(d.CumInputTokens),
				KoreanNumberShort(d.CumOutputTokens),
			})
		} else {
			t.AppendRow(table.Row{
				dateStr,
				KoreanNumberShort(d.Messages),
				KoreanNumberShort(d.InputTokens),
				KoreanNumberShort(d.OutputTokens),
				FormatCost(d.Cost),
			})
		}
	}

	return t.Render()
}

func RenderHourlyTable(hourly []models.HourlyUsage) string {
	if len(hourly) == 0 {
		return "  시간별 사용량이 없습니다."
	}

	t := newTableWriter()

	t.AppendHeader(table.Row{"시간", "응답", "Input", "Output"})

	for _, h := range hourly {
		t.AppendRow(table.Row{
			h.Hour,
			KoreanNumberShort(h.Messages),
			KoreanNumberShort(h.InputTokens),
			KoreanNumberShort(h.OutputTokens),
		})
	}

	return t.Render()
}

func RenderAgentTable(agents []models.AgentUsage) string {
	if len(agents) == 0 {
		return "  에이전트 사용량이 없습니다."
	}

	t := newTableWriter()

	t.AppendHeader(table.Row{"#", "에이전트", "응답", "Input", "Output"})

	for i, a := range agents {
		name := a.Agent
		if name == "" {
			name = "(unknown)"
		}
		t.AppendRow(table.Row{
			i + 1,
			name,
			KoreanNumberShort(a.Messages),
			KoreanNumberShort(a.InputTokens),
			KoreanNumberShort(a.OutputTokens),
		})
	}

	return t.Render()
}

func RenderSectionHeader(title string) {
	fmt.Printf("  %s\n", Header(title))
}

var SectionTitles = struct {
	Model   string
	Project string
	Daily   string
}{
	Model:   "[TOP 5] 모델별 사용량",
	Project: "[TOP 5] 프로젝트별 세션 수",
	Daily:   "일별 사용 추이",
}

func FindPeakHour(hourly []models.HourlyUsage) string {
	if len(hourly) == 0 {
		return ""
	}
	var peakHour string
	var peakMsg int64
	for _, h := range hourly {
		if h.Messages > peakMsg {
			peakMsg = h.Messages
			peakHour = h.Hour
		}
	}
	return peakHour
}

func FindPeakDaily(daily []models.DailyUsage) (string, int64) {
	if len(daily) == 0 {
		return "", 0
	}
	var peakDate string
	var peakMsg int64
	for _, d := range daily {
		if d.Messages > peakMsg {
			peakMsg = d.Messages
			peakDate = d.Date
		}
	}
	return peakDate, peakMsg
}

func Truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return "..." + s[len(s)-maxLen+3:]
	}
	return s[:maxLen]
}
