package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// parses logs from laravel in this format:
// [2024-10-18 21:27:05] local.DEBUG: select * from "files" where "id" = ? limit 1 | 00b62c03-3827-4f49-a731-ed2494ede688 | 0.7900ms | app/Models/File.php:184 | Illuminate\Database\Eloquent\Model::fresh

// LogEntry represents a parsed log entry from the Laravel log file.
type LogEntry struct {
	Timestamp          time.Time
	LogLevel           string
	SQLQuery           string
	QueryBindings      string
	QueryExecTime      string
	SourceCodeLocation string
	CallerSignature    string
}

// Precompiled regular expression to parse the log line.
var logLineRegex = regexp.MustCompile(`\[(.*?)\]\s*(\w+\.\w+):\s*(.*)`)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parser <logfile>")
		return
	}
	logfile := os.Args[1]

	file, err := os.Open(logfile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	var entries []LogEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseLine(line)
		if err != nil {
			//fmt.Printf("Error parsing line: %v\n", err)
			continue
		}
		entries = append(entries, *entry)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	groupBySourceCodeLocation(entries)
	// displayQueryOptimizationSuggestions(entries)
}

// parseLine parses a single line from the log file into a LogEntry.
func parseLine(line string) (*LogEntry, error) {
	matches := logLineRegex.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil, fmt.Errorf("line does not match expected format")
	}
	timestampStr := matches[1]
	logLevel := matches[2]
	message := matches[3]

	// Parse the timestamp
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %v", err)
	}

	// Split the message into fields
	fields := strings.SplitN(message, " | ", 5)
	if len(fields) != 5 {
		return nil, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}

	entry := &LogEntry{
		Timestamp:          timestamp,
		LogLevel:           logLevel,
		SQLQuery:           strings.TrimSpace(fields[0]),
		QueryBindings:      strings.TrimSpace(fields[1]),
		QueryExecTime:      strings.TrimSpace(fields[2]),
		SourceCodeLocation: strings.TrimSpace(fields[3]),
		CallerSignature:    strings.TrimSpace(fields[4]),
	}

	return entry, nil
}

// groupBySourceCodeLocation groups LogEntry items by SourceCodeLocation and prints the count for each.
func groupBySourceCodeLocation(entries []LogEntry) {
	counts := make(map[string]int)
	times := make(map[string]float64)

	for _, entry := range entries {
		counts[entry.SourceCodeLocation]++
		time, _ := time.ParseDuration(entry.QueryExecTime)
		times[entry.SourceCodeLocation] += float64(time.Abs().Milliseconds())
	}

	// sort by location name (alpha sort):
	sortedLocations := make([]string, 0, len(counts))

	for location, _ := range counts {
		sortedLocations = append(sortedLocations, location)
	}

	sort.Strings(sortedLocations)

	fmt.Println("SourceCodeLocation counts:")
	// for location, count := range counts {
	for _, location := range sortedLocations {
		count := counts[location]
		ms := times[location]
		average := ms / float64(count)

		if (count >= 750 || average >= 30.0) && !strings.HasPrefix(location, "unknown:") {
			fmt.Printf("    %s (count: %d, mean time: %0.4f ms, total time: %0.0f ms)\n", location, count, average, ms)
		}
	}
}

// func displayQueryOptimizationSuggestions(entries []LogEntry) {
// 	queries, err := pgstat.ConnectAndFetchPgStatStatements(pgstat.BuildPostgresDsn("localhost", 5432, "root", "password", "database"))
// 	if err != nil {
// 		fmt.Printf("Error fetching pg_stat_statements: %v\n", err)
// 		return
// 	}

// 	suggestions := pgstat.AnalyzeQueries(queries)

// 	fmt.Println("\nOptimization suggestions:")
// 	for _, suggestion := range suggestions {
// 		fmt.Printf("    QueryID: %d\n", suggestion.QueryID)
// 		fmt.Printf("    Query: %s\n", suggestion.Query)
// 		fmt.Printf("    Suggestions: %v\n", suggestion.Suggestions)
// 	}
// }
