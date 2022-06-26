package watchdog

import (
    "os"
    "log"
    "io/ioutil"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestInitDBWithFilePath(t *testing.T) {
    watchdog := &Watchdog{}

    dbFile, err := ioutil.TempFile("/tmp", "watchdogTest.*.sqlite3")
    if err != nil {
        log.Fatal("TestInitDB failed:", err)
    }
    defer os.Remove(dbFile.Name())
    log.Println("DB temp file:", dbFile.Name())

    assert.True(t, watchdog.InitDBWithFilePath(dbFile.Name()), "InitDBWithFilePath Failed")
}

func TestInitDBWithInMemory(t *testing.T) {
    watchdog := &Watchdog{}
    assert.True(t, watchdog.InitDBWithInMemory(), "InitDBWithInMemory Failed")
}

func TestWatch(t *testing.T) {
    watchdog := &Watchdog{}
    //assert.True(t, watchdog.InitDBWithInMemory(), "InitDBWithInMemory Failed")

    dbFile, err := ioutil.TempFile("/tmp", "watchdogTest.*.sqlite3")
    if err != nil {
        log.Fatal("TestInitDB failed:", err)
    }
    defer os.Remove(dbFile.Name())
    log.Println("DB temp file:", dbFile.Name())

    assert.True(t, watchdog.InitDBWithFilePath(dbFile.Name()), "InitDBWithInMemory Failed")
    assert.True(t, watchdog.Watch(
        // Request URL
        "https://tw.yahoo.com",
        // Request Method
        HTTPGetRequest,       
        // Request Header
        map[string]string{
            "Accept-Language": "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7",
        },
        // Request Cookie
        map[string]string{},

        // Response Current URL
        "https://tw.yahoo.com",
        // Response Content
        "",
        // Response Header
        map[string]string{},
        // Response Cookie
        map[string]string{},
    ), "Watch Failed")
}
