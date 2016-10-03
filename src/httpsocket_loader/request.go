package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Request struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      string `json:"id"`
	Method  string `json:"method"`
	Params  string `json:"params"`
}

func (req *Request) RenewId() {
	req.Id = fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
}

func (req *Request) Substitute(from string, to string) {
	req.Method = strings.Replace(req.Method, from, to, -1)
	req.Params = strings.Replace(req.Params, from, to, -1)
}
