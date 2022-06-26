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
    assert.True(t, watchdog.InitDBWithInMemory(), "InitDBWithInMemory Failed")
    /*
    watchTarget := &WatchdogTarget{
        "https://tw.yahoo.com",
        map[string]string{
            "Accept-Language": "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7",
        },
        HTTPGetRequest,
        "checksum-tw.yahoo.com-accept-language",
    }
    */
}
