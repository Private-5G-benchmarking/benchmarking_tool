// This package is used to send "live" updates to influx
// and is used by a grafana dashboard for refreshes
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func main() {
	// 	//Connect to Influx database and set up the writeAPI client
	clientOptions := influxdb2.DefaultOptions().
	SetPrecision(time.Nanosecond)	
	client := influxdb2.NewClientWithOptions("http://localhost:8086", "OnjSj1CE5Feqwdb1c7w1SPj2EJVV6yWpHHUe93HkfKyVeBo4TN5BrcfVezKJ6sUk50XPVyvPVH1ljSv4JaypzQ==", clientOptions)
	defer client.Close()
	org := "5gbenchmarking"
	bucket := "5gbenchmarking"
	writeAPI := client.WriteAPIBlocking(org, bucket)
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	// Open Stdin and create scanner to read from Stdin
	file := os.Stdin

	scanner := bufio.NewScanner(file)

	done := make(chan struct{})

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

				// Take the Tshark output and extract the tshark counter
				inputString := strings.TrimSpace(scanner.Text())
				fields := strings.Split(inputString, " ")
				tsharkCounter, err := strconv.ParseInt(fields[0], 10, 0)
				
				if err != nil {
					fmt.Println(err)
				}
				//Write a count every 10 000 packets 
				if (tsharkCounter % 10000 == 0) {
					p := influxdb2.NewPointWithMeasurement("live_counter").AddField("counter", tsharkCounter)
					writeAPI.WritePoint(context.Background(), p)
				}
			}
		}
	}()

	// Loop to print every 2 seconds
	for {
		select {
		case <-ticker.C:
		case <-done:
			return
		}
	}
}