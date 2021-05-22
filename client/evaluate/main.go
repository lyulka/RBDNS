package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ADD_RECORD string = "addRecord"
var QUERY string = "query"

var ADD_RECORD_LOG_MSG_INVOKING string = "%v INFO   : invoking addRecord(%s, %s), no. %v, client %v\n"
var ADD_RECORD_LOG_MSG_SUCCESS string = "%v INFO   : addRecord(%s, %s) success, no. %v, client %v, latency %vms\n"
var ADD_RECORD_LOG_MSG_TIMEOUT string = "%v WARN   : addRecord(%s, %s) timeout, no. %v, client %v\n"
var QUERY_LOG_MSG_INVOKING string = "%v INFO   : invoking query(%s), no. %v, client %v\n"
var QUERY_LOG_MSG_SUCCESS string = "%v INFO   : query(%s) success, no. %v, client %v, latency %vms\n"
var QUERY_LOG_MSG_TIMEOUT string = "%v WARN   : query(%s) timeout, no. %v, client %v\n"

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
	var outDir string
	for {
		counter += 1
		outDir = fmt.Sprintf("results%d/", counter)
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			os.Mkdir(outDir, 0777)
			break
		}
	}

	// Spawn threads
	done := make(chan bool)
	for t := 1; t <= numThreads; t++ {
		fmt.Printf("Spawn thread: %d\n", t)

		go thread(command, t, opsPerThread, outDir, done)
	}

	for t := 1; t <= numThreads; t++ {
		<-done
	}

	// Combine results
	fileInfos, err := ioutil.ReadDir(outDir)
	if err != nil {
		fmt.Println("Failed to combine test results")
		fmt.Println(err)
		return
	}

	combinedFile, err := os.Create(outDir + "combined.log")
	if err != nil {
		fmt.Println("Failed to combine test results")
		fmt.Println(err)
		return
	}
	defer combinedFile.Close()

	for _, fileInfo := range fileInfos {
		fName := fileInfo.Name()
		fContent, err := ioutil.ReadFile(outDir + "/" + fName)
		if err != nil {
			fmt.Println("Failed to combine test results")
			fmt.Println(err)
			return
		}
		combinedFile.Write(fContent)
	}

	fmt.Println("Evaluation completed.")
	fmt.Println("Test results available at: " + outDir)
	fmt.Println("Combined results available at: " + outDir + "combined.log")
}

func thread(command string, threadNum int, opsPerThread int, outDir string, done chan bool) {
	fmt.Printf("Thread start: %d\n", threadNum)

	outputFile, err := os.Create(fmt.Sprintf("%s/thread%d.log", outDir, threadNum))
	if err != nil {
		fmt.Println("Fail to create output file: " + fmt.Sprintf("dirPath/thread%d.log", threadNum))
		fmt.Println(err)
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
				os.Exit(1)
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
				os.Exit(1)
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

	fmt.Printf("Thread end: %d\n", threadNum)
	done <- true
}
