package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"packetCapturer/samplelib"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

const batchSize = 10000

// Function to extract the specified substring
func extractSubstring(input string) (string, error) {
	// Check if the input string is long enough
	if len(input) < 16 {
		return "", fmt.Errorf("input string is too short")
	}

	// Extract the substring starting after the first 16 characters
	substring := input[28:]
	return substring, nil
}

func WriteInfluxDBPoint(writeAPI api.WriteAPI, srcIP string, dstIP string, txTs time.Time, rxTs time.Time, packetSize int, found_match bool, measurementName string) {

	point := influxdb2.NewPointWithMeasurement(measurementName).
		AddField("src_ip", srcIP).
		AddField("dst_ip", dstIP).
		AddField("tx_ts", txTs).
		AddField("rx_ts", rxTs).
		AddField("psize", packetSize).
		AddField("found_match", found_match).
		SetTime(rxTs)

	writeAPI.WritePoint(point)
}

func removeFromSlice(slidingWindow []map[string]interface{}, indexToRemove int) []map[string]interface{} {
	// Ensure the index is within the valid range
	if indexToRemove < 0 || indexToRemove >= len(slidingWindow) {
		return slidingWindow
	}

	// Use append to create a new slice excluding the map at the specified index
	return append(slidingWindow[:indexToRemove], slidingWindow[indexToRemove+1:]...)
}

func IntToString(list []int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(list), " ", ",", -1), "[]")
}

func getSequenceNr(packet gopacket.Packet) string {
	// Extract UDP layer
	udpLayer := packet.Layer(layers.LayerTypeUDP)

	sequence_number := ""
	if udpLayer != nil {

		udpPacket, _ := udpLayer.(*layers.UDP)
		payload := udpPacket.Payload

		// Check if the payload is GTP encapsulated
		gtpLayer := packet.Layer(layers.LayerTypeGTPv1U)
		if gtpLayer == nil {
			sequence_number = string(payload[0:8])
		} else {
			gtpPacket, _ := gtpLayer.(*layers.GTPv1U)

			// Access the GTP payload
			gtpPayload := gtpPacket.Payload
			testPacket := gopacket.NewPacket(gtpPayload, layers.LayerTypeIPv4, gopacket.Default)
			testUdpLayer := testPacket.Layer(layers.LayerTypeUDP)
			if testUdpLayer != nil {
				testUdpPacket, _ := testUdpLayer.(*layers.UDP)
				testpayload := testUdpPacket.Payload
				sequence_number = string(testpayload[0:8])
			}
			testTcpLayer := testPacket.Layer(layers.LayerTypeTCP)
			if testTcpLayer != nil {
				testTcpPacket, _ := testTcpLayer.(*layers.TCP)
				sequence_number = strconv.FormatUint(uint64(testTcpPacket.Seq), 10)
			}

		}

	}
	return sequence_number
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

func processPacketToInfluxPoint(packet gopacket.Packet) map[string]interface{} {

	parsedPacket := make(map[string]interface{})
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ipPacket, _ := ipLayer.(*layers.IPv4)
	srcIP := ipPacket.SrcIP.String()
	dstIP := ipPacket.DstIP.String()
	sqNr := getSequenceNr(packet)
	timestamp := packet.Metadata().CaptureInfo.Timestamp
	packetSize := packet.Metadata().CaptureInfo.CaptureLength

	parsedPacket["src_ip"] = srcIP
	parsedPacket["dst_ip"] = dstIP
	parsedPacket["packet_ts"] = timestamp
	parsedPacket["sequence_nr"] = sqNr
	parsedPacket["psize"] = packetSize

	return parsedPacket

}
func isPacketMatch(parsedPacket map[string]interface{}, p map[string]interface{}) bool {
	//This is extracted into its own function to make it easier later
	slidingWindowSnr, slidingWindowPayloadExists := p["sequence_nr"].(string)
	if slidingWindowPayloadExists && parsedPacket["sequence_nr"] == slidingWindowSnr {
		return true
	}
	return false
}

func emptySlidingWindow(slidingWindow []map[string]interface{}, writeAPI api.WriteAPI, cdf []float32, dest_measurement string) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		// Type assertions for map values
		srcIP := p["src_ip"].(string)
		dstIP := p["dst_ip"].(string)
		packet_ts := p["packet_ts"].(time.Time)
		psize := p["psize"].(int)
		if samplelib.Sample(cdf) == 1 {
			WriteInfluxDBPoint(writeAPI, srcIP, dstIP, packet_ts, packet_ts, psize, false, dest_measurement)
			localRowCount += 1
		}
	}
	return localRowCount
}

func main() {
	// Setup flag
	var pcap_loc string
	var dest_measurement string
	var sample_prob float64

	flag.StringVar(&pcap_loc, "s", "", "Provide a file path for the capture file (.pcap(ng))")
	flag.StringVar(&dest_measurement, "t", "", "Provide a name for the destination measurement table in Influx")
	flag.Float64Var(&sample_prob, "p", 1.0, "Provide a sample probability for writing a packet to Influx")

	flag.Parse()

	// Setup sampling
	cdf := samplelib.GetBinaryCdf(float32(sample_prob))

	// Start CPU profiling and other performance measurement stuff
	rowCount := 0
	totalNrPackets := 0
	startTime := time.Now()

	f, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// Create a memory profile
	memoryProfile, err := os.Create("memory.pprof")
	if err != nil {
		log.Fatal(err)
	}

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
	// Iterate through each packet in the pcap file
	for packet := range packetSource.Packets() {
		if packet.ErrorLayer() != nil {
			// Handle the error
			// fmt.Println("Error decoding packet:", packet.ErrorLayer().Error())
			continue // Skip to the next packet
		}
		totalNrPackets++

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if !checkIfRelevantPacket(packet) || ipLayer == nil {
			continue
		}

		parsedPacket := processPacketToInfluxPoint(packet)

		matchFound := false

		for index, p := range slidingWindow {

			if isPacketMatch(parsedPacket, p) {
				sample := samplelib.Sample(cdf)

				if sample == 1 {
					tx_ts := p["packet_ts"].(time.Time)
					packetSize := p["psize"].(int)

					if tx_ts.Before(parsedPacket["packet_ts"].(time.Time)) {
						WriteInfluxDBPoint(writeAPI, p["src_ip"].(string), p["dst_ip"].(string), tx_ts, parsedPacket["packet_ts"].(time.Time), packetSize, true, dest_measurement)
					} else {
						WriteInfluxDBPoint(writeAPI, parsedPacket["src_ip"].(string), parsedPacket["dst_ip"].(string), parsedPacket["packet_ts"].(time.Time), tx_ts, parsedPacket["psize"].(int), true, dest_measurement)
					}

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
			// Type assertions for map values
			srcIP := exitingElement["src_ip"].(string)
			dstIP := exitingElement["dst_ip"].(string)
			packet_ts := exitingElement["packet_ts"].(time.Time)
			psize := exitingElement["psize"].(int)

			if samplelib.Sample(cdf) == 1 {
				WriteInfluxDBPoint(writeAPI, srcIP, dstIP, packet_ts, packet_ts, psize, false, dest_measurement)
				rowCount++
			}
			slidingWindow = slidingWindow[1:]
		}
	}

	rowCount += emptySlidingWindow(slidingWindow, writeAPI, cdf, dest_measurement)

	// Record the end time
	endTime := time.Now()

	// Calculate the duration
	duration := endTime.Sub(startTime)
	fmt.Printf("Script took %s to run.\n", duration)
	fmt.Printf("%d rows written to InfluxDB.\n", rowCount)
	fmt.Printf("%d The total number of packets in pcap is \n", totalNrPackets)
}
