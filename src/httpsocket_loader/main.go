package main

import (
	"log"
	"time"
	"math/rand"
	"flag"
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

func run(num int, url string, origin string, dataFile string, substitutions *map[string]interface{}, sleep int, rotate bool) {
	loader := NewLoader(num, url, origin, dataFile, substitutions, sleep, rotate)
	loader.Connect()
	go func() {
		status := <- loader.Finish
		childDone <- status
	}()
	go loader.Run()
}

var dbg bool;
var childDone chan string;

func main() {
	//Parsing command-line arguments
	dataFile := flag.String("data", "data.log", "Data file")
	url := flag.String("url", "", "WebSocket endpoint url")
	origin := flag.String("origin", "", "Origin header")
	procCount := flag.Int("proc", 1, "Number of processes to run")
	sleep := flag.Int("sleep", 0, "Sleep time between requests in milliseconds")
	substitutionsFile := flag.String("substitutions", "", "Data file")
	debug := flag.Bool("debug", false, "Show debug output")
	rotate := flag.Bool("rotate", false, "cycle logs")
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

	childDone = make(chan string)

	//Spawning child processes to replay data.log
	for i := 0; i < *procCount; i++ {
		run(i, *url, *origin, *dataFile, &substitutions, *sleep, *rotate)
	}

	for i := 0; i < *procCount; i++ {
		_ = <-childDone
	}

	log.Println("All done")
}
