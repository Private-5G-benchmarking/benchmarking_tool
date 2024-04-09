package main

import (
	"encoding/csv"
	"flag"
	"log"
	"math"
	"os"
	"sort"
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

	for kpiName, fn := range calculatorMap {
		values, error := fn(packets)
		if error != nil {
			log.Fatal(error)
		}

		valueMap[kpiName] = values
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
	var csvFileName string

	flag.StringVar(&measurementName, "m", "test", "Provide an Influx measurement name")
	flag.StringVar(&csvFileName, "c", "test", "Provide the filepath to the input csv file")

	flag.Parse()

	f, err := os.Open(csvFileName)
    if err != nil {
        log.Fatal("Unable to read input file due to " , err)
    }
    defer f.Close()
	csvReader := csv.NewReader((f))

	packets, err := parselib.ParsePcapToPacketSlice(csvReader)

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
}
