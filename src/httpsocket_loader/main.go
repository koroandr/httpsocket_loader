package main

import (
	"log"
	"flag"
	"encoding/json"
	"io/ioutil"
	"os"
	"bufio"
)

func run(num int, url string, origin string, data *[]Request, substitutions *map[string]interface{}, sleep int, rotate bool) {
	loader := NewLoader(num, url, origin, data, substitutions, sleep, rotate)
	loader.Connect()
	go func() {
		status := <- loader.Finish
		childDone <- status
	}()
	go loader.Run()
}

// Load JSON-RPC requests from data file (they must be placed line by line)
func readRequests(filename string) *[]Request {
	file, err := os.Open(filename)
	if (err != nil) {
		panic(err)
	}

	requests := make([]Request, 0)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for (scanner.Scan()) {
		var req Request

		err := json.Unmarshal(scanner.Bytes(), &req)

		if (err != nil) {
			panic(err)
		}


		requests = append(requests, req)
	}

	return &requests
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

	requests := readRequests(*dataFile)

	//Spawning child processes to replay data.log
	for i := 0; i < *procCount; i++ {
		run(i, *url, *origin, requests, &substitutions, *sleep, *rotate)
	}

	for i := 0; i < *procCount; i++ {
		_ = <-childDone
	}

	log.Println("All done")
}
