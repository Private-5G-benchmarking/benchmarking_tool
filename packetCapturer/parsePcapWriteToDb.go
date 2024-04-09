package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"packetCapturer/csvlib"
	"packetCapturer/influxlib"
	"packetCapturer/profilinglib"
	"packetCapturer/samplelib"
	"packetCapturer/slidingwindowlib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

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

func isPacketMatch(parsedPacket map[string]interface{}, p map[string]interface{}) bool {
	//This is extracted into its own function to make it easier later
	slidingWindowSnr, slidingWindowPayloadExists := p["sequence_nr"].(string)
	if slidingWindowPayloadExists && parsedPacket["sequence_nr"] == slidingWindowSnr {
		return true
	}
	return false
}

func main() {
	// Setup flag
	var pcap_loc string
	var output_csv string
	var sample_prob float64

	flag.StringVar(&pcap_loc, "s", "", "Provide a file path for the capture file (.pcap(ng))")
	flag.StringVar(&output_csv, "c", "", "Provide a name for the output csv file")
	flag.Float64Var(&sample_prob, "p", 1.0, "Provide a sample probability for writing a packet to Influx")

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

	// Open the pcap file
	handle, err := pcap.OpenOffline(pcap_loc)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	//Setup for csv writes
	file, err := os.Create(output_csv)
	if err != nil {
		log.Fatal("could not create CSV file: ", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"Srcip", "Dstip", "Psize", "Encapsulated_psize", "Rx_tx", "Tx_ts", "Found_match"})
	defer writer.Flush()

	// Create a packet source to read packets from the file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// Creating a list of mixed maps
	slidingWindow := make([]map[string]interface{}, 0)
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
			continue
		}

		parsedPacket := influxlib.ProcessPacketToInfluxPoint(packet)

		matchFound := false

		for index, p := range slidingWindow {
			if isPacketMatch(parsedPacket, p) {
				sample := samplelib.Sample(cdf)

				if sample == 1 {
					slidingwindowlib.HandlePacketMatch(writer, parsedPacket, p)
					rowCount++
				}

				slidingWindow = removeFromSlice(slidingWindow, index)
				matchFound = true
				break
			}
		}
		if !matchFound {
			slidingWindow = append(slidingWindow, parsedPacket)
		}

		if len(slidingWindow) >= 2000 {
			exitingElement := slidingWindow[0]
			exitingPacket := csvlib.NewPacketInfo(exitingElement["src_ip"].(string), exitingElement["dst_ip"].(string), exitingElement["psize"].(int),exitingElement["psize"].(int),exitingElement["packet_ts"].(time.Time),exitingElement["packet_ts"].(time.Time), false)

			if samplelib.Sample(cdf) == 1 {
				exitingPacket.WriteToCsv(writer)
				rowCount++
			}
			slidingWindow = slidingWindow[1:]
		}
	}

	rowCount += slidingwindowlib.EmptySlidingWindow(slidingWindow, writer, cdf)

	// Record the end time
	endTime := time.Now()

	// Calculate the duration
	duration := endTime.Sub(startTime)
	fmt.Printf("Script took %s to run.\n", duration)
	fmt.Printf("%d rows written to csv.\n", rowCount)
	fmt.Printf("%d The total number of packets in pcap is \n", totalNrPackets)
}
