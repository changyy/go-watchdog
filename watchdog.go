package watchdog

import (
    "log"
    "fmt"
    "database/sql"
    "sort"
    "sync"
    "crypto/sha256"
    "encoding/json"
    _ "github.com/mattn/go-sqlite3"
)

const (
    dbTableWatchdogTarget = "target"
    dbTableWatchdogTargetLog = "target_log"

    // https://www.sqlite.org/lang_createtable.html
    dbSchemaCreataTableTarget = `
        CREATE TABLE IF NOT EXISTS ` + dbTableWatchdogTarget + ` (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            resourceId TEXT NOT NULL,
            resourceContent TEXT NULL,
            resourceHeader TEXT NULL,
            resourceCookie TEXT NULL,
            resourceChecksum TEXT NOT NULL,
            resourceQuery TEXT NULL,
            resourceQueryChecksum TEXT NOT NULL,
            flag TEXT NULL,
            createtime DATETIME DEFAULT CURRENT_TIMESTAMP,
            updatetime DATETIME DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT u_id UNIQUE (resourceId,resourceChecksum,resourceQueryChecksum)
        );
    `

    dbSchemaCreataTableTargetLog = `
        CREATE TABLE IF NOT EXISTS ` + dbTableWatchdogTargetLog + ` (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            resourceId TEXT NOT NULL,
            resourceContent TEXT NULL,
            resourceHeader TEXT NULL,
            resourceCookie TEXT NULL,
            resourceChecksum TEXT NOT NULL,
            resourceQuery TEXT NULL,
            resourceQueryChecksum TEXT NOT NULL,
            flag TEXT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `

    // https://sqlite.org/lang_conflict.html
    dbSchemaInsertTarget = `
        INSERT OR IGNORE INTO ` + dbTableWatchdogTarget + ` (resourceId, resourceContent, resourceHeader, resourceCookie, resourceChecksum, resourceQuery, resourceQueryChecksum) VALUES (?, ?, ?, ?, ?, ?, ?)
    `
    dbSchemaUpdateTargetTimstamp = `UPDATE ` + dbTableWatchdogTarget + ` SET updatetime = CURRENT_TIMESTAMP WHERE resourceId = ? AND resourceChecksum = ? AND resourceQueryChecksum = ? `

    dbSchemaSelectTargetRecordIDByChecksum = `SELECT id FROM ` + dbTableWatchdogTarget + ` WHERE resourceId = ? AND resourceChecksum = ? AND resourceQueryChecksum = ? `

    dbSchemaInsertTargetLog = `
        INSERT OR IGNORE INTO ` + dbTableWatchdogTargetLog + ` (resourceId, resourceContent, resourceHeader, resourceCookie, resourceChecksum, resourceQuery, resourceQueryChecksum) VALUES (?, ?, ?, ?, ?, ?, ?)
    `

    HTTPGetRequest = "GET"
    HTTPPostRequest = "POST"
    HTTPHeadRequest = "HEAD"
)

type watchdogTargetLog struct {
    id int                              `db:"id"`
    resourceId string                   `db:"resourceID"`
    resourceContent string              `db:"resourceContent"`  
    resourceHeader string               `db:"resourceHeader"`  
    resourceCookie string               `db:"resourceCookie"`  
    resourceChecksum string             `db:"resourceChecksum"`
    resourceQuery string                `db:"resourceQuery"`
    resourceQueryChecksum string        `db:"resourceChecmsum"`
    flag string                         `db:"flag"`
}

type WatchdogTarget struct {
    requestURL string
    requestMethod string
    requestHeader map[string]string
    requestCookie map[string]string
    queryChecksum string
}

type WatchdogRequestChecksumHandler func(requestURL, requestMethod string, requestHeader, requestCookie map[string]string) string
type WatchdogResponseChecksumHandler func(responseURL, responseBody string, responseHeader, responseCookie map[string]string) string

func DefaultWatchdogRequestChecksumHandler(requestURL, requestMethod string, requestHeader, requestCookie map[string]string) (output string) {
    output = requestURL + "\t" + requestMethod + "\t"
    if requestHeader != nil {
        keys := make([]string, len(requestHeader))
        i := 0
        for k := range(requestHeader) {
            keys[i] = k
            i++
        }
        sort.Strings(keys)
        for _, k := range(keys) {
            output = output +"\t" + k + "\t" + requestHeader[k]
        }
    }
    output += "\t"
    if requestCookie != nil {
        keys := make([]string, len(requestCookie))
        i := 0
        for k := range(requestCookie) {
            keys[i] = k
            i++
        }
        sort.Strings(keys)
        for _, k := range(keys) {
            output = output +"\t" + k + "\t" + requestCookie[k]
        }
    }
    output = fmt.Sprintf("%x", sha256.Sum256([]byte(output)))
    return
}

func DefaultWatchdogResponseChecksumHandler(responseURL, responseContent string, responseHeader, responseCookie map[string]string) (output string) {
    output = responseURL + "\t" + responseContent + "\t"
    if responseHeader != nil {
        keys := make([]string, len(responseHeader))
        i := 0
        for k := range(responseHeader) {
            keys[i] = k
            i++
        }
        sort.Strings(keys)
        for _, k := range(keys) {
            output = output +"\t" + k + "\t" + responseHeader[k]
        }
    }
    output += "\t"
    if responseCookie != nil {
        keys := make([]string, len(responseCookie))
        i := 0
        for k := range(responseCookie) {
            keys[i] = k
            i++
        }
        sort.Strings(keys)
        for _, k := range(keys) {
            output = output +"\t" + k + "\t" + responseCookie[k]
        }
    }
    output = fmt.Sprintf("%x", sha256.Sum256([]byte(output)))
    return
}

type Watchdog struct {
    dbResource string
    dbHandler *sql.DB
    dbMutex sync.Mutex
    requestChecksumFunc WatchdogRequestChecksumHandler
    responseChecksumFunc WatchdogResponseChecksumHandler
}

func (t *Watchdog) InitDBWithFilePath (file string) (output bool) {
    t.dbResource = file
    return t.InitDB()
}

func (t *Watchdog) InitDBWithInMemory () (output bool) {
    t.dbResource = ":memory:"
    return t.InitDB()
}

func (t *Watchdog) CloseDB() (output bool) {
    t.dbMutex.Lock()

    if t.dbResource == "" {
        output = false
    } else if t.dbHandler != nil {
        if t.dbResource == ":memory:" {
            log.Println("[WARNING] Use InitDBWithInMemory method")
        }
        t.dbHandler.Close()
        t.dbHandler = nil
    }
    output = true
    t.dbMutex.Unlock()
    return
}

func (t *Watchdog) InitDB () (output bool) {
    t.dbMutex.Lock()
    if t.dbResource == "" {
        log.Println("[WARNING] Use InitDBWithInMemory method")
        t.dbResource = ":memory:"
    }

    db, err := sql.Open("sqlite3", t.dbResource)
    if err != nil {
        log.Fatalln(err)
    }
    t.dbHandler = db
    if t.dbResource != ":memory:" {
        //defer db.Close()
        defer func(){
            t.dbHandler.Close()
            t.dbHandler = nil
        }()
    }

    output = true
    for _, createTable := range([]string{
        dbSchemaCreataTableTarget,
        dbSchemaCreataTableTargetLog,
    }) {
        if stmt, err := t.dbHandler.Prepare(createTable); err != nil {
            log.Println(err)
        } else {
            defer stmt.Close()
            if _, err = stmt.Exec(); err != nil {
                log.Println(err)
                output = false
            }
        }
    }
    t.dbMutex.Unlock()
    return 
}

func (t *Watchdog) Watch (requestURL, requestMethod string, requestHeader, requestCookie map[string]string, responseURL, responseContent string, responseHeader, responseCookie map[string]string) (output bool) {

    requestChecksum := ""
    if t.requestChecksumFunc != nil {
        requestChecksum = t.requestChecksumFunc(requestURL, requestMethod, requestHeader, requestCookie)
    } else {
        requestChecksum = DefaultWatchdogRequestChecksumHandler(requestURL, requestMethod, requestHeader, requestCookie)
    }

    responseChecksum := ""
    if t.responseChecksumFunc != nil {
        responseChecksum = t.responseChecksumFunc(responseURL, responseContent, requestHeader, requestCookie)
    } else {
        responseChecksum = DefaultWatchdogResponseChecksumHandler(responseURL, responseContent, requestHeader, requestCookie)
    }

    queryInfoJSON, err := json.Marshal(map[string]interface{}{
        "request_url": requestURL, 
        "request_header": requestHeader,
        "request_cookie": requestCookie,
    })
    if err != nil {
        log.Println("queryInfoJSON error:", err)
        output = false
        return
    }

    responseHeaderJSON, err := json.Marshal(responseHeader)
    if err != nil {
        log.Println("responseHeaderJSON error:", err)
        output = false
        return output
    }

    responseCookieJSON, err := json.Marshal(responseCookie)
    if err != nil {
        log.Println("responseCookieJSON error:", err)
        output = false
        return output
    }

    if t.dbResource == "" {
        log.Println("[ERROR] Use InitDB First")
        output = false
        return 
    }

    t.dbMutex.Lock()
    defer t.dbMutex.Unlock()

    if t.dbHandler == nil {
        db, err := sql.Open("sqlite3", t.dbResource)
        if err != nil {
            log.Println("[ERROR] sql.Open error:", err)
            output = false
            return
        }
        t.dbHandler = db
    }

    if stmt, err := t.dbHandler.Prepare(dbSchemaSelectTargetRecordIDByChecksum); err != nil {
            log.Println(err, ", raw: ", dbSchemaSelectTargetRecordIDByChecksum)
            output = false
            return
    } else {
        defer stmt.Close()
        recordId := ""
        if err := stmt.QueryRow(requestURL, responseChecksum, requestChecksum).Scan(&recordId); err != nil {
            if err != sql.ErrNoRows {
                log.Println("stmt.Exec error:", err)
                output = false
                return
            } else {
                // Insert
                if insertStmt, err := t.dbHandler.Prepare(dbSchemaInsertTarget); err != nil {
                    log.Println(err, ", raw: ", dbSchemaInsertTarget)
                    output = false
                    return
                } else {
                    defer insertStmt.Close()
                    if res, err := insertStmt.Exec(requestURL, responseContent, responseHeaderJSON, responseCookieJSON, responseChecksum, queryInfoJSON, requestChecksum); err != nil {
                        log.Println("insertStmt.Exec error:", err)
                        output = false
                        return
                    } else if recordID, err := res.LastInsertId(); err != nil {
                        log.Println("insertStmt res.LastInsertId() error:", err)
                        output = false
                        return output
                    } else if recordID == 0 {   // Need Update
                        log.Println("insertStmt res.LastInsertId() == 0")
                        output = false
                        return output
                    } else {
                        output = true
                    }
                }
            }
        } else {
            // Updated
            if updateStmt, err := t.dbHandler.Prepare(dbSchemaUpdateTargetTimstamp); err != nil {
                log.Println("db.Prepare dbSchemaUpdateTargetTimstamp error:", err, ", sql:", dbSchemaUpdateTargetTimstamp)
                output = false
                return output
            } else {
                defer updateStmt.Close()
                if _, err = updateStmt.Exec(requestURL, responseChecksum, requestChecksum); err != nil {
                    log.Println("dbSchemaUpdateTargetTimstamp execute error:", err)
                    output = false
                    return output
                }
                output = true
            }
        }
    }
    return
    // INSERT and Update
    //if stmt, err := t.dbHandler.Prepare(dbSchemaInsertTarget); err != nil {
    //        log.Println(err, ", raw: ", dbSchemaInsertTarget)
    //        output = false
    //        return
    //} else {
    //        defer stmt.Close()
    //        
    //        if res, err := stmt.Exec(requestURL, responseContent, responseHeaderJSON, responseCookieJSON, responseChecksum, queryInfoJSON, requestChecksum); err != nil {
    //            log.Println("stmt.Exec error:", err)
    //            output = false
    //            return
    //        } else if recordID, err := res.LastInsertId(); err != nil {
    //            log.Println("res.LastInsertId() error:", err)
    //            output = false
    //            return output
    //        } else if recordID == 0 {   // Need Update
    //            if updateStmt, err := t.dbHandler.Prepare(dbSchemaUpdateTargetTimstamp); err != nil {
    //                log.Println("db.Prepare dbSchemaUpdateTargetTimstamp error:", err, ", sql:", dbSchemaUpdateTargetTimstamp)
    //                output = false
    //                return output
    //            } else {
    //                defer updateStmt.Close()
    //                if _, err = updateStmt.Exec(requestURL, responseChecksum, requestChecksum); err != nil {
    //                    log.Println("dbSchemaUpdateTargetTimstamp execute error:", err)
    //                    output = false
    //                    return output
    //                }
    //            }
    //        }
    //        output = true
    //}
    //return
}

