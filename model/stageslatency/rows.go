package stageslatency

import (
	"database/sql"

	"github.com/sjmudd/ps-top/log"
)

// Rows contains a slice of Rows
type Rows []Row

// select the rows into table
func collect(db *sql.DB) Rows {
	var t Rows

	log.Println("events_stages_summary_global_by_event_name.collect()")
	sql := "SELECT EVENT_NAME, COUNT_STAR, SUM_TIMER_WAIT FROM events_stages_summary_global_by_event_name WHERE SUM_TIMER_WAIT > 0"

	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var r Row
		if err := rows.Scan(
			&r.Name,
			&r.CountStar,
			&r.SumTimerWait); err != nil {
			log.Fatal(err)
		}

		// convert the stage name, removing any leading stage/sql/
		if len(r.Name) > 10 && r.Name[0:10] == "stage/sql/" {
			r.Name = r.Name[10:]
		}

		t = append(t, r)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	_ = rows.Close()

	log.Printf("recovered %v row(s):", len(t))
	log.Println(t)

	return t
}

// if the data in t2 is "newer", "has more values" than t then it needs refreshing.
// check this by comparing totals.
func (rows Rows) needsRefresh(otherRows Rows) bool {
	return totals(rows).SumTimerWait > totals(otherRows).SumTimerWait
}

// generate the totals of a table
func totals(rows Rows) Row {
	total := Row{Name: "Totals"}

	for _, row := range rows {
		total.SumTimerWait += row.SumTimerWait
		total.CountStar += row.CountStar
	}

	return total
}

// remove the initial values from those rows where there's a match
// - if we find a row we can't match ignore it
func (rows *Rows) subtract(initial Rows) {
	initialByName := make(map[string]int)

	// iterate over rows by name
	for i := range initial {
		initialByName[initial[i].Name] = i
	}

	for i := range *rows {
		name := (*rows)[i].Name
		if _, ok := initialByName[name]; ok {
			initialIndex := initialByName[name]
			(*rows)[i].subtract(initial[initialIndex])
		}
	}
}
