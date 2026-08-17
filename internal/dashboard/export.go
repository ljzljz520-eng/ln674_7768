package dashboard

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"campus-device-hub/internal/domain"
)

type Export struct {
	GeneratedAt time.Time
	Summary     domain.DashboardSummary
	Cards       []VendorCard
	Rows        []DeviceRow
}

func (v *View) Export(now time.Time) (Export, error) {
	summary, err := v.Summary()
	if err != nil {
		return Export{}, err
	}
	cards, err := v.VendorCards()
	if err != nil {
		return Export{}, err
	}
	rows, err := v.Rows(now)
	if err != nil {
		return Export{}, err
	}
	return Export{GeneratedAt: now, Summary: summary, Cards: cards, Rows: rows}, nil
}

func (e Export) JSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}

func (e Export) CSV() string {
	lines := []string{"id,vendor,kind,label,status,updated_at"}
	rows := append([]DeviceRow(nil), e.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	for _, row := range rows {
		lines = append(lines, strings.Join([]string{row.ID, row.Vendor, string(row.Kind), quoteCSV(row.Label), row.Status, row.Updated.UTC().Format(time.RFC3339)}, ","))
	}
	return strings.Join(lines, "\n")
}

func quoteCSV(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	}
	return value
}

func (e Export) Text() string {
	return fmt.Sprintf("generated=%s\n%s\n%s", e.GeneratedAt.UTC().Format(time.RFC3339), RenderSummary(e.Summary), RenderCards(e.Cards))
}

func (v *View) ExportFiltered(now time.Time, text string, healthyOnly bool) (Export, error) {
	export, err := v.Export(now)
	if err != nil {
		return Export{}, err
	}
	export.Rows = FilterRows(export.Rows, text, healthyOnly)
	return export, nil
}
