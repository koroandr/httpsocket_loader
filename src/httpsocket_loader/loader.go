package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Loader struct {
	num             int
	url             string
	origin          string
	data            *[]Request
	substitutions   *map[string]interface{}
	sleep           int
	rotate          bool

	Finish          chan string
	conn            *websocket.Conn
	send_timestamps map[string]int64
	sumTime         int64
	requestsCount   int
}

func NewLoader(num int, url string, origin string, data *[]Request, substitutions *map[string]interface{}, sleep int, rotate bool) *Loader {
	return &Loader{
		num: num,
		url: url,
		origin: origin,
		data: data,
		substitutions: substitutions,
		sleep: sleep,
		rotate: rotate,
		Finish: make(chan string),
		send_timestamps: make(map[string]int64),
	}
}

func (loader *Loader) Connect() {
	//Establishing websocket connection
	headers := http.Header{}

	//Set the Origin header, if present
	if (loader.origin != "") {
		headers.Add("Origin", loader.origin)
	}
	conn, _, err := websocket.DefaultDialer.Dial(loader.url, headers)
	if (err != nil) {
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
				log.Printf("[%d] recv: %s", loader.num, message)
			}

			var resp RpcResp
			if (json.Unmarshal(message, &resp) != nil) {
				log.Printf("[%d] strange response: %s", loader.num, message)
			} else {
				loader.recieve(resp.Id)
			}
		}
	}()

	log.Printf("[%d] established connection", loader.num)
}

func (loader *Loader) Run() {
	iter := 0

	for {
		//Summary time for all requests
		loader.sumTime = 0
		loader.requestsCount = 0

		loader.send()

		if (loader.requestsCount > 0) {
			log.Printf("[%d] - iter %d - average time: %d ms", loader.num, iter, loader.sumTime / int64(loader.requestsCount) / 1000000)
		} else {
			log.Printf("[%d] - iter %d - no successful requests", loader.num, iter)
		}

		if (!loader.rotate) {
			break
		}

		iter += 1
	}
	log.Printf("[%d] run completed", loader.num)
	loader.Finish <- "ok"

	loader.conn.Close()
}

func (loader *Loader) send() {
	//Updating JSON with process-specific substitutions and sending it to WS
	for _, req := range *loader.data {
		//Setting new id to prevent conflicts between different loaders
		req.RenewId()

		//Substituting some data (mustache-style)
		for key, value := range *loader.substitutions {
			switch value.(type) {
			case string:
				req.Substitute(fmt.Sprintf("{{%s}}", key), value.(string))
			case []interface{}:
				arr := value.([]interface{})
				if val, ok := arr[loader.num % len(arr)].(string); ok {
					req.Substitute(fmt.Sprintf("{{%s}}", key), val)
				}
			}
		}

		s, err := json.Marshal(req)
		if (err != nil) {
			panic(err)
		}

		if dbg {
			log.Printf("[%d] req: %s", loader.num, s)
		}

		loader.conn.WriteMessage(websocket.TextMessage, []byte(s))
		loader.send_timestamps[req.Id] = time.Now().UnixNano()

		time.Sleep(time.Duration(loader.sleep) * time.Millisecond)
	}
}

func (loader *Loader) recieve(id string) {
	loader.sumTime += (time.Now().UnixNano() - loader.send_timestamps[id])
	loader.requestsCount += 1
}

type RpcResp struct {
	Id string `json:"id"`
}
