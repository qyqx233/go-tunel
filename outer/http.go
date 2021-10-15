package outer

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type H struct{}

var h H

func (h H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	url, err := url.Parse(uri)
	if err != nil {
		w.Write([]byte("Error: parse URL failed"))
		return
	}
	switch url.Path {
	case "/xqpl/showtransports":

	case "/xqpl/shutdowncmd":
		values := url.Query()
		idStr := values.Get("id")
		id, _ := strconv.Atoi(idStr)
		transportMng.tl[id].cmdConn.Close()
		h.handleShutdown(w, r)
	case "/register":
		values := url.Query()
		host := values.Get("host")
		port := values.Get("port")
		name := values.Get("name")
		symkey := values.Get("symkey")
		h.handleRegister(w, host, port, name, symkey)
	}
	w.Write([]byte("hello"))
}

func (h H) handleShutdown(w http.ResponseWriter, r *http.Request) {

}

func (h H) handleRegister(w http.ResponseWriter, host, port, name, symkey string) {
	portInt, _ := strconv.Atoi(port)
	transportMng.rwl.Lock()
	// var nameArr, symkeyArr [16]byte
	nameBytes := []byte(name)
	symkeyBytes := []byte(symkey)
	tl = transportMng.add(&transportImpl{IP: host, TargetPort: portInt, Name: nameBytes, SymKey: symkeyBytes})
	transportMng.rwl.Unlock()
	fmt.Println(transportMng.tl)
}

func httpSvr(addr string) {
	server := http.Server{Addr: addr, Handler: H{}}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
