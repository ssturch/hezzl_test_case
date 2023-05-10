package main

import (
	"context"
	"fmt"
	dbp "hzzl/internal/db"
	mb "hzzl/internal/msgbroker"
	wb "hzzl/internal/web"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	pgdb, err := dbp.Pgdbconnect()

	if err != nil {
		fmt.Println(err)
	}

	stmp, err := dbp.PrepareDB(pgdb, ctx)

	if err != nil {
		fmt.Println(err)
	}

	rdsClient := dbp.PrepareRedis(ctx)

	ntsConn, err := mb.PrepareNats()

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	exitChan := make(chan int)

	// Отслеживание системных сигналов (частично совместим с Windows)
	go func() {
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGINT:
				fmt.Println("Catch: SIGNAL INTERRUPT | Server stopped | DB Closed")
				cancel()
				pgdb.Close()
				ntsConn.Close()
				exitChan <- 0
			case os.Interrupt:
				fmt.Println("Catch: SIGNAL INTERRUPT | Server stopped | DB Closed")
				cancel()
				pgdb.Close()
				ntsConn.Close()
				exitChan <- 0
			case syscall.SIGTERM:
				fmt.Println("Catch: SIGNAL TERMINATE | Server stopped | DB Closed")
				cancel()
				pgdb.Close()
				ntsConn.Close()
				exitChan <- 0
			case syscall.SIGKILL:
				fmt.Println("Catch: SIGNAL KILL | Server stopped | DB Closed")
				cancel()
				pgdb.Close()
				ntsConn.Close()
				exitChan <- 0
			}
		}
	}()

	// Отслеживание входящих запросов по REST API
	wb.Server(ctx, stmp, rdsClient, ntsConn)

	exitCode := <-exitChan
	close(exitChan)
	os.Exit(exitCode)
}
