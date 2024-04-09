package calculatorlib

import (
	"fmt"
	"math"

	"benchmarking/packetAnalyzer/parselib"
)

// CalculateOneWayDelay accepts an array of Packets and calculates the difference
// between their receive timestamp (Rx_ts) and their transmit timestamp (Tx_ts).
// It returns the difference in seconds.
func CalculateOneWayDelay(packets []*parselib.Packet) ([]float64, error) {
	one_way_delays := make([]float64, len(packets))

	for index, packet := range packets {
		one_way_delay := packet.Rx_ts - packet.Tx_ts
		one_way_delays[index] = one_way_delay
	}

	return one_way_delays, nil
}

// CalculateInterArrivalTime accepts and array of Packets and calculates the
// difference in transmit timestamps (Tx_ts) between successive packets.
// It returns the differences in seconds.
func CalculateInterArrivalTime(packets []*parselib.Packet) ([]float64, error) {
	inter_arrival_times := make([]float64, len(packets))

	for i := 1; i < len(packets); i++ {
		inter_arrival_time := packets[i].Tx_ts - packets[i-1].Tx_ts
		inter_arrival_times[i] = inter_arrival_time
	}

	return inter_arrival_times, nil
}

// CalculateJitter (CalculateIPDV) accepts an array of Packets and calculates
// the IPDV for each packet according to
// RFC 3393. It returns the IPDVs in seconds.
func CalculateJitter(packets []*parselib.Packet) ([]float64, error) {
	fmt.Println(len(packets))
	jitters := make([]float64, len(packets))
	one_way_delays, err := CalculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	d := func(i, j int) float64 {
		return one_way_delays[j] - one_way_delays[i]
	}

	jitters[0] = 0

	for i := 1; i < len(packets); i++ {
		jitter := jitters[i-1] + (math.Abs(d(i-1, i)-jitters[i-1]) / 16)
		jitters[i] = jitter
	}

	return jitters, nil
}

type CalculatorMap map[string]func([]*parselib.Packet) ([]float64, error)

func GetCalculatorMap() CalculatorMap {
	m := make(CalculatorMap)

	m["packet_owd"] = CalculateOneWayDelay
	m["packet_interarrival_time"] = CalculateInterArrivalTime
	m["packet_jitter"] = CalculateJitter

	return m
}
