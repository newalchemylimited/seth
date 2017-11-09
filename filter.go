package seth

import (
	"encoding/json"
	"log"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"
)

// Filter represents a Log filter
type Filter struct {
	id  int64
	c   *Client
	out chan *Log

	lock   sync.Mutex // guards below
	exit   chan struct{}
	err    error
	closed bool
	poll   bool // continue polling after first fetch
}

// Out returns the channel of output logs.
// The output channel will be closed when the
// filter is closed, or when the filter encounters
// an error, in which case (*Filter).Err() will be non-nil.
func (f *Filter) Out() <-chan *Log { return f.out }

func (f *Filter) seterr(err error) {
	f.lock.Lock()
	f.err = err
	f.lock.Unlock()
}

// Err returns an error if the filter
// has encountered an error while fetching logs.
func (f *Filter) Err() error {
	f.lock.Lock()
	err := f.err
	f.lock.Unlock()
	return err
}

// Close will close the filter. Closing
// the filter will cause the output channel
// to be closed. Close is safe to call from
// any goroutine.
func (f *Filter) Close() {
	f.lock.Lock()
	if !f.closed {
		close(f.exit)
		f.closed = true
		f.c.deleteFilter(f.id)
	}
	f.lock.Unlock()
}

type newFilterReq struct {
	FromBlock json.RawMessage `json:"fromBlock,omitempty"`
	ToBlock   json.RawMessage `json:"toBlock,omitempty"`
	Address   *Address        `json:"address,omitempty"`
	Topics    []*Hash         `json:"topics,omitempty"`
}

func frecv(f *Filter) {
	logs, err := f.c.getLogs(f.id)
	if err != nil {
		f.seterr(err)
		goto done
	}
	for i := range logs {
		f.out <- &logs[i]
	}
	if !f.poll {
		goto done
	}
	for {
		ticker := time.NewTicker(5 * time.Second)
		select {
		case <-f.exit:
			ticker.Stop()
			goto done
		case <-ticker.C:
			logs, err := f.c.getUpdates(f.id)
			if err != nil {
				f.seterr(err)
				ticker.Stop()
				goto done
			}
			for i := range logs {
				f.out <- &logs[i]
			}
		}
	}
done:
	close(f.out)
}

func (c *Client) getLogs(id int64) ([]Log, error) {
	var o []Log
	p := []json.RawMessage{itox(id)}
	err := c.do("eth_getFilterLogs", p, &o)
	return o, err
}

func (c *Client) getUpdates(id int64) ([]Log, error) {
	var o []Log
	p := []json.RawMessage{itox(id)}
	err := c.do("eth_getFilterChanges", p, &o)
	return o, err
}

func (c *Client) deleteFilter(id int64) {
	var out bool
	err := c.do("eth_uninstallFilter", []json.RawMessage{itox(id)}, &out)
	if err != nil || !out {
		log.Printf("uninstallFilter: %s %v", err)
	}
}

func itox(i int64) json.RawMessage {
	o := make([]byte, 3, 64)
	o[0] = '"'
	o[1] = '0'
	o[2] = 'x'
	o = strconv.AppendInt(o, i, 16)
	return append(o, '"')
}

var rawearliest = json.RawMessage(`"earliest"`)

// FilterTopics creates a log filter that matches the given topics. (Topics are order-dependent.)
// If 'addr' is non-nil, only logs generated from that address are yielded by the filter.
// If 'start' and 'end' are non-negative, then they specify the range of blocks in which to
// search. Otherwise, the filter starts at the latest block.
func (c *Client) FilterTopics(topics []*Hash, addr *Address, start, end int64) (*Filter, error) {
	_ = math.MaxInt64
	req := &newFilterReq{
		Address: addr,
		Topics:  topics,
	}
	poll := false
	if start < 0 {
		poll = true
		req.FromBlock = rawlatest
	} else {
		req.FromBlock = itox(start)
	}
	if end < 0 {
		req.ToBlock = rawlatest
	} else {
		req.ToBlock = itox(end)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var out Int
	err = c.do("eth_newFilter", []json.RawMessage{buf}, &out)
	if err != nil {
		return nil, err
	}
	id := (*big.Int)(&out).Int64()
	f := &Filter{c: c, out: make(chan *Log, 20), exit: make(chan struct{}, 1), id: id, poll: poll}
	go frecv(f)
	return f, nil
}

// TokenTransfers returns a filter that searches for token transfers
// matching the given arguments. If any of the argments are nil, the
// filter matches that argument as a wildcard. In other words,
// if from, to, and tok are all nil, this filter finds all token
// transfers in the given block range.
func (c *Client) TokenTransfers(from *Address, to *Address, tok *Address, start, end int64) (*Filter, error) {
	var hashes [3]Hash
	var arg0 [3]*Hash
	hashes[0] = ERC20Transfer
	arg0[0] = &hashes[0]
	if from != nil {
		copy(hashes[1][12:], from[:])
		arg0[1] = &hashes[1]
	}
	if to != nil {
		copy(hashes[2][12:], to[:])
		arg0[2] = &hashes[2]
	}
	return c.FilterTopics(arg0[:], tok, start, end)
}
