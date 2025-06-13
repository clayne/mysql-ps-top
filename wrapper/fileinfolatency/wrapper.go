// Package fileinfolatency holds the routines which manage the file_summary_by_instance table.
package fileinfolatency

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/sjmudd/ps-top/config"
	"github.com/sjmudd/ps-top/model/fileinfo"
	"github.com/sjmudd/ps-top/utils"
)

// Wrapper wraps a FileIoLatency struct representing the contents of the data collected from file_summary_by_instance, but adding formatting for presentation in the terminal
type Wrapper struct {
	fiol *fileinfo.FileIoLatency
}

// NewFileSummaryByInstance creates a wrapper around FileIoLatency
func NewFileSummaryByInstance(cfg *config.Config, db *sql.DB) *Wrapper {
	return &Wrapper{
		fiol: fileinfo.NewFileSummaryByInstance(cfg, db),
	}
}

// ResetStatistics resets the statistics to last values
func (fiolw *Wrapper) ResetStatistics() {
	fiolw.fiol.ResetStatistics()
}

// Collect data from the db, then merge it in.
func (fiolw *Wrapper) Collect() {
	fiolw.fiol.Collect()
	sort.Sort(byLatency(fiolw.fiol.Results))
}

// Headings returns the headings for a table
func (fiolw Wrapper) Headings() string {
	return fmt.Sprintf("%10s %6s|%6s %6s %6s|%8s %8s|%8s %6s %6s %6s|%s",
		"Latency",
		"%",
		"Read",
		"Write",
		"Misc",
		"Rd bytes",
		"Wr bytes",
		"Ops",
		"R Ops",
		"W Ops",
		"M Ops",
		"Table Name")
}

// RowContent returns the rows we need for displaying
func (fiolw Wrapper) RowContent() []string {
	rows := make([]string, 0, len(fiolw.fiol.Results))

	for i := range fiolw.fiol.Results {
		rows = append(rows, fiolw.content(fiolw.fiol.Results[i], fiolw.fiol.Totals))
	}

	return rows
}

// TotalRowContent returns all the totals
func (fiolw Wrapper) TotalRowContent() string {
	return fiolw.content(fiolw.fiol.Totals, fiolw.fiol.Totals)
}

// EmptyRowContent returns an empty string of data (for filling in)
func (fiolw Wrapper) EmptyRowContent() string {
	var empty fileinfo.Row

	return fiolw.content(empty, empty)
}

// Description returns a description of the table
func (fiolw Wrapper) Description() string {
	var count int

	for row := range fiolw.fiol.Results {
		if fiolw.fiol.Results[row].HasData() {
			count++
		}
	}

	return fmt.Sprintf("File I/O Latency (file_summary_by_instance) %d rows", count)
}

// HaveRelativeStats is true for this object
func (fiolw Wrapper) HaveRelativeStats() bool {
	return fiolw.fiol.HaveRelativeStats()
}

// FirstCollectTime returns the time the first value was collected
func (fiolw Wrapper) FirstCollectTime() time.Time {
	return fiolw.fiol.FirstCollected
}

// LastCollectTime returns the time the last value was collected
func (fiolw Wrapper) LastCollectTime() time.Time {
	return fiolw.fiol.LastCollected
}

// WantRelativeStats indiates if we want relative statistics
func (fiolw Wrapper) WantRelativeStats() bool {
	return fiolw.fiol.WantRelativeStats()
}

// content generate a printable result for a row, given the totals
func (fiolw Wrapper) content(row, totals fileinfo.Row) string {
	var name = row.Name

	// We assume that if CountStar = 0 then there's no data at all...
	// when we have no data we really don't want to show the name either.
	if (row.SumTimerWait == 0 && row.CountStar == 0 && row.SumNumberOfBytesRead == 0 && row.SumNumberOfBytesWrite == 0) && name != "Totals" {
		name = ""
	}

	return fmt.Sprintf("%10s %6s|%6s %6s %6s|%8s %8s|%8s %6s %6s %6s|%s",
		utils.FormatTime(row.SumTimerWait),
		utils.FormatPct(utils.Divide(row.SumTimerWait, totals.SumTimerWait)),
		utils.FormatPct(utils.Divide(row.SumTimerRead, row.SumTimerWait)),
		utils.FormatPct(utils.Divide(row.SumTimerWrite, row.SumTimerWait)),
		utils.FormatPct(utils.Divide(row.SumTimerMisc, row.SumTimerWait)),
		utils.FormatAmount(row.SumNumberOfBytesRead),
		utils.FormatAmount(row.SumNumberOfBytesWrite),
		utils.FormatAmount(row.CountStar),
		utils.FormatPct(utils.Divide(row.CountRead, row.CountStar)),
		utils.FormatPct(utils.Divide(row.CountWrite, row.CountStar)),
		utils.FormatPct(utils.Divide(row.CountMisc, row.CountStar)),
		name)
}

type byLatency fileinfo.Rows

func (rows byLatency) Len() int      { return len(rows) }
func (rows byLatency) Swap(i, j int) { rows[i], rows[j] = rows[j], rows[i] }
func (rows byLatency) Less(i, j int) bool {
	return (rows[i].SumTimerWait > rows[j].SumTimerWait) ||
		((rows[i].SumTimerWait == rows[j].SumTimerWait) && (rows[i].Name < rows[j].Name))
}
