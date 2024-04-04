package parselib

import (
	"encoding/csv"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
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

// ConvertToCSVFormat returns a comma-separated string without whitespace to
// be used in a csv-file. The order of the values are
// 1. Srcip 2. Dstip 3. Psize 4. Encapsulated_psize 5. Rx_ts 6. Tx_ts
func (packet Packet) ConvertToCSVFormat() string {
	txSec := int64(packet.Tx_ts)
	txNanosec := int64(math.Mod(packet.Tx_ts, 1) * math.Pow10(9))
	txTime := time.Unix(txSec, txNanosec)
	rxSec := int64(packet.Rx_ts)
	rxNanosec := int64(math.Mod(packet.Rx_ts, 1) * math.Pow10(9))
	rxTime := time.Unix(rxSec, rxNanosec)

	values := []string{packet.Srcip, packet.Dstip, strconv.Itoa(packet.Psize), strconv.Itoa(packet.Encapsulated_psize), rxTime.String(), txTime.String()}

	return strings.Join(values, ",")
}

// parseCSVRecordToPacket converts a record read from a csv-file into a
// Packet-object and returns a pointer to said object. It also accepts
// a timestamp layout to parse the timestamp from the record.
// It assumes the order of the columns to be 1. Srcip 2. Dstip 3. Psize
// 4. Encapsulated_psize 5. Rx_ts 6. Tx_ts
func parseCSVRecordToPacket(record []string, timestampLayout string) (*Packet, error) {
	srcip := record[0]
	dstip := record[1]
	psize := record[2]
	encapsulated_psize := record[3]
	rx_ts := record[4]
	tx_ts := record[5]

	psize_int, err := strconv.Atoi(psize)
	if err != nil {
		return nil, err
	}
	encapsulated_psize_int, err := strconv.Atoi(encapsulated_psize)
	if err != nil {
		return nil, err
	}

	rx_timestamp, err := time.Parse(timestampLayout, rx_ts)
	if err != nil {
		return nil, err
	}
	tx_timestamp, err := time.Parse(timestampLayout, tx_ts)
	if err != nil {
		return nil, err
	}

	packet := Packet{srcip, dstip, psize_int, encapsulated_psize_int, convertNanosecondsToSeconds(rx_timestamp.UnixNano()), convertNanosecondsToSeconds(tx_timestamp.UnixNano())}

	return &packet, nil
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

		packet, parseError := parseCSVRecordToPacket(record, layout)

		if parseError != nil {
			log.Fatal(parseError)
		}

		packets = append(packets, packet)
	}

	return packets, nil
}

func convertNanosecondsToSeconds(nanoseconds int64) float64 {
	return float64(nanoseconds) / math.Pow10(9)
}
