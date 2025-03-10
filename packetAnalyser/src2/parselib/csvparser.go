package parselib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type PacketInfo struct {
	Srcip              string
	Dstip              string
	Psize              int
	Encapsulated_psize int
	Rx_ts              float64 // Format: seconds with nanosecond precision
	Tx_ts              float64 // Format: seconds with nanosecond precision
	Found_match        bool
}

func (packet PacketInfo) OneWayDelay() (float64, error) {
	if !packet.Found_match {
		return -1, errors.New("attempted to calculate one-way delay on packet with missing match")
	}
	return packet.Rx_ts - packet.Tx_ts, nil
}

// ConvertToCSVFormat returns a comma-separated string without whitespace to
// be used in a csv-file. The order of the values are
// 1. Srcip 2. Dstip 3. Psize 4. Encapsulated_psize 5. Rx_ts 6. Tx_ts
// 7. Found_match
func (packet PacketInfo) ConvertToCSVFormat() string {

	values := []string{
		packet.Srcip,
		packet.Dstip,
		strconv.Itoa(packet.Psize),
		strconv.Itoa(packet.Encapsulated_psize),
		//TODO: could probably use strconv to be more uniform for rx and tx ts.
		fmt.Sprintf("%f", packet.Rx_ts),
		fmt.Sprintf("%f", packet.Tx_ts),
		strconv.FormatBool(packet.Found_match),
	}

	return strings.Join(values, ",")
}

// parseCSVRecordToPacket converts a record read from a csv-file into a
// Packet-object and returns a pointer to said object. It also accepts
// a timestamp layout to parse the timestamp from the record.
// It assumes the order of the columns to be 1. Srcip 2. Dstip 3. Psize
// 4. Encapsulated_psize 5. Rx_ts 6. Tx_ts 7. Found_match
func parseCSVRecordToPacketInfo(record []string) (*PacketInfo, error) {
	srcip := record[0]
	dstip := record[1]
	psize := record[2]
	encapsulated_psize := record[3]
	rx_ts := record[4]
	tx_ts := record[5]
	found_match := record[6]

	if psize == "" {
		return nil, errors.New("psize is undefined")
	}
	psize_int, err := strconv.Atoi(psize)
	if err != nil {
		return nil, err
	}
	if encapsulated_psize == "" {
		return nil, errors.New("encapsulated_psize is undefined")
	}
	encapsulated_psize_int, err := strconv.Atoi(encapsulated_psize)
	if err != nil {
		return nil, err
	}

	rx_ts_ns, err := strconv.ParseFloat(rx_ts, 64)
	if err != nil {
		return nil, err
	}
	if tx_ts == "" {
		return nil, errors.New("tx_ts is undefined")
	}
	tx_ts_ns, err := strconv.ParseFloat(tx_ts, 64)
	if err != nil {
		return nil, err
	}

	if found_match == "" {
		return nil, errors.New("found_match is undefined")
	}
	found_match_bool, err := strconv.ParseBool(found_match)
	if err != nil {
		return nil, err
	}

	packet := PacketInfo{
		srcip,
		dstip,
		psize_int,
		encapsulated_psize_int,
		rx_ts_ns,
		tx_ts_ns,
		found_match_bool,
	}

	return &packet, nil
}

// ParsePcapToPacketSlice accepts an io.Reader object which it expects is
// connected to a .csv-file with information to be formed into the Packet
// struct.
func ParsePcapToPacketSlice(reader *csv.Reader) ([]*PacketInfo, error) {
	packets := make([]*PacketInfo, 0)

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

		packet, parseError := parseCSVRecordToPacketInfo(record)

		if parseError != nil {
			log.Fatal(parseError)
		}

		packets = append(packets, packet)

	}
	return packets, nil
}
