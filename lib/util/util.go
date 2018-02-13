package util

import (
	"os"
	"strings"
	"time"
)

// StopOnError prints error to STDERR and exits with exit code 3
func StopOnError(err error) {
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(3)
	}
}

const timestampMacro = "{TIMESTAMP}"

// ExpandMacros expands the following macros in the filename:
// {TIMESTAMP} - current date and time in `YYYY-MM-DD-hh-mm-ss`` format
func ExpandMacros(filename string) string {
	if strings.Contains(filename, timestampMacro) {
		ts := time.Now().Format("2006-01-02-15-04-05")
		filename = strings.Replace(filename, timestampMacro, ts, -1)
	}
	return filename
}
