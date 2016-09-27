package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"os"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Loader struct {
	num           int
	url           string
	origin        string
	dataFile      string
	substitutions *map[string]interface{}
	sleep         int
	rotate        bool

	Finish        chan string
	conn          *websocket.Conn
}

func NewLoader(num int, url string, origin string, dataFile string, substitutions *map[string]interface{}, sleep int, rotate bool) Loader {
	return Loader{
		num,
		url,
		origin,
		dataFile,
		substitutions,
		sleep,
		rotate,
		make(chan string),
		nil,
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

	//Response handling
	done := make(chan string)
	go func() {
		defer conn.Close()
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				//done <- err.Error()
				return
			}
			if dbg {
				log.Printf("[%d] recv: %s", loader.num, message)
			}
		}
	}()

	log.Printf("[%d] established connection", loader.num)
}

func (loader *Loader) Run() {
	iter := 0
	finished := false

	for !finished {
		log.Printf("[%d] run iteration %d", loader.num, iter)
		loader.send()
		log.Printf("[%d] finished iteration %d", loader.num, iter)

		if (!loader.rotate) {
			finished = true
		}
	}
	log.Printf("[%d] run completed", loader.num)
	loader.Finish <- "ok"
}

func (loader *Loader) Close() {
	loader.conn.Close()
}

func (loader *Loader) send() {
	//Reading the data file line by line, updating JSON and sending it to WS
	file, err := os.Open(loader.dataFile)
	if (err != nil) {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for (scanner.Scan()) {
		req := Request{}
		err := json.Unmarshal(scanner.Bytes(), &req)

		if (err != nil) {
			panic(err)
		}

		req.RenewId()

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

		time.Sleep(time.Duration(loader.sleep) * time.Millisecond)
	}
}
