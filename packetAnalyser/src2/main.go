package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"benchmarking/packetAnalyzer/calculatorlib"
	"benchmarking/packetAnalyzer/parselib"
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

func main() {
	p_in := `srcip,dstip,psize,encapsulated_psize,rx_ts,tx_ts
8.8.8.8,8.8.8.9,58,104,2024-03-12 14:20:03.824593711 +0000 UTC,2024-03-12 14:20:03.824624512 +0000 UTC
8.8.8.8,8.8.8.9,56,104,2024-03-12 14:20:03.824596771 +0000 UTC,2024-03-12 14:20:03.833596771 +0000 UTC`

	packets, err := parselib.ParsePcapToPacketSlice(strings.NewReader(p_in))

	if err != nil {
		log.Fatal(err)
	}

	SortPackets(packets, true)

	for index, packet := range packets {
		fmt.Print(index, ": ")
		fmt.Print(*packet, "\n")

	}

	owds, err := calculatorlib.CalculateOneWayDelay(packets)
	if err != nil {
		log.Fatal(err)
	}
	for _, owd := range owds {
		fmt.Println(owd)
	}
}
