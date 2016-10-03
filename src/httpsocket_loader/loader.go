package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type LoaderOptions struct {
	Num           int
	Url           string
	Origin        string
	Requests      []Request
	Substitutions map[string]interface{}
	Sleep         int
	Rotate        bool
}

type Loader struct {
	LoaderOptions
	Finish          chan string
	conn            *websocket.Conn
	send_timestamps map[string]time.Time
	sumTime         time.Duration
	requestsCount   int
}

func NewLoader(opts *LoaderOptions) *Loader {
	return &Loader{
		LoaderOptions:   *opts,
		Finish:          make(chan string),
		send_timestamps: make(map[string]time.Time),
	}
}

func (loader *Loader) Connect() {
	//Establishing websocket connection
	headers := http.Header{}

	//Set the Origin header, if present
	if loader.Origin != "" {
		headers.Add("Origin", loader.Origin)
	}
	conn, _, err := websocket.DefaultDialer.Dial(loader.Url, headers)
	if err != nil {
		panic(err)
	}
	loader.conn = conn

	go func() {
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if dbg {
				log.Printf("[%d] recv: %s", loader.Num, message)
			}

			var resp RpcResp
			if json.Unmarshal(message, &resp) != nil {
				log.Printf("[%d] strange response: %s", loader.Num, message)
			} else {
				loader.recieve(resp.Id)
			}
		}
	}()

	log.Printf("[%d] established connection", loader.Num)
}

func (loader *Loader) Run() {
	iter := 0

	for {
		//Summary time for all requests
		loader.sumTime = 0
		loader.requestsCount = 0

		loader.send()

		if loader.requestsCount > 0 {
			log.Printf("[%d] - iter %d - average time: %d ms", loader.Num, iter, loader.sumTime.Nanoseconds()/int64(loader.requestsCount)/1000000)
		} else {
			log.Printf("[%d] - iter %d - no successful requests", loader.Num, iter)
		}

		if !loader.Rotate {
			break
		}

		iter += 1
	}
	log.Printf("[%d] run completed", loader.Num)
	loader.Finish <- "ok"

	loader.conn.Close()
}

func (loader *Loader) send() {
	//Updating JSON with process-specific substitutions and sending it to WS
	for _, req := range loader.Requests {
		//Setting new id to prevent conflicts between different loaders
		req.RenewId()

		//Substituting some data (mustache-style)
		for key, value := range loader.Substitutions {
			switch value := value.(type) {
			case string:
				req.Substitute(fmt.Sprintf("{{%s}}", key), value)
			case []interface{}:
				if val, ok := value[loader.Num%len(value)].(string); ok {
					req.Substitute(fmt.Sprintf("{{%s}}", key), val)
				}
			}
		}

		s, err := json.Marshal(req)
		dieOnError(err)

		if dbg {
			log.Printf("[%d] req: %s", loader.Num, s)
		}

		err = loader.conn.WriteMessage(websocket.TextMessage, []byte(s))

		if err != nil {
			log.Printf("[%d] Got error while sending a message", loader.Num)
			log.Println(err.Error())
		} else {
			loader.send_timestamps[req.Id] = time.Now()
		}

		time.Sleep(time.Duration(loader.Sleep) * time.Millisecond)
	}
}

func (loader *Loader) recieve(id string) {
	loader.sumTime += time.Since(loader.send_timestamps[id])
	loader.requestsCount += 1
}

type RpcResp struct {
	Id string `json:"id"`
}
