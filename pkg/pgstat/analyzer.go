package pgstat

import (
	"fmt"
	"strings"
)

// OptimizationSuggestion represents a suggestion for a query optimization.
type OptimizationSuggestion struct {
	QueryID     int64
	Query       string
	Suggestions []string
}

// AnalyzeQueries analyzes an array of PgStatStatementEntry and returns optimization suggestions.
func AnalyzeQueries(entries []PgStatStatementEntry) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion
	for _, entry := range entries {
		var entrySuggestions []string

		if strings.Contains(entry.Query, " pg_") || strings.Contains(entry.Query, " information_schema") {
			continue
		}

		if strings.Contains(entry.Query, "alter table") || strings.Contains(entry.Query, "create index") {
			continue
		}

		// Check for high total execution time.
		// if suggestion := checkHighTotalExecTime(entry); suggestion != "" {
		// 	entrySuggestions = append(entrySuggestions, suggestion)
		// }

		// Check for high average execution time.
		if suggestion := checkHighMeanExecTime(entry); suggestion != "" {
			entrySuggestions = append(entrySuggestions, suggestion)
		}

		// Check for high standard deviation in execution time.
		if suggestion := checkHighStdDevExecTime(entry); suggestion != "" {
			entrySuggestions = append(entrySuggestions, suggestion)
		}

		// Check for high number of temporary blocks read or written.
		if suggestion := checkTempBlksUsage(entry); suggestion != "" {
			entrySuggestions = append(entrySuggestions, suggestion)
		}

		// Check for high number of shared blocks read vs. hit.
		if suggestion := checkSharedBlksReadVsHit(entry); suggestion != "" {
			entrySuggestions = append(entrySuggestions, suggestion)
		}

		// Check for low rows returned per call.
		// if suggestion := checkRowsPerCall(entry); suggestion != "" {
		// 	entrySuggestions = append(entrySuggestions, suggestion)
		// }

		// Check for high WAL usage.
		if suggestion := checkHighWalUsage(entry); suggestion != "" {
			entrySuggestions = append(entrySuggestions, suggestion)
		}

		// Add suggestions for this query if any exist.
		if len(entrySuggestions) > 0 {
			suggestions = append(suggestions, OptimizationSuggestion{
				QueryID:     entry.QueryID,
				Query:       entry.Query,
				Suggestions: entrySuggestions,
			})
		}
	}
	return suggestions
}

// checkHighTotalExecTime checks if the total execution time is high.
func checkHighTotalExecTime(entry PgStatStatementEntry) string {
	const totalExecTimeThreshold = 1000.0 // in milliseconds
	if entry.TotalExecTime > totalExecTimeThreshold {
		return fmt.Sprintf("Total execution time is high (%.2f ms). Consider optimizing the query or adding indexes.", entry.TotalExecTime)
	}
	return ""
}

// checkHighMeanExecTime checks if the mean execution time per call is high.
func checkHighMeanExecTime(entry PgStatStatementEntry) string {
	const meanExecTimeThreshold = 100.0 // in milliseconds
	if entry.MeanExecTime > meanExecTimeThreshold {
		return fmt.Sprintf("Mean execution time per call is high (%.2f ms). Consider optimizing the query.", entry.MeanExecTime)
	}
	return ""
}

// checkHighStdDevExecTime checks if the standard deviation of execution time is high.
func checkHighStdDevExecTime(entry PgStatStatementEntry) string {
	const stdDevThreshold = 50.0 // in milliseconds
	if entry.StddevExecTime > stdDevThreshold {
		return fmt.Sprintf("Execution time varies widely (stddev %.2f ms). Investigate possible causes for inconsistent performance.", entry.StddevExecTime)
	}
	return ""
}

// checkTempBlksUsage checks if the query uses a high number of temporary blocks.
func checkTempBlksUsage(entry PgStatStatementEntry) string {
	if entry.TempBlksRead > 0 || entry.TempBlksWritten > 0 {
		return "Query uses temporary disk space. Consider optimizing to reduce disk I/O, such as adding indexes or rewriting the query."
	}
	return ""
}

// checkSharedBlksReadVsHit checks if the query reads many shared blocks compared to hits.
func checkSharedBlksReadVsHit(entry PgStatStatementEntry) string {
	totalSharedBlks := entry.SharedBlksHit + entry.SharedBlksRead
	if totalSharedBlks == 0 {
		return ""
	}
	readRatio := float64(entry.SharedBlksRead) / float64(totalSharedBlks)
	if readRatio > 25.0 {
		return fmt.Sprintf("High shared block read ratio (%.2f%%). Consider adding indexes to reduce I/O.", readRatio*100)
	}
	return ""
}

// checkRowsPerCall checks if the query returns a low number of rows per call.
func checkRowsPerCall(entry PgStatStatementEntry) string {
	if entry.Calls == 0 {
		return ""
	}
	rowsPerCall := float64(entry.Rows) / float64(entry.Calls)
	if rowsPerCall < 10.0 {
		return fmt.Sprintf("Low rows returned per call (%.2f). Verify if the query returns the expected results.", rowsPerCall)
	}
	return ""
}

// checkHighWalUsage checks if the query generates a high amount of WAL records or bytes.
func checkHighWalUsage(entry PgStatStatementEntry) string {
	const walBytesThreshold = 1024 * 1024 * 50 // 50 MB
	if entry.WalBytes > walBytesThreshold {
		return fmt.Sprintf("High WAL usage (%.2f MB). Consider batching writes or optimizing the query.", float64(entry.WalBytes)/(1024*1024))
	}
	return ""
}
