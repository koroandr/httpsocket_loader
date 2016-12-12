package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func run(opts *LoaderOptions) {
	loader := NewLoader(opts)
	loader.Connect()
	go loader.Run()
}

// Load JSON-RPC requests from data file (they must be placed line by line)
func readRequests(filename string) []Request {
	file, err := os.Open(filename)
	defer file.Close()

	dieOnError(err)

	requests := make([]Request, 0)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var req Request

		err := json.Unmarshal(scanner.Bytes(), &req)

		dieOnError(err)

		requests = append(requests, req)
	}

	return requests
}

var dbg bool
var total_cnt int64
var total_cnt_without_upstream int64
var total_lock sync.Mutex

func main() {
	//Parsing command-line arguments
	dataFile := flag.String("data", "data.log", "Data file")
	url := flag.String("url", "", "WebSocket endpoint url")
	origin := flag.String("origin", "", "Origin header")
	procCount := flag.Int("proc", 1, "Number of processes to run")
	sleep := flag.Int("sleep", 0, "Sleep time between requests in milliseconds (cannot be used with rps)")
	rps := flag.Int("rps", 0, "requests per second for each process (cannot be used with sleep)")
	substitutionsFile := flag.String("substitutions", "", "Data file")
	debug := flag.Bool("debug", false, "Show debug output")
	rotate := flag.Bool("rotate", false, "cycle logs")
	randomizeStart := flag.Bool("randomize-start", false, "Randomize start delay (between 0 and sleep)")
	flag.Parse()

	dbg = *debug

	if *sleep != 0 && *rps != 0 {
		fmt.Println("Cannot use rps and sleep flags simultaneously")
		return
	}

	if *sleep == 0 && *rps == 0 {
		fmt.Println("You must specify either rps or sleep flag")
		return
	}

	if *sleep == 0 {
		*sleep = 1000 / *rps
	}

	rand.Seed(time.Now().UnixNano())

	//Initializing substitutions map
	substitutions := make(map[string]interface{})

	if substitutionsFile != nil && *substitutionsFile != "" {
		substitutionsText, err := ioutil.ReadFile(*substitutionsFile)
		dieOnError(err)

		err = json.Unmarshal(substitutionsText, &substitutions)
		dieOnError(err)
	}

	requests := readRequests(*dataFile)

	wg := sync.WaitGroup{}

	//Spawning child processes to replay data.log
	for i := 0; i < *procCount; i++ {
		run(&LoaderOptions{
			Num:            i,
			Url:            *url,
			Origin:         *origin,
			Requests:       requests,
			Substitutions:  substitutions,
			Sleep:          *sleep,
			Rotate:         *rotate,
			WaitGroup:      &wg,
			RandomizeStart: *randomizeStart,
		})
	}

	wg.Wait()

	fmt.Printf("avg time %.1f,\tavg proxy time %.1f\n", float64(total_cnt)/float64(*procCount), float64(total_cnt_without_upstream)/float64(*procCount))

	log.Println("All done")
}
