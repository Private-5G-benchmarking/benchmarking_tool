package csvlib

import (
	"encoding/csv"
	"fmt"
	"log"
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