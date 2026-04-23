package format

import (
	"github.com/fatih/color"
)

var colorEnabled = true

func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
	color.NoColor = !enabled
}

func Header(text string) string {
	if !colorEnabled {
		return text
	}
	return color.New(color.FgBlue, color.Bold).Sprint(text)
}

func Peak(text string) string {
	if !colorEnabled {
		return text
	}
	return color.New(color.FgRed, color.Bold).Sprint(text)
}

func Highlight(text string) string {
	if !colorEnabled {
		return text
	}
	return color.New(color.FgYellow, color.Bold).Sprint(text)
}

func Separator() string {
	return "────────────────────────────────────────"
}
