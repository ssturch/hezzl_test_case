package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"hzzl/internal/db"
	mb "hzzl/internal/msgbroker"
	"io"
	"net/http"
	"strconv"
)

// Хэндлер для /item/create/
func createHandler(stmt db.SQLstmpQueries, rds db.RedisClient, nts mb.NatsConn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chk bool
		var campIdRaw string
		var campId int
		var payloadData map[string]any
		var jsonRequest []byte

		campIdRaw = r.URL.Query().Get("campaignId")
		if r.Method == "POST" {
			if campIdRaw == "" {
				err = errors.New("URL parameter 'campaignId' is empty or not initialized!")
				fmt.Println(err)
				return
			}

			campId, err = strconv.Atoi(campIdRaw)

			if err != nil {
				err = errors.New("URL parameter 'campaignId' is not integer!")
				fmt.Println(err)
				return
			}

			chk, err = stmt.CheckIDQuery_campaigns(campId)
			if !chk {
				err = errors.New("This 'campaignId' is not available in DB!")
				fmt.Println(err)
				return
			}

			payloadRaw, _ := io.ReadAll(r.Body)
			if err = json.Unmarshal(payloadRaw, &payloadData); err != nil {
				err = errors.New("JSON data from BODY is wrong")
				r.Body.Close()
				fmt.Println(err)
				return
			}
			r.Body.Close()

			name, ok := payloadData["name"].(string)
			if !ok {
				err = errors.New("JSON data don't have key 'name'!")
				r.Body.Close()
				fmt.Println(err)
				return
			}

			err = stmt.PostQuery(campId, name, false)
			nts.PublishMessage(r.RemoteAddr, "POST", r.URL.Path, err)
			if err != nil {
				fmt.Println(err)
				return
			}
			rds.Invalidate()
			jsonRequest, err = stmt.RequestQuery(campId, name)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(jsonRequest)
		}
	}
}

// Хэндлер для /item/update
func updateHandler(stmt db.SQLstmpQueries, rds db.RedisClient, nts mb.NatsConn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chk bool
		var idRaw, campIdRaw string
		var id, campId int
		var payloadData map[string]any
		var jsonRequest []byte

		idRaw = r.URL.Query().Get("id")
		campIdRaw = r.URL.Query().Get("campaignId")

		if r.Method == "PATCH" {
			if campIdRaw == "" {
				err = errors.New("URL parameter 'campaignId' is empty or not initialized!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}
			if idRaw == "" {
				err = errors.New("URL parameter 'id' is empty or not initialized!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			campId, err = strconv.Atoi(campIdRaw)

			if err != nil {
				err = errors.New("URL parameter 'campaignId' is not integer!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			id, err = strconv.Atoi(idRaw)

			if err != nil {
				err = errors.New("URL parameter 'id' is not integer!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			chk, err = stmt.CheckIDQuery_items(id, campId)
			if !chk {
				err = errors.New("This 'id' and 'campaignId' is not available in DB!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			payloadRaw, _ := io.ReadAll(r.Body)
			if err = json.Unmarshal(payloadRaw, &payloadData); err != nil {
				err = errors.New("JSON data from BODY is wrong")
				fmt.Println(err)
				r.Body.Close()
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}
			r.Body.Close()

			name, ok := payloadData["name"].(string)

			if !ok || name == "" {
				err = errors.New("JSON data don't have key 'name' or value of 'name'!")
				fmt.Println(err)
				r.Body.Close()
				return
			}

			desc, ok := payloadData["description"].(string)

			if !ok {
				desc = ""
			}

			err = stmt.PatchQuery(name, desc, id, campId)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = nts.PublishMessage(r.RemoteAddr, "PATCH", r.URL.Path, err)

			if err != nil {
				fmt.Println(err)
				return
			}
			rds.Invalidate()
			jsonRequest, err = stmt.RequestQuery(campId, name)
			if err != nil {
				err = errors.New("JSON data don't have key 'name'!")
				fmt.Println(err)
				return
			}
			w.Write(jsonRequest)
		}
	}
}

// Хэндлер для /item/remove
func removeHandler(stmt db.SQLstmpQueries, rds db.RedisClient, nts mb.NatsConn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var chk bool
		var idRaw, campIdRaw string
		var id, campId int
		var jsonRequest []byte

		idRaw = r.URL.Query().Get("id")
		campIdRaw = r.URL.Query().Get("campaignId")

		if r.Method == "PATCH" {
			if campIdRaw == "" {
				err = errors.New("URL parameter 'campaignId' is empty or not initialized!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}
			if idRaw == "" {
				err = errors.New("URL parameter 'id' is empty or not initialized!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			campId, err = strconv.Atoi(campIdRaw)

			if err != nil {
				err = errors.New("URL parameter 'campaignId' is not integer!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			id, err = strconv.Atoi(idRaw)

			if err != nil {
				err = errors.New("URL parameter 'id' is not integer!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			chk, err = stmt.CheckIDQuery_items(id, campId)
			if !chk {
				err = errors.New("This 'id' and 'campaignId' is not available in DB!")
				fmt.Println(err)
				Return404Response(&w, 3, "errors.item.notFound", nil)
				return
			}

			r.Body.Close()

			err = stmt.DeleteQuery(id, campId)
			nts.PublishMessage(r.RemoteAddr, "PATCH", r.URL.Path, err)
			if err != nil {
				fmt.Println(err)
				return
			}
			rds.Invalidate()
			jsonRequest, err = stmt.ShortRequestQuery(id, campId)

			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(jsonRequest)
		}
	}
}

// Хэндлер для /items/list
func listHandler(stmt db.SQLstmpQueries, rds db.RedisClient) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var rdsData []byte
		var sqlData *sql.Rows
		var JSONData []byte

		if r.Method == "PATCH" {
			defer r.Body.Close()
			rdsData, err = rds.GetData()
			if string(rdsData) == "null" || string(rdsData) == "keys *: []" || err != nil {
				sqlData, err = stmt.GetQuery()
				if err != nil {
					fmt.Println(err)
					return
				}
				JSONData, err = db.ConvertSQLRowsToJSON(sqlData)
				if err != nil {
					fmt.Println(err)
					return
				}
				rds.AddData(JSONData)
				rdsData, err = rds.GetData()

			}
			w.Write(rdsData)
		}
	}
}

func Return404Response(w *http.ResponseWriter, code int, msg string, details []byte) {
	(*w).Header().Add("code", strconv.Itoa(code))
	(*w).Header().Add("message", msg)
	(*w).Header().Add("details", string(details))
	(*w).WriteHeader(http.StatusNotFound)
}
