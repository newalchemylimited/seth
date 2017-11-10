package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/newalchemylimited/seth/tevm"
)

var addr string

func init() {
	flag.StringVar(&addr, "a", ":8043", "bind address to listen on")
}

type server struct {
	lock sync.Mutex
	c    *tevm.Chain
}

type request struct {
	Version string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      int               `json:"id"`
}

type response struct {
	ID      int             `json:"id"`
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	} `json:"error"`
}

func (s *server) rpc(req *request, res *response) {
	res.ID = req.ID
	res.Version = req.Version
	var resbody json.RawMessage
	s.lock.Lock()
	err := s.c.Execute(req.Method, req.Params, &resbody)
	s.lock.Unlock()
	if err != nil {
		res.Result = nil
		res.Error.Code = -1 // FIXME
		res.Error.Message = err.Error()
		res.Error.Data = nil
		return
	}
	res.Result = resbody
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var jsr request
	err := json.NewDecoder(r.Body).Decode(&jsr)
	if err != nil {
		log.Printf("decode body error: %s", err)
		w.WriteHeader(401)
		return
	}
	var res response
	s.rpc(&jsr, &res)
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Printf("error writing response: %s", err)
		w.WriteHeader(500)
		return
	}
}

func main() {
	flag.Parse()
	c := tevm.NewChain()
	srv := &server{
		c: c,
	}
	acct := c.NewAccount(10)
	log.Println("default account:", acct.String())
	log.Printf("binding to %s...", addr)
	log.Fatal(http.ListenAndServe(addr, srv))
}
