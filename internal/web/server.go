package web

import (
	"context"
	"fmt"
	"hzzl/internal/db"
	mb "hzzl/internal/msgbroker"
	"net/http"
)

const (
	Host = "localhost"
	Port = "8080"
)

// Отслеживание endpoint'ов
func Server(ctx context.Context, stmp *db.SQLstmpQueries, rds *db.RedisClient, nts *mb.NatsConn) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var err error
				http.HandleFunc("/item/create", createHandler(*stmp, *rds, *nts))
				http.HandleFunc("/item/update", updateHandler(*stmp, *rds, *nts))
				http.HandleFunc("/item/remove", removeHandler(*stmp, *rds, *nts))
				http.HandleFunc("/items/list", listHandler(*stmp, *rds))
				//realhost, _ := os.Hostname()
				realhost := Host // для отладки
				err = http.ListenAndServe(realhost+":"+Port, nil)
				//err = http.ListenAndServe(realhost, nil)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}
