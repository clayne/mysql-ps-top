// Package rc provides routines to read ~/.pstoprc
// ps-top / ps-stats configuration
// - and to munge some table names based on the [munge] section (if present)
package rc

import (
	"log"
	"os"
	"regexp"
	"strings"

	go_ini "github.com/vaughan0/go-ini" // not sure what to do with dashes in names

	"github.com/sjmudd/ps-top/logger"
)

const (
	pstoprc = "~/.pstoprc" // location of the default pstop config file
)

// A single regexp expression from ~/.pstoprc
type mungeRegexp struct {
	pattern string
	replace string
	re      *regexp.Regexp
	valid   bool
}

var (
	regexps       []mungeRegexp
	loadedRegexps bool // Have we [attempted to] loaded data?
	haveRegexps   bool // Do we have any valid data?
)

// There must be a better way of doing this. Fix me...
// Copied from github.com/sjmudd/mysql_defaults_file so I should share this common code or fix it.
// Return the environment value of a given name.
func getEnviron(name string) string {
	for i := range os.Environ() {
		s := os.Environ()[i]
		keyValue := strings.Split(s, "=")

		if keyValue[0] == name {
			return keyValue[1]
		}
	}
	return ""
}

// Convert ~ to $HOME
// Copied from github.com/sjmudd/mysql_defaults_file so I should share this common code or fix it.
func convertFilename(filename string) string {
	for i := range filename {
		if filename[i] == '~' {
			//                      fmt.Println("Filename before", filename )
			filename = filename[:i] + getEnviron("HOME") + filename[i+1:]
			//                      fmt.Println("Filename after", filename )
			break
		}
	}

	return filename
}

// Load the ~/.pstoprc regexp expressions in section [munge]
func loadRegexps() {
	if loadedRegexps {
		return
	}
	loadedRegexps = true

	logger.Println("rc.loadRegexps()")

	haveRegexps = false
	filename := convertFilename(pstoprc)

	// Is the file is there?
	f, err := os.Open(filename)
	if err != nil {
		logger.Println("- unable to open " + filename + ", nothing to munge")
		return // can't open file. This is not fatal. We just can't do anything useful.
	}
	// If we get here the file is readable, so close it again.
	f.Close()

	// Load and process the ini file.
	i, err := go_ini.LoadFile(filename)
	if err != nil {
		log.Fatal("Could not load ~/.pstoprc", filename, ":", err)
	}

	// Note: This is wrong if I want to have an _ordered_ list of regexps
	// as go-ini provides me a hash so I lose the ordering. This may not
	// be desirable but as a first step accept this is broken.
	section := i.Section("munge")

	regexps = make([]mungeRegexp, 0, len(section))

	// now look for regexps and load them in...
	for k, v := range section {
		var m mungeRegexp
		var err error

		m.pattern, m.replace = k, v
		m.re, err = regexp.Compile(m.pattern)
		if err == nil {
			m.valid = true
		}
		regexps = append(regexps, m)
	}

	if len(regexps) > 0 {
		haveRegexps = true
	}
	logger.Println("- found", len(regexps), "regexps to use to munge output")
}

// Munge Optionally munges table names so they can be combined.
// - this reads ~/.pstoprc for configuration information.
// - e.g.
// [munge]
// <re_match> = <replace>
// _[0-9]{8}$ = _YYYYMMDD
// _[0-9]{6}$ = _YYYYMM
func Munge(name string) string {
	if !loadedRegexps {
		loadRegexps()
	}
	if !haveRegexps {
		return name // nothing to do so return what we were given.
	}

	munged := name

	for i := range regexps {
		if regexps[i].valid {
			if regexps[i].re.MatchString(munged) {
				munged = regexps[i].re.ReplaceAllLiteralString(munged, regexps[i].replace)
			}
		}
	}

	return munged
}
