package sql_server_checker

import (
	"Monitor/config"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

const DbCheckIntervalSeconds = 3
const SqlServerTimeoutSeconds = 10

type SqlServerChecker struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	cfg        *config.Config
	mutex      *sync.RWMutex
	writeChan  chan string
	checkChan  chan string
	checkQueue []string
	messageMap map[string]string
}

func NewSqlServerChecker(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg *config.Config,
	mutex *sync.RWMutex,
	writeChan chan string,
	checkChan chan string,
) *SqlServerChecker {
	messageMap := make(map[string]string)
	return &SqlServerChecker{
		ctx:        ctx,
		wg:         wg,
		cfg:        cfg,
		mutex:      mutex,
		writeChan:  writeChan,
		checkChan:  checkChan,
		messageMap: messageMap,
	}
}

func (ssc *SqlServerChecker) Start() {
	go func() {
		ticker := time.NewTicker(time.Second * DbCheckIntervalSeconds)
		defer ticker.Stop()

		for {
			select {
			case <-ssc.ctx.Done():
				log.Printf("Sql Server Checker stopped")
				return
			case <-ticker.C:
				log.Printf("Sql Server Checker tick \n")
				ssc.checkSqlServerTable()
			case message := <-ssc.checkChan:
				log.Printf("Sql Server Checker: Chan Message: %s\n", message)
				var result map[string]any
				if err := json.Unmarshal([]byte(message), &result); err != nil {
					panic(err)
				}
				// TODO: check names
				table := fmt.Sprintf("[%s].[%s]", result["SourceSchema"], result["SourceTable"])
				ssc.checkQueue = append(ssc.checkQueue, table)
				ssc.messageMap[table] = message
			}
		}
	}()
	log.Printf("Sql Server Checker started ...")
}

func (ssc *SqlServerChecker) checkSqlServerTable() {

	ssc.mutex.RLock()
	user := ssc.cfg.SqlServerConfig.User
	password := ssc.cfg.SqlServerConfig.Password
	host := ssc.cfg.SqlServerConfig.Host
	port := ssc.cfg.SqlServerConfig.Port
	database := ssc.cfg.SqlServerConfig.Database
	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", user, password, host, port, database)
	ssc.mutex.RUnlock()

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*SqlServerTimeoutSeconds)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, table := range ssc.checkQueue {
		query := fmt.Sprintf("SELECT MAX(ID) AS MAX_ID FROM %s WITH (NOLOCK);", table)
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var id sql.NullInt64
			if err := rows.Scan(&id); err != nil {
				log.Fatal(err)
			}
			if !id.Valid {
				// TODO: Log to postgres DB
				log.Printf("ID is NULL\n")
				continue
			}

			ssc.writeChan <- ssc.messageMap[table]
			ssc.checkQueue = removeFirstString(ssc.checkQueue, table)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		err = rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}
}

func removeFirstString(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
