package seth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
)

var ErrNotFound = errors.New("seth: not found")

// represent a pending rpc request
type pending struct {
	notify chan struct{}
	res    *RPCResponse
	err    error
}

// An RPCTrasport is a client transport for making requests over an RPC
// connection.
type RPCTransport struct {
	lock    sync.Mutex
	conn    io.ReadWriteCloser
	enc     *json.Encoder // wraps send side of conn
	pending map[int]*pending
	res     RPCResponse
	dial    func() (io.ReadWriteCloser, error)
}

func (t *RPCTransport) background(conn io.ReadWriteCloser) {
	dec := json.NewDecoder(conn)
	for {
		t.res = RPCResponse{}
		err := dec.Decode(&t.res)
		if err != nil {
			log.Printf("seth: conn read: %s", err)
			t.lock.Lock()
			// only abort if we haven't already reconnected
			if t.conn == conn {
				t.abort(err)
			}
			t.lock.Unlock()
			return
		}
		t.lock.Lock()
		p := t.pending[t.res.ID]
		if p != nil {
			delete(t.pending, t.res.ID)
		}
		t.lock.Unlock()
		if p == nil {
			log.Printf("spurious response ID %d", t.res.ID)
			continue
		}
		if t.res.Error.Code != 0 || t.res.Error.Message != "" {
			c := t.res.Error
			p.err = &c
		} else if bytes.Equal(t.res.Result, rawnull) {
			p.err = ErrNotFound
		} else {
			*p.res = t.res
		}
		close(p.notify)
	}
}

func (t *RPCTransport) abort(err error) {
	for id, p := range t.pending {
		p.err = err
		close(p.notify)
		delete(t.pending, id)
	}
	t.enc = nil
	t.conn.Close()
	t.conn = nil
}

func (t *RPCTransport) reconnect() error {
	if t.enc != nil {
		t.enc = nil
	}
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	conn, err := t.dial()
	if err != nil {
		return err
	}
	t.conn = conn
	t.enc = json.NewEncoder(conn)
	go t.background(conn)
	return nil
}

// Do makes a raw rpc request; it does not interpret the method or param
// strings, and tries to unmarshal the result directly into "result." Use
// another method instead, if you can.
func (c *Client) Do(method string, params []json.RawMessage, result interface{}) error {
	req := &RPCRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      int(atomic.AddUintptr(&c.nextid, 1)),
	}
	res := new(RPCResponse)
	if err := c.tport.Execute(req, res); err != nil {
		return err
	}
	if res.Error.Code != 0 || res.Error.Message != "" {
		e := res.Error
		return &e
	} else if bytes.Equal(res.Result, rawnull) {
		return ErrNotFound
	}

	return json.Unmarshal(res.Result, result)
}

func (t *RPCTransport) Execute(req *RPCRequest, res *RPCResponse) error {
	notify := make(chan struct{}, 1)
	t.lock.Lock()
	if t.enc == nil {
		if err := t.reconnect(); err != nil {
			t.lock.Unlock()
			return err
		}
	}
	p := &pending{notify: notify, res: res}
	t.pending[req.ID] = p
	err := t.enc.Encode(req)
	if err != nil {
		t.abort(err)
	}
	t.lock.Unlock()
	<-p.notify
	return p.err
}

// An HTTPTransport is a client transport for making requests over HTTP.
type HTTPTransport struct {
	URL string
}

// Execute implements Transport.
func (t *HTTPTransport) Execute(req *RPCRequest, res *RPCResponse) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	hres, err := http.Post(t.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer hres.Body.Close()
	if hres.StatusCode != http.StatusOK {
		return errors.New("http error: " + hres.Status)
	}
	return json.NewDecoder(hres.Body).Decode(res)
}

// InfuraTransport is a transport that operates on
// https://api.infura.io/v1/jsonrpc/mainnet
type InfuraTransport struct{}

func infuraPost(req *RPCRequest, res *RPCResponse) error {
	// The infura POST case looks exactly
	// like the regular JSON-RPC API...
	t := HTTPTransport{URL: "https://api.infura.io/v1/jsonrpc/mainnet"}
	return t.Execute(req, res)
}

func infuraGet(req *RPCRequest, res *RPCResponse) error {
	infura := "https://api.infura.io/v1/jsonrpc/mainnet/" + req.Method

	reqbody, err := json.Marshal(req.Params)
	if err != nil {
		return err
	}

	// TODO: not this. This is gross.
	query := make(url.Values)
	query.Add("params", string(reqbody))
	infura += "?" + query.Encode()

	hres, err := http.Get(infura)
	if err != nil {
		return err
	}
	defer hres.Body.Close()
	if hres.StatusCode != http.StatusOK {
		return errors.New("http error: " + hres.Status)
	}
	return json.NewDecoder(hres.Body).Decode(res)
}

func (i InfuraTransport) Execute(req *RPCRequest, res *RPCResponse) error {
	switch req.Method {
	case "eth_sendRawTransaction",
		"eth_estimateGas",
		"eth_submitWork",
		"eth_submitHashrate":
		return infuraPost(req, res)
	default:
		return infuraGet(req, res)
	}
}
