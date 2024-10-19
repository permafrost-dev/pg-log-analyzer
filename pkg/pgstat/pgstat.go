package pgstat

import (
	//"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PgStatStatementEntry represents a single entry from pg_stat_statements.
type PgStatStatementEntry struct {
	UserID               int64   `db:"userid"`
	DbID                 int64   `db:"dbid"`
	TopLevel             bool    `db:"toplevel"`
	QueryID              int64   `db:"queryid"`
	Query                string  `db:"query"`
	Plans                int64   `db:"plans"`
	TotalPlanTime        float64 `db:"total_plan_time"`
	MinPlanTime          float64 `db:"min_plan_time"`
	MaxPlanTime          float64 `db:"max_plan_time"`
	MeanPlanTime         float64 `db:"mean_plan_time"`
	StddevPlanTime       float64 `db:"stddev_plan_time"`
	Calls                int64   `db:"calls"`
	TotalExecTime        float64 `db:"total_exec_time"`
	MinExecTime          float64 `db:"min_exec_time"`
	MaxExecTime          float64 `db:"max_exec_time"`
	MeanExecTime         float64 `db:"mean_exec_time"`
	StddevExecTime       float64 `db:"stddev_exec_time"`
	Rows                 int64   `db:"rows"`
	SharedBlksHit        int64   `db:"shared_blks_hit"`
	SharedBlksRead       int64   `db:"shared_blks_read"`
	SharedBlksDirtied    int64   `db:"shared_blks_dirtied"`
	SharedBlksWritten    int64   `db:"shared_blks_written"`
	LocalBlksHit         int64   `db:"local_blks_hit"`
	LocalBlksRead        int64   `db:"local_blks_read"`
	LocalBlksDirtied     int64   `db:"local_blks_dirtied"`
	LocalBlksWritten     int64   `db:"local_blks_written"`
	TempBlksRead         int64   `db:"temp_blks_read"`
	TempBlksWritten      int64   `db:"temp_blks_written"`
	BlkReadTime          float64 `db:"blk_read_time"`
	BlkWriteTime         float64 `db:"blk_write_time"`
	TempBlkReadTime      float64 `db:"temp_blk_read_time"`
	TempBlkWriteTime     float64 `db:"temp_blk_write_time"`
	WalRecords           int64   `db:"wal_records"`
	WalFPI               int64   `db:"wal_fpi"`
	WalBytes             int64   `db:"wal_bytes"`
	JitFunctions         int64   `db:"jit_functions"`
	JitGenerationTime    float64 `db:"jit_generation_time"`
	JitInliningCount     int64   `db:"jit_inlining_count"`
	JitInliningTime      float64 `db:"jit_inlining_time"`
	JitOptimizationCount int64   `db:"jit_optimization_count"`
	JitOptimizationTime  float64 `db:"jit_optimization_time"`
	JitEmissionCount     int64   `db:"jit_emission_count"`
	JitEmissionTime      float64 `db:"jit_emission_time"`
}

// GetPgStatStatements fetches the pg_stat_statements from the database.
func GetPgStatStatements(db *sqlx.DB) ([]PgStatStatementEntry, error) {
	var entries []PgStatStatementEntry
	query := `SELECT * FROM pg_stat_statements ORDER BY calls DESC;`
	err := db.Select(&entries, query)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func BuildPostgresDsn(host string, port int, user string, password string, dbname string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

func ConnectAndFetchPgStatStatements(dsn string) ([]PgStatStatementEntry, error) {
	db, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	return GetPgStatStatements(db)
}
