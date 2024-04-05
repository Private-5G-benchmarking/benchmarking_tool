// profilinglib provides utility functions for setting up processor and memory
// profiling using pprof
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	// Open Stdin
	file := os.Stdin
	// 	//Connect to Influx database and set up the writeAPI client
	clientOptions := influxdb2.DefaultOptions().
	SetPrecision(time.Nanosecond)
	
	client := influxdb2.NewClientWithOptions("http://localhost:8086", "OnjSj1CE5Feqwdb1c7w1SPj2EJVV6yWpHHUe93HkfKyVeBo4TN5BrcfVezKJ6sUk50XPVyvPVH1ljSv4JaypzQ==", clientOptions)
	defer client.Close()
	
	org := "5gbenchmarking"
	bucket := "5gbenchmarking"
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// Create a ticker to print every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Create a scanner to read from Stdin
	scanner := bufio.NewScanner(file)

	// Channel to signal exit
	done := make(chan struct{})

	totalCounter := int64(0)

	// Goroutine to read from Stdin
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				// Check if there's more input
				if !scanner.Scan() {
					close(done)
					return
				}
				input := scanner.Text()
				fields := strings.Split(input, " ")
				fmt.Println(fields)
				// fmt.Println(len(fields))
				fmt.Println(fields[3])
				fmt.Println(fields[4])
				fmt.Println(fields[5])
				// counter, _ := strconv.ParseInt(scanner.Text(), 10, 0)
				totalCounter ++ // = totalCounter + counter
				if (totalCounter % 10 == 0) {
					fmt.Println("Read from Stdin:", totalCounter)
					p := influxdb2.NewPointWithMeasurement("live_counter").AddField("counter", totalCounter)
					writeAPI.WritePoint(context.Background(), p)
				}
				// if err != nil {
					// fmt.Println(err)
				// }


				// Print the data
			}
		}
	}()

	// Loop to print every 2 seconds
	for {
		select {
		case <-ticker.C:
			// Do nothing, let the data reading goroutine handle printing
		case <-done:
			return // Exit if done
		}
	}
}