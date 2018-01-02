package seth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
)

var ErrNotFound = errors.New("seth: not found")

// represent a pending rpc request
type pending struct {
	notify chan struct{}
	res    interface{}
	err    error
}

func (c *Client) do(method string, params []json.RawMessage, res interface{}) error {
	if c.tport != nil {
		return c.tport.Execute(method, params, res)
	}
	notify := make(chan struct{}, 1)
	c.lock.Lock()
	if c.enc == nil {
		if err := c.reconnect(); err != nil {
			c.lock.Unlock()
			return err
		}
	}
	c.req.Version = "2.0"
	c.req.Method = method
	c.req.Params = params
	c.req.ID = c.id()
	p := &pending{notify: notify, res: res}
	c.pending[c.req.ID] = p
	err := c.enc.Encode(&c.req)
	if err != nil {
		c.abort(err)
	}
	c.lock.Unlock()
	<-p.notify
	return p.err
}

func (c *Client) background(conn io.ReadWriteCloser) {
	dec := json.NewDecoder(conn)
	for {
		c.res = RPCResponse{}
		err := dec.Decode(&c.res)
		if err != nil {
			log.Printf("seth: conn read: %s", err)
			c.lock.Lock()
			// only abort if we haven't already reconnected
			if c.conn == conn {
				c.abort(err)
			}
			c.lock.Unlock()
			return
		}
		c.lock.Lock()
		p := c.pending[c.res.ID]
		if p != nil {
			delete(c.pending, c.res.ID)
		}
		c.lock.Unlock()
		if p == nil {
			log.Printf("spurious response ID %d", c.res.ID)
			continue
		}
		if c.res.Error.Code != 0 || c.res.Error.Message != "" {
			c := c.res.Error
			p.err = &c
		} else if bytes.Equal(c.res.Result, rawnull) {
			p.err = ErrNotFound
		} else {
			p.err = json.Unmarshal(c.res.Result, p.res)
		}
		close(p.notify)
	}
}

func (c *Client) abort(err error) {
	for id, p := range c.pending {
		p.err = err
		close(p.notify)
		delete(c.pending, id)
	}
	c.enc = nil
	c.conn.Close()
	c.conn = nil
}

func (c *Client) reconnect() error {
	if c.enc != nil {
		c.enc = nil
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	conn, err := c.dial()
	if err != nil {
		return err
	}
	c.conn = conn
	c.enc = json.NewEncoder(conn)
	go c.background(conn)
	return nil
}

// Do makes a raw rpc request; it does not interpret the method or param
// strings, and tries to unmarshal the result directly into "res." Use
// another method instead, if you can.
func (c *Client) Do(method string, params []json.RawMessage, res interface{}) error {
	return c.do(method, params, res)
}

// An HTTPTransport is a client transport for making requests over HTTP.
type HTTPTransport struct {
	URL  string
	lock sync.Mutex
	id   int
}

func (t *HTTPTransport) Execute(method string, params []json.RawMessage, res interface{}) error {
	t.lock.Lock()
	req := &RPCRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      t.id + 1,
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.lock.Unlock()
		return err
	}
	t.id++
	t.lock.Unlock()

	tres := new(RPCResponse)
	hres, err := http.Post(t.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	} else if hres.StatusCode != http.StatusOK {
		return errors.New("http error: " + hres.Status)
	} else if err := json.NewDecoder(hres.Body).Decode(tres); err != nil {
		return err
	}

	if tres.Error.Code != 0 || tres.Error.Message != "" {
		e := tres.Error
		return &e
	} else if bytes.Equal(tres.Result, rawnull) {
		return ErrNotFound
	}
	return json.Unmarshal(tres.Result, res)
}
