package parselib

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
)

type Packet struct {
	Srcip              string
	Dstip              string
	Psize              int
	Encapsulated_psize int
	Rx_ts              float64 // Format: seconds with nanosecond precision
	Tx_ts              float64 // Format: seconds with nanosecond precision
}

// ParsePcapToPacketSlice accepts an io.Reader object which it expects is
// connected to a .csv-file with information to be formed into the Packet
// struct.
func ParsePcapToPacketSlice(r *csv.Reader) ([]*Packet, error) {
	packets := make([]*Packet, 0)

	i := 0

	for {
		record, err := r.Read()

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
			return packets, err
		}
		encapsulated_psize_int, err := strconv.Atoi(encapsulated_psize)
		if err != nil {
			return packets, err
		}

		rx_ts_ns, err := strconv.ParseFloat(rx_ts, 64)
		if err != nil {
			return packets, err
		}

		tx_ts_ns, err := strconv.ParseFloat(tx_ts, 64)
		if err != nil {
			return packets, err
		}

		packet := Packet{srcip, dstip, psize_int, encapsulated_psize_int, rx_ts_ns, tx_ts_ns}
		packets = append(packets, &packet)
	}

	return packets, nil
}

