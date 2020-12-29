// Package context stores some common information used in various places
package context

import (
	"strings"
	"time"

	"github.com/sjmudd/ps-top/global"
	"github.com/sjmudd/ps-top/lib"
	"github.com/sjmudd/ps-top/model/filter"
	"github.com/sjmudd/ps-top/version"
)

// Context holds the common information
type Context struct {
	databaseFilter    *filter.DatabaseFilter
	last              time.Time
	status            *global.Status
	uptime            int
	variables         *global.Variables
	version           string
	wantRelativeStats bool
}

// NewContext returns the pointer to a new (empty) context
func NewContext(status *global.Status, variables *global.Variables, databaseFilter *filter.DatabaseFilter) *Context {
	return &Context{
		databaseFilter: databaseFilter,
		status:         status,
		variables:      variables,
	}
}

// DatabaseFilter returns the database filter to apply on queries (if appropriate)
func (c Context) DatabaseFilter() *filter.DatabaseFilter {
	return c.databaseFilter
}

// Hostname returns the current short hostname
func (c Context) Hostname() string {
	hostname := c.variables.Get("hostname")
	if index := strings.Index(hostname, "."); index >= 0 {
		hostname = hostname[0:index]
	}
	return hostname
}

// MySQLVersion returns the current MySQL version
func (c Context) MySQLVersion() string {
	return c.variables.Get("version")
}

// Version returns the Application version
func (c Context) Version() string {
	return version.Version
}

// ProgName returns the program's name
func (c Context) ProgName() string {
	return lib.ProgName
}

// Uptime returns the time that MySQL has been up
func (c Context) Uptime() int {
	return c.status.Get("Uptime")
}

// Variables returns a pointer to global.Variables
func (c Context) Variables() *global.Variables {
	return c.variables
}

// SetWantRelativeStats tells what we want to see
func (c *Context) SetWantRelativeStats(w bool) {
	c.wantRelativeStats = w
}

// WantRelativeStats tells us what we have asked for
func (c Context) WantRelativeStats() bool {
	return c.wantRelativeStats
}
