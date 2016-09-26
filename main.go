package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"math/rand"
	"flag"
	"bufio"
	"os"
	"encoding/json"
	"strings"
	"io/ioutil"
	"fmt"
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

func run(num int, url string, dataFile string, substitutions *map[string]interface{}, sleep int) {
	//Establishing websocket connection
	headers := http.Header{}
	headers.Add("Origin", "http://koroandr.vision.tv.v.netstream.ru")
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if (err != nil) {
		panic(err)
	}
	defer conn.Close()

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
				log.Printf("[%d] recv: %s", num, message)
			}
		}
	}()

	log.Printf("[%d] started", num)

	//Reading the data file line by line, updating JSON and sending it to WS
	file, err := os.Open(dataFile)
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

		for key, value := range *substitutions {
			switch value.(type) {
				case string:
					req.Substitute(fmt.Sprintf("{{%s}}", key), value.(string))
				case []interface{}:
					arr := value.([]interface{})
					if val, ok := arr[num % len(arr)].(string); ok {
						req.Substitute(fmt.Sprintf("{{%s}}", key), val)
					}
			}
		}

		s, err := json.Marshal(req)
		if (err != nil) {
			panic(err)
		}

		if dbg {
			log.Printf("[%d] req: %s", num, s)
		}

		conn.WriteMessage(websocket.TextMessage, []byte(s))

		time.Sleep(time.Duration(sleep) * time.Millisecond)
	}


	log.Printf("[%d] finished", num)
	childDone<-"ok"
}

var dbg bool;
var childDone chan interface {};

func main() {
	//Parsing command-line arguments
	dataFile := flag.String("data", "data.log", "Data file")
	url := flag.String("url", "", "WebSocket endpoint url")
	procCount := flag.Int("proc", 1, "Number of processes to run")
	sleep := flag.Int("sleep", 0, "Sleep time between requests in milliseconds")
	substitutionsFile := flag.String("substitutions", "", "Data file")
	debug := flag.Bool("debug", false, "Show debug output")
	flag.Parse()

	dbg = *debug

	//Initializing substitutions map
	substitutions := make(map[string]interface{})

	if (substitutionsFile != nil && *substitutionsFile != "") {
		substitutionsText, err := ioutil.ReadFile(*substitutionsFile)
		if (err != nil) {
			panic(err)
		}
		err = json.Unmarshal(substitutionsText, &substitutions)
		if (err != nil) {
			panic(err)
		}
	}

	childDone = make(chan interface{})

	for i := 0; i < *procCount; i++ {
		go run(i, *url, *dataFile, &substitutions, *sleep)
	}

	for i := 0; i < *procCount; i++ {
		_ = <-childDone
	}

	log.Println("All done")
}
