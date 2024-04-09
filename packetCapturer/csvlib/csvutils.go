package csvlib

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"time"
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

func NewPacketInfo(Srcip string, Dstip string, Psize int, Enccapsulated_psize int, Rx_ts time.Time, Tx_ts time.Time, Found_match bool) *PacketInfo {
	return &PacketInfo{
		Srcip: Srcip,
		Dstip: Dstip,
		Psize: Psize,
		Encapsulated_psize: Enccapsulated_psize,
		Rx_ts: convertNanosecondsToSeconds(Rx_ts),
		Tx_ts: convertNanosecondsToSeconds(Tx_ts),
		Found_match: Found_match,
	}
}

func (packet PacketInfo) WriteToCsv(writer *csv.Writer, measurementName string) {
	
	row := []string{
		packet.Srcip,
		packet.Dstip,
		fmt.Sprintf("%d", packet.Psize),
		fmt.Sprintf("%d", packet.Encapsulated_psize),
		fmt.Sprintf("%f", packet.Rx_ts),
		fmt.Sprintf("%f", packet.Tx_ts),
		fmt.Sprintf("%t", packet.Found_match),
	}

	if err := writer.Write(row); err != nil {
		log.Fatalln("error writing record to file", err)
	}

}

func convertNanosecondsToSeconds(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano()) / math.Pow10(9)
}