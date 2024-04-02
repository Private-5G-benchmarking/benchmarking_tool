package parselib

import (
	"encoding/csv"
	"io"
	"log"
	"math"
	"strconv"
	"time"
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

		packet := Packet{srcip, dstip, psize_int, encapsulated_psize_int, convertNanosecondsToSeconds(rx_timestamp.UnixNano()), convertNanosecondsToSeconds(tx_timestamp.UnixNano())}
		packets = append(packets, &packet)
	}

	return packets, nil
}

func convertNanosecondsToSeconds(nanoseconds int64) float64 {
	return float64(nanoseconds) / math.Pow10(9)
}
