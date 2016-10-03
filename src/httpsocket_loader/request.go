package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (req *Request) RenewId() {
	req.Id = fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
}

func (req *Request) Substitute(from string, to string) {
	req.Method = strings.Replace(req.Method, from, to, -1)
	req.Params = []byte(strings.Replace(string(req.Params), from, to, -1))
}
