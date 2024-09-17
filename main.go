package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

func main() {
	dsn := flag.String("db", "", "MySQL DSN")
	flag.Parse()

	if *dsn == "" {
		panic("-db is required")
	}

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		panic(err)
	}
	defer func() { _ = db.Close() }()

	runAs("A", context.Background(), db, "create table if not exists lolcat ( i int )")
	runAs("A", context.Background(), db, "insert into lolcat(i) values (1)")

	a := txAs("A", context.Background(), db)
	b := txAs("B", context.Background(), db)

	runAs("A", context.Background(), a, "select * from lolcat for update")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	runAs("B[timeout=1s]", ctx, b, "select * from lolcat for update")
}

func runAs(who string, ctx context.Context, where interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, what string) {
	if _, err := fmt.Fprintf(os.Stderr, "%s: %s\n", who, what); err != nil {
		panic(err)
	}

	if _, err := where.ExecContext(ctx, what); err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintln(os.Stderr, "DONE"); err != nil {
		panic(err)
	}
}

func txAs(who string, ctx context.Context, where *sql.DB) *sql.Tx {
	if _, err := fmt.Fprintf(os.Stderr, "%s: BEGIN\n", who); err != nil {
		panic(err)
	}

	tx, err := where.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	if _, err := fmt.Fprintln(os.Stderr, "DONE"); err != nil {
		panic(err)
	}

	return tx
}
