package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ADD_RECORD string = "addRecord"
var QUERY string = "query"

var ADD_RECORD_LOG_MSG_INVOKING string = "%v INFO   : invoking addRecord(%s, %s), no. %v, client %v"
var ADD_RECORD_LOG_MSG_SUCCESS string = "%v INFO   : addRecord(%s, %s) success, no. %v, client %v, latency %vms"
var ADD_RECORD_LOG_MSG_TIMEOUT string = "%v WARN   : addRecord(%s, %s) timeout, no. %v, client %v"
var QUERY_LOG_MSG_INVOKING string = "%v INFO   : invoking query(%s), no. %v, client %v"
var QUERY_LOG_MSG_SUCCESS string = "%v INFO   : query(%s) success, no. %v, client %v, latency %vms"
var QUERY_LOG_MSG_TIMEOUT string = "%v WARN   : query(%s) timeout, no. %v, client %v"

func main() {
	if len(os.Args) != 4 {
		fmt.Println("./evaluate <command> <number of threads> <ops per thread>")
		return
	}

	command := os.Args[1]
	if command != ADD_RECORD && command != QUERY {
		fmt.Printf("Available commands: %s, %s\n", ADD_RECORD, QUERY)
		fmt.Printf("You inputted: %s\n", command)
		return
	}

	numThreads, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("number of threads should be an integer")
		return
	}

	opsPerThread, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("ops per thread should be an integer")
		return
	}

	fmt.Println("Running tests:")
	fmt.Println("command: " + command)
	fmt.Printf("numThreads: %d\n", numThreads)
	fmt.Printf("opsPerThread: %d\n", opsPerThread)

	// Make uniquely named directory for evaluation results
	var counter int
	var dirPath string
	for {
		counter += 1
		dirPath = fmt.Sprintf("results%d/", counter)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			os.Mkdir(dirPath, 0755)
			break
		}
	}

	// Spawn threads
	for t := 1; t <= numThreads; t++ {
		go func(threadNum int) {
			outputFile, err := os.Create(fmt.Sprintf("dirPath/thread%d.log", t))
			if err != nil {
				fmt.Println("Fail to create output file: " + fmt.Sprintf("dirPath/thread%d.log", t))
			}
			defer outputFile.Close()

			// Seed random number generator
			rand.Seed(time.Now().UnixNano())

			for i := 0; i <= opsPerThread; i++ {

				if command == ADD_RECORD {
					key := fmt.Sprint(rand.Intn(1000))
					value := fmt.Sprint(rand.Intn(1000))

					startTime := time.Now()
					out, err := exec.Command("rbdns-client", ADD_RECORD, key, value).Output()
					endTime := time.Now()

					// Originally in nanoseconds. Divide by 1 million to get milliseconds
					elapsed := float64(endTime.Sub(startTime)) / float64(1000000)

					if err != nil {
						fmt.Println("Error whilst running command: " + ADD_RECORD)
						return
					}

					// The reason I TrimSpace is to avoid CRLF and LF shenanigans
					// in different platforms and environments.
					if strings.TrimSpace(string(out)) == "Internal server error" {
						outputFile.WriteString(fmt.Sprintf(ADD_RECORD_LOG_MSG_TIMEOUT,
							time.Now(), key, value, i, threadNum))
					} else {
						outputFile.WriteString(fmt.Sprintf(ADD_RECORD_LOG_MSG_SUCCESS,
							time.Now(), key, value, i, threadNum, elapsed))
					}

				} else if command == QUERY {
					key := fmt.Sprint(rand.Intn(1000))

					startTime := time.Now()
					out, err := exec.Command("rbdns-client", QUERY, key).Output()
					endTime := time.Now()
					elapsed := float64(endTime.Sub(startTime)) / float64(1000000)

					if err != nil {
						fmt.Println("Error whilst running command: " + QUERY)
						return
					}

					if strings.TrimSpace(string(out)) == "Internal server error" {
						outputFile.WriteString(fmt.Sprintf(QUERY_LOG_MSG_TIMEOUT,
							time.Now(), key, i, threadNum))
					} else {
						outputFile.WriteString(fmt.Sprintf(QUERY_LOG_MSG_SUCCESS,
							time.Now(), key, i, threadNum, elapsed))
					}

				}
			}
		}(t)
	}
}
