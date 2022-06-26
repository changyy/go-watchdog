package watchdog

import (
    "log"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

const (
    dbTableWatchdogTarget = "target"
    dbTableWatchdogTargetLog = "target_log"

    // https://www.sqlite.org/lang_createtable.html
    dbSchemaCreataTableTarget = `
        CREATE TABLE IF NOT EXISTS ` + dbTableWatchdogTarget + ` (
            id INTEGER PRIMARY KEY,
            resourceId TEXT NOT NULL,
            resourceContent TEXT NULL,
            resourceChecksum TEXT NOT NULL,
            resourceQuery TEXT NULL,
            resourceQueryChecksum TEXT NOT NULL,
            flag TEXT NULL,
            createtime DATETIME DEFAULT CURRENT_TIMESTAMP,
            updatetime DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE UNIQUE INDEX u_id ON ? (resourceId,resourceChecksum,resourceQueryChecksum);
    `
    dbSchemaCreataTableTargetLog = `
        CREATE TABLE IF NOT EXISTS ` + dbTableWatchdogTargetLog + ` (
            id INTEGER PRIMARY KEY,
            resourceId TEXT NOT NULL,
            resourceContent TEXT NULL,
            resourceChecksum TEXT NOT NULL,
            resourceQuery TEXT NULL,
            resourceQueryChecksum TEXT NOT NULL,
            flag TEXT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `

    HTTPGetRequest = "GET"
    HTTPPostRequest = "POST"
    HTTPHeadRequest = "HEAD"
)

type watchdogTargetLog struct {
    id int                              `db:"id"`
    resourceId string                   `db:"resourceID"`
    resourceContent string              `db:"resourceContent"`  
    resourceChecksum string             `db:"resourceChecksum"`
    resourceQuery string                `db:"resourceQuery"`
    resourceQueryChecksum string        `db:"resourceChecmsum"`
    flag string                         `db:"flag"`
}

type WatchdogTarget struct {
    requestURL string
    requestHeader map[string]string
    requestMethod string
    queryChecksum string
}

type Watchdog struct {
    dbPath string
}

func (t *Watchdog) InitDBWithFilePath (file string) (output bool) {
    return t.InitDB()
}

func (t *Watchdog) InitDBWithInMemory () (output bool) {
    t.dbPath = ":memory:"
    return t.InitDB()
}

func (t *Watchdog) InitDB () (output bool) {
    if t.dbPath == "" {
        log.Println("[WARNING] Use InitDBWithInMemory method")
        t.dbPath = ":memory:"
    }

    db, err := sql.Open("sqlite3", t.dbPath)
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()

    output = true
    for _, createTable := range([]string{
        dbSchemaCreataTableTarget,
        dbSchemaCreataTableTargetLog,
    }) {
        if stmt, err := db.Prepare(createTable); err != nil {
            log.Println(err)
        } else {
            defer stmt.Close()
            if _, err = stmt.Exec(); err != nil {
                log.Println(err)
                output = false
            }
        }
    }
    return 
}

func (t *Watchdog) Watch (target WatchdogTarget) (output bool)  {
    output = true

    switch target.requestMethod {
        case HTTPGetRequest:
        case HTTPPostRequest:
        case HTTPHeadRequest:
        default:
            log.Println("target.requestMethod failed")
            output = false
            return
    }
    
    return 
}
