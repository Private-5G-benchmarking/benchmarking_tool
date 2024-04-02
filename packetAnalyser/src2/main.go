package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"benchmarking/packetAnalyzer/calculatorlib"
	"benchmarking/packetAnalyzer/parselib"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// SortPackets sorts a slice of Packets on either their rx ts or ts tx.
// It does so in place, and in ascending order.
func SortPackets(packets []*parselib.Packet, on_rx bool) {
	var less func(i, j int) bool
	if on_rx {
		less = func(i, j int) bool {
			return packets[i].Rx_ts-packets[j].Rx_ts <= 0
		}
	} else {
		less = func(i, j int) bool {
			return packets[i].Tx_ts-packets[j].Tx_ts <= 0
		}
	}

	sort.Slice(packets, less)
}

func calculatePerPacketKPIsAndWriteToInflux(packets []*parselib.Packet, writeAPI api.WriteAPI, measurementName string) {
	owd, owd_error := calculatorlib.CalculateOneWayDelay(packets)
	iat, iat_error := calculatorlib.CalculateInterArrivalTime(packets)
	ipdv, ipdv_error := calculatorlib.CalculateJitter(packets)

	if owd_error != nil || iat_error != nil || ipdv_error != nil {
		fmt.Println("Error occured! Returning...")
		return
	}

	for index, packet := range packets {
		nsec := int64(packet.Tx_ts)
		nnanosec := int64(math.Mod(packet.Tx_ts, 1) * math.Pow10(9))
		point := influxdb2.NewPointWithMeasurement(measurementName).
			AddField("packet_owd", owd[index]).
			AddField("packet_interarrival_time", iat[index]).
			AddField("packet_jitter", ipdv[index]).
			SetTime(time.Unix(nsec, nnanosec))

		writeAPI.WritePoint(point)
	}
}

func main() {
	p_in := `srcip,dstip,psize,encapsulated_psize,rx_ts,tx_ts
8.8.8.8,8.8.8.9,58,104,2024-03-12 14:20:03.824793711 +0000 UTC,2024-03-12 14:20:03.824624512 +0000 UTC
8.8.8.8,8.8.8.9,56,104,2024-03-12 14:20:03.824796771 +0000 UTC,2024-03-12 14:20:03.833596771 +0000 UTC`

	packets, err := parselib.ParsePcapToPacketSlice(strings.NewReader(p_in))

	if err != nil {
		log.Fatal(err)
	}

	SortPackets(packets, true)

	clientOptions := influxdb2.DefaultOptions().SetBatchSize(10000).SetPrecision(time.Nanosecond).SetUseGZip(true)

	client := influxdb2.NewClientWithOptions("http://localhost:8086", "OnjSj1CE5Feqwdb1c7w1SPj2EJVV6yWpHHUe93HkfKyVeBo4TN5BrcfVezKJ6sUk50XPVyvPVH1ljSv4JaypzQ==", clientOptions)
	defer client.Close()

	org := "5gbenchmarking"
	bucket := "5gbenchmarking"
	writeAPI := client.WriteAPI(org, bucket)

	calculatePerPacketKPIsAndWriteToInflux(packets, writeAPI, "mandag")
}
