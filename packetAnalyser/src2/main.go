package main

import (
	"flag"
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
func SortPackets(packets []*parselib.PacketInfo, on_rx bool) {
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

func calculatePerPacketKPIsAndWriteToInflux(packets []*parselib.PacketInfo, calculatorMap calculatorlib.PerPacketCalculatorMap, writeAPI api.WriteAPI, measurementName string) {
	valueMap, error := calculatorlib.CalculatePerPacketKPIs(calculatorMap, packets)

	if error != nil {
		log.Fatal(error)
	}

	for index, packet := range packets {
		numSec := int64(packet.Tx_ts)
		numNanosec := int64(math.Mod(packet.Tx_ts, 1) * math.Pow10(9))
		point := influxdb2.NewPointWithMeasurement(measurementName)

		for kpiName, kpi_values := range valueMap {
			value := kpi_values[index]
			if value >= 0 {
				point = point.AddField(kpiName, kpi_values[index])
			}
		}

		point = point.SetTime(time.Unix(numSec, numNanosec))

		writeAPI.WritePoint(point)
	}
}

func calculateAggregateKPIsAndWriteToInflux(packets []*parselib.PacketInfo, calculatorMap calculatorlib.AggregateCalculatorMap, writeAPI api.WriteAPI, measurementName string) {
	valueMap, error := calculatorlib.CalculateAggregateKPIs(calculatorMap, packets)

	if error != nil {
		log.Fatal(error)
	}

	newMap := make(map[int64]map[string]float32)

	for key, innerMap := range valueMap {
		for innerKey, value := range innerMap {
			if _, ok := newMap[innerKey]; !ok {
				newMap[innerKey] = make(map[string]float32)
			}
			newMap[innerKey][key] = value
		}
	}

	for timeSeconds, innerMap := range newMap {
		point := influxdb2.NewPointWithMeasurement(measurementName + "_aggregate")

		for kpiName, value := range innerMap {
			if value >= 0 {
				point = point.AddField(kpiName, value)
			}
		}

		point = point.SetTime(time.Unix(timeSeconds, 0))

		writeAPI.WritePoint(point)
	}
}

func main() {
	var measurementName string

	flag.StringVar(&measurementName, "m", "test", "Provide an Influx measurement name")

	flag.Parse()

	p_in := `srcip,dstip,psize,encapsulated_psize,rx_ts,tx_ts,found_match
	8.8.8.8,8.8.8.9,58,104,2024-03-12 14:20:03.824793711 +0000 UTC,2024-03-12 14:20:03.824624512 +0000 UTC,true
	8.8.8.8,8.8.8.9,56,104,2024-03-12 14:20:03.834796771 +0000 UTC,2024-03-12 14:20:03.833596771 +0000 UTC,true`

	packets, err := parselib.ParsePcapToPacketInfoSlice(strings.NewReader(p_in))

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

	calculatePerPacketKPIsAndWriteToInflux(packets, calculatorlib.GetPerPacketCalculatorMap(), writeAPI, measurementName)
	calculateAggregateKPIsAndWriteToInflux(packets, calculatorlib.GetAggregateCalculatorMap(), writeAPI, measurementName)
	// packets := []*parselib.Packet{
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 0.004, Tx_ts: 0.003},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 0.005, Tx_ts: 0.004},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 0.006, Tx_ts: 0.005},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 0.007, Tx_ts: 0.006},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 1.001, Tx_ts: 1.000},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 1.007, Tx_ts: 1.006},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 1.007, Tx_ts: 1.006},
	// 	{Srcip: "1", Dstip: "2", Psize: 56, Encapsulated_psize: 100, Rx_ts: 1.007, Tx_ts: 1.006},
	// }
}
