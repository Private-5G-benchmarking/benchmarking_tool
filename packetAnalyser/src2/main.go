package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Packet struct {
	srcip              string
	dstip              string
	psize              int
	encapsulated_psize int
	rx_ts              int64
	tx_ts              int64
}

// ParsePcapToPacketSlice accepts an io.Reader object which it expects is
// connected to a .csv-file with information to be formed into the Packet
// struct.
func ParsePcapToPacketSlice(r io.Reader) ([]*Packet, error) {
	reader := csv.NewReader(r)
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	packets := make([]*Packet, 0)

	i := 0

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Error while reading line")
		}
		if i == 0 {
			i += 1
			continue
		}

		srcip := record[0]
		dstip := record[1]
		psize := record[2]
		encapsulated_psize := record[3]
		rx_ts := record[4]
		tx_ts := record[5]

		psize_int, err := strconv.Atoi(psize)
		if err != nil {
			// log.Fatal("Error parsing variable psize")
			return packets, err
		}
		encapsulated_psize_int, err := strconv.Atoi(encapsulated_psize)
		if err != nil {
			// log.Fatal("Error parsing variable encapsulated_psize")
			return packets, err
		}

		rx_timestamp, err := time.Parse(layout, rx_ts)
		if err != nil {
			// log.Fatal("Error parsing rx timestamp")
			return packets, err
		}
		tx_timestamp, err := time.Parse(layout, tx_ts)
		if err != nil {
			// log.Fatal("Error parsing tx timestamp")
			return packets, err
		}

		packet := Packet{srcip, dstip, psize_int, encapsulated_psize_int, rx_timestamp.UnixNano(), tx_timestamp.UnixNano()}
		packets = append(packets, &packet)
	}

	return packets, nil
}

// SortPackets sorts a slice of Packets on either their rx ts or ts tx.
// It does so in place, and in ascending order.
func SortPackets(packets []*Packet, on_rx bool) {
	var less func(i, j int) bool
	if on_rx {
		less = func(i, j int) bool {
			return packets[i].rx_ts-packets[j].rx_ts <= 0
		}
	} else {
		less = func(i, j int) bool {
			return packets[i].tx_ts-packets[j].tx_ts <= 0
		}
	}

	sort.Slice(packets, less)
}

func main() {
	p_in := `srcip,dstip,psize,encapsulated_psize,rx_ts,tx_ts
8.8.8.8,8.8.8.9,58,104,2024-03-12 14:20:03.824593711 +0000 UTC,2024-03-12 14:20:03.824624512 +0000 UTC
8.8.8.8,8.8.8.9,56,104,2024-03-12 14:20:03.824596771 +0000 UTC,2024-03-12 14:20:03.833596771 +0000 UTC`

	packets, err := ParsePcapToPacketSlice(strings.NewReader(p_in))

	if err != nil {
		log.Fatal(err)
	}

	for index, packet := range packets {
		fmt.Print(index, ": ")
		fmt.Print(*packet, "\n")
	}

	SortPackets(packets, true)

	for index, packet := range packets {
		fmt.Print(index, ": ")
		fmt.Print(*packet, "\n")

	}
}
