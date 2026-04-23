package format

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/models"
)

func RenderJSON(data *models.UsageData) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %w", err)
	}
	return nil
}
