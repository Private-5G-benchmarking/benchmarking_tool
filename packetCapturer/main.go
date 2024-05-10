package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"packetCapturer/packetlib"
	"packetCapturer/profilinglib"
	"packetCapturer/samplelib"
	"packetCapturer/slidingwindowlib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)



func checkIfRelevantPacket(packet gopacket.Packet) bool {
	gtpLayer := packet.Layer(layers.LayerTypeGTPv1U)
	if gtpLayer != nil {
		encapsulatedPacket, _ := gtpLayer.(*layers.GTPv1U)

		payload := gopacket.NewPacket(encapsulatedPacket.Payload, layers.LayerTypeIPv4, gopacket.Default)

		ipLayer := payload.Layer(layers.LayerTypeIPv4)
		ipPacket, _ := ipLayer.(*layers.IPv4)

		//Done to skip gtp encapsulated icmp messages
		if ipPacket.Protocol == 1 {
			return false
		}


//		if ipPacket.SrcIP.String() != "172.30.0.16" && ipPacket.SrcIP.String() != "10.45.0.16" && ipPacket.SrcIP.String() != "10.45.0.17" {
		ipSrc :=ipPacket.SrcIP.String()
		if ipSrc == "10.45.0.42" || ipSrc == "10.45.0.43" || ipSrc == "10.45.0.46" || ipSrc =="10.45.0.37" || ipSrc == "10.45.0.51" || ipSrc=="10.45.0.52" {
			return true
		}
		return false

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
	var output_csv string
	var sample_prob float64
	var l4_protocol string

	flag.StringVar(&pcap_loc, "s", "", "Provide a file path for the capture file (.pcap(ng))")
	flag.StringVar(&output_csv, "c", "", "Provide a name for the output csv file")
	flag.Float64Var(&sample_prob, "p", 1.0, "Provide a sample probability for writing a packet to Influx")
	flag.StringVar(&l4_protocol, "l4", "udp", "Provide a transport layer protocol to get sequence number for matching")

	flag.Parse()

	// Setup sampling
	cdf := samplelib.GetBinaryCdf(float32(sample_prob))

	// Start CPU profiling and other performance measurement stuff
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

	slidingWindowTx := slidingwindowlib.SlidingWindow{Window: []*packetlib.ParsedPacket{}, WindowSize:20000}
	slidingWindowRx := slidingwindowlib.SlidingWindow{Window: []*packetlib.ParsedPacket{}, WindowSize:20000}
	
	// Iterate through each packet in the pcap file
	for packet := range packetSource.Packets() {

		totalNrPackets++

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if !checkIfRelevantPacket(packet) || ipLayer == nil {
			continue
		}

		//Convert the new packet to an instance of the parsedPacket struct
		parsedPacket := packetlib.NewParsedPacket(packet, l4_protocol)
		//Search through the sliding window and handle any potential matches or overflowing window

		if parsedPacket.Psize == 58 {
			foundMatch := slidingWindowRx.SearchSlidingWindow(parsedPacket, cdf, writer)
			if !foundMatch {
				slidingWindowTx.HandleUnmatchedPacket(parsedPacket, cdf, writer)
			}
			} else {
			foundMatch := slidingWindowTx.SearchSlidingWindow(parsedPacket, cdf, writer)
			if !foundMatch {
				slidingWindowRx.HandleUnmatchedPacket(parsedPacket, cdf, writer)
			}
		}

	}

	slidingWindowTx.EmptySlidingWindow(writer, cdf)
	// slidingWindowRx.EmptySlidingWindow(writer, cdf)

	// Record the end time
	endTime := time.Now()

	// Calculate the duration
	duration := endTime.Sub(startTime)
	fmt.Printf("Script took %s to run.\n", duration)
	fmt.Printf("%d The total number of packets in pcap is \n", totalNrPackets)

	durationMilli := duration.Milliseconds()

	profile_csv, profile_err := os.OpenFile("/home/shared/output_files/profiling/matcher.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if profile_err != nil {
		log.Fatal("Could not open CSV file: ", profile_err)
	}
	defer profile_csv.Close()

	profile_writer := csv.NewWriter(profile_csv)
	defer profile_writer.Flush()
	data := []string{strconv.Itoa(totalNrPackets), strconv.FormatInt(durationMilli, 10), strconv.Itoa(slidingWindowTx.WindowSize)}

	write_err := profile_writer.Write(data)
	if write_err != nil {
		log.Fatal("Could not write to csv file: ", write_err)
	}
}

