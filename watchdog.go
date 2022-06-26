package watchdog

import (
    "log"
    "fmt"
    "database/sql"
    "sort"
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
            updatetime DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE UNIQUE INDEX u_id ON ? (resourceId,resourceChecksum,resourceQueryChecksum);
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

func (t *Watchdog) InitDB () (output bool) {
    if t.dbResource == "" {
        log.Println("[WARNING] Use InitDBWithInMemory method")
        t.dbResource = ":memory:"
    }

    db, err := sql.Open("sqlite3", t.dbResource)
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

    if t.dbResource == "" {
        log.Println("[ERROR] Use InitDB First")
        output = false
        return 
    }

    db, err := sql.Open("sqlite3", t.dbResource)
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()

    if stmt, err := db.Prepare(dbSchemaInsertTarget); err != nil {
            log.Println(err, ", raw: ", dbSchemaInsertTarget)
            output = false
            return output
    } else {
            defer stmt.Close()

            queryInfo := map[string]interface{}{
                "url": requestURL, 
                "header": requestHeader,
                "cookie": requestCookie,
            }
            queryInfoJSON, err := json.Marshal(queryInfo)
             if err != nil {
                log.Println("queryInfoJSON error:", err)
                output = false
                return output
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
            
            if _, err = stmt.Exec(requestURL, responseContent, responseHeaderJSON, responseCookieJSON, responseChecksum, queryInfoJSON, requestChecksum); err != nil {
                log.Println("stmt.Exec error:", err)
                output = false
                return output
            } else {
                output = true
            }
    }

    return
}

