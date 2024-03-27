package calculatorlib

import (
	"math"

	"benchmarking/packetAnalyzer/parselib"
)

func CalculateOneWayDelay(packets []*parselib.Packet) ([]int64, error) {
	one_way_delays := make([]int64, len(packets))

	for index, packet := range packets {
		one_way_delay := packet.Rx_ts - packet.Tx_ts
		one_way_delays[index] = one_way_delay
	}

	return one_way_delays, nil
}

func CalculateInterArrivalTime(packets []*parselib.Packet) ([]int64, error) {
	inter_arrival_times := make([]int64, len(packets))

	for i := 1; i < len(packets); i++ {
		inter_arrival_time := packets[i].Rx_ts - packets[i-1].Rx_ts
		inter_arrival_times[i] = inter_arrival_time
	}

	return inter_arrival_times, nil
}

func CalculateJitter(packets []*parselib.Packet) ([]int64, error) {
	jitters := make([]int64, len(packets))
	one_way_delays, err := CalculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	d := func(i, j int) int64 {
		return one_way_delays[j] - one_way_delays[i]
	}

	jitters[0] = 0

	for i := 1; i < len(packets); i++ {
		jitter := jitters[i-1] + int64(math.Abs((float64(d(i-1, i)-jitters[i-1]) / 16)))
		jitters[i] = jitter
	}

	return jitters, nil
}
