package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

// структура подготовленных выражений
type SQLstmpQueries struct {
	DB                  *sql.DB
	ctx                 context.Context
	stmtREQUEST         *sql.Stmt
	stmtSHORT_REQUEST   *sql.Stmt
	stmtCHECK_campaigns *sql.Stmt
	stmtCHECK_items     *sql.Stmt
	stmtPOST            *sql.Stmt
	stmtPATCH           *sql.Stmt
	stmtDELETE          *sql.Stmt
	stmtGET             *sql.Stmt
}

// Инициализация таблиц и выражений
func (m *SQLstmpQueries) Prepare() error {
	// Список основных запросов
	const (
		sqlCheckCampaigns = "SELECT EXISTS(SELECT 1 FROM campaigns WHERE id = $1)"
		sqlCheckItems     = "SELECT EXISTS(SELECT 1 FROM items WHERE id = $1 AND campaign_id = $2)"
		sqlPost           = "INSERT INTO items (campaign_id, name, removed) VALUES ($1, $2, $3)"
		sqlRequest        = "SELECT * FROM items WHERE campaign_id = $1 AND name = $2"
		sqlShortRequest   = "SELECT id, campaign_id, removed FROM items WHERE id = $1 AND campaign_id = $2"
		sqlPatch          = "UPDATE items SET name = $1, description = $2 WHERE id = $3 AND campaign_id = $4"
		sqlDelete         = "UPDATE items SET removed = TRUE WHERE id = $1 AND campaign_id = $2"
		sqlGet            = "SELECT * FROM items"
	)

	//Создание таблиц
	var err error
	for _, v := range CreateTables() {
		_, err = m.DB.Exec(v)
		if err != nil {
			break
			return err
		}
	}

	// Инициализация подготовленных выражений
	m.stmtPOST, err = m.DB.Prepare(sqlPost)
	if err != nil {
		return err
	}
	m.stmtPATCH, err = m.DB.Prepare(sqlPatch)
	if err != nil {
		return err
	}
	m.stmtDELETE, err = m.DB.Prepare(sqlDelete)
	if err != nil {
		return err
	}
	m.stmtGET, err = m.DB.Prepare(sqlGet)
	if err != nil {
		return err
	}
	m.stmtCHECK_campaigns, err = m.DB.Prepare(sqlCheckCampaigns)
	if err != nil {
		return err
	}
	m.stmtCHECK_items, err = m.DB.Prepare(sqlCheckItems)
	if err != nil {
		return err
	}
	m.stmtREQUEST, err = m.DB.Prepare(sqlRequest)
	if err != nil {
		return err
	}
	m.stmtSHORT_REQUEST, err = m.DB.Prepare(sqlShortRequest)
	if err != nil {
		return err
	}
	return nil
}

// Метод для закрытия подготовленных выражений
func (m *SQLstmpQueries) Close() error {
	var anyErr error
	for _, stmt := range []*sql.Stmt{m.stmtCHECK_items, m.stmtCHECK_campaigns, m.stmtPOST, m.stmtPATCH, m.stmtDELETE, m.stmtGET, m.stmtREQUEST, m.stmtSHORT_REQUEST} {
		if stmt == nil {
			continue
		}
		if err := stmt.Close(); anyErr != nil {
			anyErr = err
		}
	}
	return anyErr
}

// Функция для подготовки выбранной БД
func PrepareDB(db *sql.DB, ctx context.Context) (*SQLstmpQueries, error) {
	m := &SQLstmpQueries{DB: db, ctx: ctx}
	err := m.Prepare()
	if err != nil {
		m.Close()
		return nil, err
	}
	return m, nil
}

// Список запросов для создания таблиц
func CreateTables() []string {
	query1 := "CREATE TABLE IF NOT EXISTS campaigns (" +
		"id SERIAL PRIMARY KEY," +
		"name varchar NOT NULL);"
	query2 := "CREATE INDEX IF NOT EXISTS campaigns_idx ON campaigns (id);"
	query3 := "INSERT INTO campaigns (id, name)" +
		"SELECT 1, 'Первая запись' WHERE NOT EXISTS" +
		"(SELECT id, name FROM campaigns WHERE id = 1 AND name = 'Первая запись');"
	query4 := "CREATE SEQUENCE IF NOT EXISTS priority_increment START 1;"
	query5 := "CREATE TABLE IF NOT EXISTS items (" +
		"id SERIAL," +
		"campaign_id SERIAL," +
		"PRIMARY KEY (id, campaign_id)," +
		"FOREIGN KEY (campaign_id) REFERENCES campaigns (id)," +
		"name varchar NOT NULL," +
		"description text," +
		"priority integer NOT NULL DEFAULT nextval('priority_increment')," +
		"removed bool NOT NULL," +
		"created_at timestamp NOT NULL DEFAULT NOW());"
	query6 := "CREATE INDEX IF NOT EXISTS items_idx ON items (id, campaign_id, name);"
	queryArr := []string{query1, query2, query3, query4, query5, query6}
	return queryArr
}

// Методы для выполнения запросов
func (m *SQLstmpQueries) CheckIDQuery_campaigns(id int) (bool, error) {
	var res any
	resRaw := m.stmtCHECK_campaigns.QueryRowContext(m.ctx, id)
	err := resRaw.Scan(&res)
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}
func (m *SQLstmpQueries) CheckIDQuery_items(id int, campId int) (bool, error) {
	var res any
	resRaw := m.stmtCHECK_items.QueryRowContext(m.ctx, id, campId)
	err := resRaw.Scan(&res)
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}
func (m *SQLstmpQueries) PostQuery(id int, name string, removed bool) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txStmt := tx.StmtContext(m.ctx, m.stmtPOST)
	_, err = txStmt.Exec(id, name, removed)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (m *SQLstmpQueries) PatchQuery(name, desc string, id, campId int) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txStmt := tx.StmtContext(m.ctx, m.stmtPATCH)
	_, err = txStmt.Exec(name, desc, id, campId)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (m *SQLstmpQueries) DeleteQuery(id, campId int) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txStmt := tx.StmtContext(m.ctx, m.stmtDELETE)
	_, err = txStmt.Exec(id, campId)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (m *SQLstmpQueries) GetQuery() (*sql.Rows, error) {
	resRaw, err := m.stmtGET.QueryContext(m.ctx)
	if err != nil {
		return nil, err
	}
	return resRaw, nil
}
func (m *SQLstmpQueries) RequestQuery(id int, name string) ([]byte, error) {
	resRaw, err := m.stmtREQUEST.QueryContext(m.ctx, id, name)
	if err != nil {
		return nil, err
	}
	res, err := ConvertSQLRowsToJSON(resRaw)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (m *SQLstmpQueries) ShortRequestQuery(id, campId int) ([]byte, error) {
	resRaw, err := m.stmtSHORT_REQUEST.QueryContext(m.ctx, id, campId)
	if err != nil {
		return nil, err
	}
	res, err := ConvertSQLRowsToJSON(resRaw)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Для отладки
const (
	hostPG     = "localhost"
	portPG     = 5432
	userPG     = "postgres"
	passwordPG = "qwerty"
	dbnamePG   = "db_hezzlTestApi"
)

// Создание соединения с БД и ее подготовка
func Pgdbconnect() (*sql.DB, error) {
	var pgdb *sql.DB
	var err error
	pgdbconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", hostPG, portPG, userPG, passwordPG, dbnamePG) //для отладки
	//pgdbconn := "postgresql://postgres:qwerty@clair_postgres:5432?sslmode=disable"
	pgdb, err = sql.Open("postgres", pgdbconn)
	if err != nil {
		return pgdb, err
	}
	pgdb.SetConnMaxIdleTime(250 * time.Millisecond)
	pgdb.SetConnMaxLifetime(50 * time.Millisecond)
	return pgdb, nil
}
