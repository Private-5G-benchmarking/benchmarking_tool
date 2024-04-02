package main

import (
	"flag"
	"fmt"
	"log"
	"runtime/pprof"
	"time"

	"packetCapturer/influxlib"
	"packetCapturer/matchlib"
	"packetCapturer/profilinglib"
	"packetCapturer/samplelib"
	"packetCapturer/slidingwindowlib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const batchSize = 10000

func removeFromSlice(slidingWindow []map[string]interface{}, indexToRemove int) []map[string]interface{} {
	// Ensure the index is within the valid range
	if indexToRemove < 0 || indexToRemove >= len(slidingWindow) {
		return slidingWindow
	}

	// Use append to create a new slice excluding the map at the specified index
	return append(slidingWindow[:indexToRemove], slidingWindow[indexToRemove+1:]...)
}

func checkIfRelevantPacket(packet gopacket.Packet) bool {
	gtpLayer := packet.Layer(layers.LayerTypeGTPv1U)
	if gtpLayer != nil {
		encapsulatedPacket, _ := gtpLayer.(*layers.GTPv1U)

		payload := gopacket.NewPacket(encapsulatedPacket.Payload, layers.LayerTypeIPv4, gopacket.Default)

		ipLayer := payload.Layer(layers.LayerTypeIPv4)
		ipPacket, _ := ipLayer.(*layers.IPv4)

		if ipPacket.SrcIP.String() != "172.30.0.16" {
			return false
		}
		return true
	} else {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		ipPacket, _ := ipLayer.(*layers.IPv4)
		if ipLayer != nil {
			srcIP := ipPacket.SrcIP.String()
			if srcIP == "192.168.2.100" || srcIP == "192.168.2.111" || srcIP == "172.30.1.8" {
				return true
			}
		}
	}
	return false
}

func main() {
	// Setup flag
	var pcap_loc string
	var dest_measurement string
	var sample_prob float64
	var traffic_type string

	flag.StringVar(&pcap_loc, "s", "", "Provide a file path for the capture file (.pcap(ng))")
	flag.StringVar(&dest_measurement, "t", "", "Provide a name for the destination measurement table in Influx")
	flag.Float64Var(&sample_prob, "p", 1.0, "Provide a sample probability for writing a packet to Influx")
	flag.StringVar(&traffic_type, "traf", "udp", "Provide a transport layer protocol to get sequence number for matching")

	flag.Parse()

	// Setup sampling
	cdf := samplelib.GetBinaryCdf(float32(sample_prob))

	// Start CPU profiling and other performance measurement stuff
	rowCount := 0
	totalNrPackets := 0
	startTime := time.Now()

	cpuProfile := profilinglib.CreateCPUProfiler()
	pprof.StartCPUProfile(cpuProfile)
	defer pprof.StopCPUProfile()

	// Create a memory profile
	memoryProfile := profilinglib.CreateMemoryProfiler()

	// Start profiling
	pprof.WriteHeapProfile(memoryProfile)
	defer memoryProfile.Close()

	//Connect to Influx database and set up the writeAPI client
	clientOptions := influxdb2.DefaultOptions().
		SetBatchSize(batchSize).
		SetPrecision(time.Nanosecond).
		SetUseGZip(true)

	client := influxdb2.NewClientWithOptions("http://localhost:8086", "OnjSj1CE5Feqwdb1c7w1SPj2EJVV6yWpHHUe93HkfKyVeBo4TN5BrcfVezKJ6sUk50XPVyvPVH1ljSv4JaypzQ==", clientOptions)
	defer client.Close()

	org := "5gbenchmarking"
	bucket := "5gbenchmarking"
	writeAPI := client.WriteAPI(org, bucket)

	// Open the pcap file
	handle, err := pcap.OpenOffline(pcap_loc)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Create a packet source to read packets from the file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// Creating a list of mixed maps
	slidingWindow := make([]map[string]interface{}, 0)
	irrelevantNrPackets := 0
	nrOfMatchedPackets := 0

	// Iterate through each packet in the pcap file
	for packet := range packetSource.Packets() {
		if packet.ErrorLayer() != nil {
			// Handle the error
			fmt.Println("Error decoding packet:", packet.ErrorLayer().Error())
			continue // Skip to the next packet
		}
		totalNrPackets++

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if !checkIfRelevantPacket(packet) || ipLayer == nil {
			irrelevantNrPackets ++
			continue
		}

		parsedPacket := influxlib.ProcessPacketToInfluxPoint(packet, traffic_type)

		matchFound := false

		for index, p := range slidingWindow {
			if matchlib.IsPacketMatchSequenceNr(parsedPacket, p) {
				sample := samplelib.Sample(cdf)

				if sample == 1 {
					slidingwindowlib.HandlePacketMatch(influxlib.WriteInfluxDBPoint, writeAPI, parsedPacket, p, dest_measurement)
					rowCount++
				}

				slidingWindow = removeFromSlice(slidingWindow, index)
				matchFound = true
				nrOfMatchedPackets ++
				break
			}
		}
		if !matchFound {
			slidingWindow = append(slidingWindow, parsedPacket)
		}

		if len(slidingWindow) >= 2000 {
			exitingElement := slidingWindow[0]
			// Type assertions for map values
			srcIP := exitingElement["src_ip"].(string)
			dstIP := exitingElement["dst_ip"].(string)
			packet_ts := exitingElement["packet_ts"].(time.Time)
			psize := exitingElement["psize"].(int)

			if samplelib.Sample(cdf) == 1 {
				influxlib.WriteInfluxDBPoint(writeAPI, srcIP, dstIP, packet_ts, packet_ts, psize, false, dest_measurement)
				rowCount++
			}
			slidingWindow = slidingWindow[1:]
		}
	}

	rowCount += slidingwindowlib.EmptySlidingWindow(slidingWindow, writeAPI, cdf, dest_measurement)

	// Record the end time
	endTime := time.Now()

	// Calculate the duration
	duration := endTime.Sub(startTime)
	fmt.Printf("Script took %s to run.\n", duration)
	fmt.Printf("%d rows written to InfluxDB.\n", rowCount)
	fmt.Printf("%d The total number of packets in pcap is \n", totalNrPackets)
	fmt.Printf("%d The irrelevant number of packets in pcap is \n", irrelevantNrPackets)
	fmt.Printf("%d The matched number of packets in pcap is \n", nrOfMatchedPackets)
}
