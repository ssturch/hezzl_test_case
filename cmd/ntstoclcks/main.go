package main

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/uptrace/go-clickhouse/ch"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type MsgToClickhouse struct {
	Time    time.Time
	Message string
}

func StartReader(ctx context.Context) error {
	var err error
	var bufferSize int

	//Стартуем nats
	bufferSize = 5
	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		fmt.Println(err)
		return err
	}

	natsChan := make(chan *nats.Msg, bufferSize)

	sub, err := nc.ChanSubscribe("list", natsChan)

	if err != nil {
		fmt.Println(err)
		return err
	}

	//Стартуем clickhouse
	db := ch.Connect(
		// clickhouse://<user>:<password>@<host>:<port>/<database>?sslmode=disable
		ch.WithDSN("clickhouse://localhost:9000/default?sslmode=disable"),
	)
	_, err = db.NewCreateTable().Model((*MsgToClickhouse)(nil)).
		Order("time").
		Exec(ctx)

	if err != nil {
		fmt.Println(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msgSlice []MsgToClickhouse
				if len(natsChan) == bufferSize {
					for len(natsChan) > 0 {
						msg := <-natsChan
						msgSlice = append(msgSlice, MsgToClickhouse{
							Time:    time.Now(),
							Message: string(msg.Data),
						})
					}

					_, err = db.NewInsert().Model(&msgSlice).Exec(ctx)

					if err != nil {
						fmt.Println(err)
					}
				}

			}
		}
		sub.Unsubscribe()
		nc.Close()
		db.Close()
		close(natsChan)
	}()
	return nil
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	exitChan := make(chan int)

	// Отслеживание системных сигналов (частично совместим с Windows)
	go func() {
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGINT:
				fmt.Println("Catch: SIGNAL INTERRUPT")
				cancel()

				exitChan <- 0
			case os.Interrupt:
				fmt.Println("Catch: SIGNAL INTERRUPT")
				cancel()
				exitChan <- 0
			case syscall.SIGTERM:
				fmt.Println("Catch: SIGNAL TERMINATE")
				cancel()
				exitChan <- 0
			case syscall.SIGKILL:
				fmt.Println("Catch: SIGNAL KILL")
				cancel()
				exitChan <- 0
			}
		}
	}()

	StartReader(ctx)
	exitCode := <-exitChan
	close(exitChan)
	os.Exit(exitCode)
}
