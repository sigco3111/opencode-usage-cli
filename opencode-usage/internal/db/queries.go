package db

import (
	"database/sql"
	"fmt"

	"github.com/sigco3111/opencode-usage-cli/opencode-usage/internal/models"
)

func GetSummary(db *sql.DB, startMs, endMs int64) (*models.Summary, error) {
	query := `
		SELECT
			COUNT(*) as total_messages,
			COALESCE(SUM(json_extract(data, '$.tokens.input')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.output')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.reasoning')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.cache.read')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.cache.write')), 0),
			COALESCE(SUM(json_extract(data, '$.cost')), 0)
		FROM message
		WHERE time_created >= ? AND time_created <= ?
			AND json_extract(data, '$.role') = 'assistant'
			AND json_extract(data, '$.error') IS NULL
	`
	var s models.Summary
	err := db.QueryRow(query, startMs, endMs).Scan(
		&s.TotalMessages, &s.InputTokens, &s.OutputTokens,
		&s.ReasoningTokens, &s.CacheReadTokens, &s.CacheWriteTokens, &s.TotalCost,
	)
	if err != nil {
		return nil, fmt.Errorf("summary query failed: %w", err)
	}
	return &s, nil
}

func GetUserRequestCount(db *sql.DB, startMs, endMs int64) (int64, error) {
	var count int64
	err := db.QueryRow(
		"SELECT COUNT(*) FROM message WHERE time_created >= ? AND time_created <= ? AND json_extract(data, '$.role') = 'user'",
		startMs, endMs,
	).Scan(&count)
	return count, err
}

func GetAbortedCount(db *sql.DB, startMs, endMs int64) (int64, error) {
	var count int64
	err := db.QueryRow(
		"SELECT COUNT(*) FROM message WHERE time_created >= ? AND time_created <= ? AND json_extract(data, '$.role') = 'assistant' AND json_extract(data, '$.error') IS NOT NULL",
		startMs, endMs,
	).Scan(&count)
	return count, err
}

func GetModelUsage(db *sql.DB, startMs, endMs int64) ([]models.ModelUsage, error) {
	query := `
		SELECT
			COALESCE(json_extract(data, '$.modelID'), 'unknown'),
			COALESCE(json_extract(data, '$.providerID'), 'unknown'),
			COUNT(*),
			COALESCE(SUM(json_extract(data, '$.tokens.input')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.output')), 0),
			COALESCE(SUM(json_extract(data, '$.cost')), 0)
		FROM message
		WHERE time_created >= ? AND time_created <= ?
			AND json_extract(data, '$.role') = 'assistant'
			AND json_extract(data, '$.error') IS NULL
		GROUP BY json_extract(data, '$.modelID'), json_extract(data, '$.providerID')
		ORDER BY COUNT(*) DESC
	`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ModelUsage
	for rows.Next() {
		var m models.ModelUsage
		if err := rows.Scan(&m.Model, &m.Provider, &m.Messages, &m.InputTokens, &m.OutputTokens, &m.Cost); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

func GetDailyUsage(db *sql.DB, startMs, endMs int64) ([]models.DailyUsage, error) {
	query := `
		SELECT
			date(time_created / 1000, 'unixepoch', '+9 hours'),
			COUNT(*),
			COALESCE(SUM(json_extract(data, '$.tokens.input')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.output')), 0),
			COALESCE(SUM(json_extract(data, '$.cost')), 0)
		FROM message
		WHERE time_created >= ? AND time_created <= ?
			AND json_extract(data, '$.role') = 'assistant'
			AND json_extract(data, '$.error') IS NULL
		GROUP BY date(time_created / 1000, 'unixepoch', '+9 hours')
		ORDER BY date(time_created / 1000, 'unixepoch', '+9 hours')
	`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.DailyUsage
	var cumMsg, cumIn, cumOut int64
	for rows.Next() {
		var d models.DailyUsage
		if err := rows.Scan(&d.Date, &d.Messages, &d.InputTokens, &d.OutputTokens, &d.Cost); err != nil {
			return nil, err
		}
		cumMsg += d.Messages
		cumIn += d.InputTokens
		cumOut += d.OutputTokens
		d.CumMessages = cumMsg
		d.CumInputTokens = cumIn
		d.CumOutputTokens = cumOut
		result = append(result, d)
	}
	return result, nil
}

func GetProjectUsage(db *sql.DB, startMs, endMs int64) ([]models.ProjectUsage, error) {
	query := `
		SELECT
			CASE
				WHEN INSTR(s.directory, '/') > 0 THEN
					SUBSTR(s.directory, LENGTH(RTRIM(s.directory, REPLACE(s.directory, '/', ''))) + 1)
				ELSE s.directory
			END as project,
			COUNT(DISTINCT m.session_id),
			MIN(date(m.time_created / 1000, 'unixepoch', '+9 hours')),
			MAX(date(m.time_created / 1000, 'unixepoch', '+9 hours'))
		FROM message m
		JOIN session s ON m.session_id = s.id
		WHERE m.time_created >= ? AND m.time_created <= ?
			AND json_extract(m.data, '$.role') = 'assistant'
			AND json_extract(m.data, '$.error') IS NULL
		GROUP BY project
		ORDER BY COUNT(DISTINCT m.session_id) DESC
	`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ProjectUsage
	for rows.Next() {
		var p models.ProjectUsage
		if err := rows.Scan(&p.Project, &p.Sessions, &p.FirstUsed, &p.LastUsed); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func GetHourlyUsage(db *sql.DB, startMs, endMs int64) ([]models.HourlyUsage, error) {
	query := `
		SELECT
			strftime('%H:00', time_created / 1000, 'unixepoch', '+9 hours'),
			COUNT(*),
			COALESCE(SUM(json_extract(data, '$.tokens.input')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.output')), 0)
		FROM message
		WHERE time_created >= ? AND time_created <= ?
			AND json_extract(data, '$.role') = 'assistant'
			AND json_extract(data, '$.error') IS NULL
		GROUP BY strftime('%H:00', time_created / 1000, 'unixepoch', '+9 hours')
		ORDER BY COUNT(*) DESC
	`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.HourlyUsage
	for rows.Next() {
		var h models.HourlyUsage
		if err := rows.Scan(&h.Hour, &h.Messages, &h.InputTokens, &h.OutputTokens); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, nil
}

func GetAgentUsage(db *sql.DB, startMs, endMs int64) ([]models.AgentUsage, error) {
	query := `
		SELECT
			REPLACE(REPLACE(TRIM(json_extract(data, '$.agent')), char(0x200B), ''), char(0xFEFF), ''),
			COUNT(*),
			COALESCE(SUM(json_extract(data, '$.tokens.input')), 0),
			COALESCE(SUM(json_extract(data, '$.tokens.output')), 0)
		FROM message
		WHERE time_created >= ? AND time_created <= ?
			AND json_extract(data, '$.role') = 'assistant'
			AND json_extract(data, '$.error') IS NULL
		GROUP BY REPLACE(REPLACE(TRIM(json_extract(data, '$.agent')), char(0x200B), ''), char(0xFEFF), '')
		ORDER BY COUNT(*) DESC
	`
	rows, err := db.Query(query, startMs, endMs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.AgentUsage
	for rows.Next() {
		var a models.AgentUsage
		if err := rows.Scan(&a.Agent, &a.Messages, &a.InputTokens, &a.OutputTokens); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}
