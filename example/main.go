package main

import (
    wd "github.com/changyy/go-watchdog"

    "context"
    "sync"

    "os"
    "os/signal"
    "io"
    "log"
    "net/http"
)

type SimpleProxy struct {
    wg *sync.WaitGroup
    srv *http.Server
    address string
}

func (sp *SimpleProxy) ServeHTTP (rw http.ResponseWriter, req *http.Request) {
    log.Println("ServeHTTP:", req.RequestURI, ", req.Method:", req.Method, ", req.URL.Scheme:", req.URL.Scheme)
    if req.URL.Scheme != "http" {
        http.Error(rw, req.URL.Scheme, http.StatusBadRequest)
        return
    }

    client := &http.Client{}

    req.RequestURI = ""
    if resp, err := client.Do(req); err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    } else {
        defer resp.Body.Close()

        for k, v := range(resp.Header) {
            for _, s := range v {
                rw.Header().Add(k , s)
            }
        }
        log.Println("resp.StatusCode:", resp.StatusCode)
        rw.WriteHeader(resp.StatusCode)
        io.Copy(rw, resp.Body)
    }
}

func (sp *SimpleProxy) startProxyServer (address string) *http.Server {
    sp.wg = &sync.WaitGroup{}
    sp.wg.Add(1)
    sp.address = address
    sp.srv = &http.Server{Addr: sp.address, Handler: sp}
    go func() {
        defer sp.wg.Done()
        if err := sp.srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal("http.ListenAndServe error:", err)
        }
    }()
    return sp.srv
}

func (sp *SimpleProxy) stopProxyServer () {
    if err := sp.srv.Shutdown(context.TODO()); err != nil {
        log.Fatal("http.Shutdown error:", err)
    }
    sp.wg.Wait()
}

func main() {
    w := &wd.Watchdog{}
    w.InitDB()

    proxy := &SimpleProxy{}
    proxy.startProxyServer("127.0.0.1:8080")

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    for sig := range c {
        log.Println("Get a signal", sig)
        proxy.stopProxyServer()
        w.CloseDB()
        break
    }
}
