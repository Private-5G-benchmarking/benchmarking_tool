package calculatorlib

import (
	"fmt"
	"math"

	"benchmarking/packetAnalyzer/parselib"
)

// CalculateOneWayDelay accepts an array of Packets and calculates the difference
// between their receive timestamp (Rx_ts) and their transmit timestamp (Tx_ts).
// It returns the difference in seconds.
func calculateOneWayDelay(packets []*parselib.PacketInfo) ([]float64, error) {
	one_way_delays := make([]float64, len(packets))

	for index, packet := range packets {
		one_way_delay, err := packet.OneWayDelay()
		if err != nil {
			fmt.Println(err)
		}
		one_way_delays[index] = one_way_delay
	}

	return one_way_delays, nil
}

// CalculateInterArrivalTime accepts and array of Packets and calculates the
// difference in transmit timestamps (Tx_ts) between successive packets.
// It returns the differences in seconds.
func calculateInterArrivalTime(packets []*parselib.PacketInfo) ([]float64, error) {
	inter_arrival_times := make([]float64, len(packets))

	for i := 1; i < len(packets); i++ {
		inter_arrival_time := packets[i].Tx_ts - packets[i-1].Tx_ts
		inter_arrival_times[i] = inter_arrival_time
	}

	return inter_arrival_times, nil
}

// CalculateJitter (CalculateIPDV) accepts an array of Packets and calculates
// the IPDV for each packet according to
// RFC 3550. It returns the IPDVs in seconds.
func calculateRFC3550Jitter(packets []*parselib.PacketInfo) ([]float64, error) {
	jitters := make([]float64, len(packets))
	one_way_delays, err := calculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	d := func(i, j int) float64 {
		if one_way_delays[i] < 0 || one_way_delays[j] < 0 {
			return -101010
		}
		return one_way_delays[j] - one_way_delays[i]
	}

	jitters[0] = 0

	for i := 1; i < len(packets); i++ {
		diff := d(i-1, i)
		var jitter float64

		if diff < -101010 {
			jitter = -1
		} else {
			jitter = jitters[i-1] + (math.Abs(d(i-1, i)-jitters[i-1]) / 16)
		}

		jitters[i] = jitter
	}

	return jitters, nil
}

// CalculateJitter (CalculateIPDV) accepts an array of Packets and calculates
// the IPDV for each packet according to
// RFC 3393. It returns the IPDVs in seconds.
func calculateRFC3393Jitter(packets []*parselib.PacketInfo) ([]float64, error) {
	jitters := make([]float64, len(packets))
	one_way_delays, err := calculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	d := func(i, j int) float64 {
		if one_way_delays[i] < 0 || one_way_delays[j] < 0 {
			return -101010
		}
		return one_way_delays[j] - one_way_delays[i]
	}

	jitters[0] = 0

	for i := 1; i < len(packets); i++ {
		diff := d(i-1, i)
		var jitter float64

		if diff < -101010 {
			jitter = -1
		} else {
			jitter = diff
		}

		jitters[i] = jitter
	}

	return jitters, nil
}

// CalculateThroughput calculates the raw amount of bytes being transmitted
// per second. It returns a map where the second is the key and the
// throughput is the value
func calculateThroughput(packets []*parselib.PacketInfo) (map[int64]float32, error) {
	tputs := make(map[int64]float32)
	var currentSecond int64

	for _, packet := range packets {
		if !packet.Found_match {
			continue
		}

		packetSecond := int64(packet.Tx_ts)
		if packetSecond != currentSecond {
			currentSecond = packetSecond
		}

		tputs[currentSecond] += float32(packet.Psize)
	}

	return tputs, nil
}

func calculatePacketLoss(packets []*parselib.PacketInfo) (map[int64]float32, error) {
	ploss := make(map[int64]float32)
	var currentSecond int64 = int64(packets[0].Tx_ts)
	var numPacketsCurrentSecond float32
	var numLostPacketsCurrentSecond float32

	for _, packet := range packets {
		packetSecond := int64(packet.Tx_ts)
		if packetSecond != currentSecond {
			ploss[currentSecond] = numLostPacketsCurrentSecond / numPacketsCurrentSecond

			numLostPacketsCurrentSecond = 0
			numPacketsCurrentSecond = 0

			currentSecond = packetSecond
		}

		numPacketsCurrentSecond += 1

		if !packet.Found_match {
			numLostPacketsCurrentSecond += 1
		}
	}
	ploss[currentSecond] = numLostPacketsCurrentSecond / numPacketsCurrentSecond

	return ploss, nil
}

func calculateAvailability(packets []*parselib.PacketInfo, threshold float32) (map[int64]float32, error) {
	availabilities := make(map[int64]float32)
	currentSecond := int64(packets[0].Tx_ts)
	var numPacketsCurrentSecond float32
	var numPacketsWithinThresholdCurrentSecond float32

	for _, packet := range packets {
		packetSecond := int64(packet.Tx_ts)

		if packetSecond != currentSecond {
			availabilities[currentSecond] = numPacketsWithinThresholdCurrentSecond / numPacketsCurrentSecond

			numPacketsCurrentSecond = 0
			numPacketsWithinThresholdCurrentSecond = 0

			currentSecond = packetSecond
		}

		numPacketsCurrentSecond += 1

		owd, err := packet.OneWayDelay()
		if err != nil || !packet.Found_match {
			continue
		}
		if float32(owd) <= threshold {
			numPacketsWithinThresholdCurrentSecond += 1
		}
	}
	availabilities[currentSecond] = numPacketsWithinThresholdCurrentSecond / numPacketsCurrentSecond

	return availabilities, nil
}

func getAvailabilityCalculators() map[string]func(packets []*parselib.PacketInfo) (map[int64]float32, error) {
	thresholds := map[string]float32{
		"2ms":   0.002,
		"4ms":   0.004,
		"8ms":   0.008,
		"16ms":  0.016,
		"32ms":  0.032,
		"64ms":  0.064,
		"128ms": 0.128,
	}

	availabilityFuncs := make(map[string]func(packets []*parselib.PacketInfo) (map[int64]float32, error))

	for thresh_str, thresh_val := range thresholds {
		foo := func(packets []*parselib.PacketInfo) (map[int64]float32, error) {
			return calculateAvailability(packets, thresh_val)
		}
		availabilityFuncs[thresh_str] = foo
	}

	return availabilityFuncs
}

type PerPacketCalculatorMap map[string]func([]*parselib.PacketInfo) ([]float64, error)
type AggregateCalculatorMap map[string]func([]*parselib.PacketInfo) (map[int64]float32, error)

func GetPerPacketCalculatorMap() PerPacketCalculatorMap {
	m := make(PerPacketCalculatorMap)

	m["packet_owd"] = calculateOneWayDelay
	m["packet_interarrival_time"] = calculateInterArrivalTime
	m["packet_jitter_weighted"] = calculateRFC3550Jitter
	m["packet_jitter_raw"] = calculateRFC3393Jitter

	return m
}

func GetAggregateCalculatorMap() AggregateCalculatorMap {
	m := make(AggregateCalculatorMap)

	m["throughput"] = calculateThroughput
	m["packet_loss"] = calculatePacketLoss

	availabilityCalculators := getAvailabilityCalculators()

	for calc_name, fn := range availabilityCalculators {
		m["availability"+calc_name] = fn
	}

	return m
}

func CalculatePerPacketKPIs(calculatorMap PerPacketCalculatorMap, packets []*parselib.PacketInfo) (map[string][]float64, error) {
	valueMap := make(map[string][]float64)

	for kpiName, fn := range calculatorMap {
		values, error := fn(packets)

		if error != nil {
			return nil, error
		}

		valueMap[kpiName] = values
	}

	return valueMap, nil
}

func CalculateAggregateKPIs(calculatorMap AggregateCalculatorMap, packets []*parselib.PacketInfo) (map[string]map[int64]float32, error) {
	valueMap := make(map[string]map[int64]float32)

	for kpiName, fn := range calculatorMap {
		values, error := fn(packets)

		if error != nil {
			return nil, error
		}

		valueMap[kpiName] = values
	}

	return valueMap, nil
}
