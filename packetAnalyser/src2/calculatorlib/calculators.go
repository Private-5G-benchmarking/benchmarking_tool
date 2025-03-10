package calculatorlib

import (
	"fmt"
	"math"

	"benchmarking/packetAnalyzer/parselib"
)

// calculateDiff subtracts val1 from val2. Some custom error-handling is added
// for cases where val1 or val2 are negative as this is not within their
// intended domain. If either of the values are negative, the integer
// -101010 is returned as a sufficiently random value that will never be
// calculated randomly given the nature of the values.
func calculateDiff(val1, val2 float64) float64 {
	if val1 < 0 || val2 < 0 {
		return -101010
	}
	return val2 - val1
}

// CalculateOneWayDelay accepts an array of Packets and calculates the difference
// between their receive timestamp (Rx_ts) and their transmit timestamp (Tx_ts).
// It returns the difference in seconds.
func calculateOneWayDelay(packets []*parselib.PacketInfo) (map[float64]float64, error) {
	one_way_delays := make(map[float64]float64, len(packets))

	for _, packet := range packets {
		one_way_delay, err := packet.OneWayDelay()
		if err != nil {
			fmt.Println(err)
			continue
		}
		if one_way_delay < 0 {
			one_way_delays[packet.Tx_ts] = -101010 // INVALID
		} else {
			one_way_delays[packet.Tx_ts] = one_way_delay
		}
	}

	return one_way_delays, nil
}

// CalculateInterArrivalTime accepts and array of Packets and calculates the
// difference in transmit timestamps (Tx_ts) between successive packets.
// It returns the differences in seconds.
func calculateInterarrivalTime(packets []*parselib.PacketInfo) (map[float64]float64, error) {
	inter_arrival_times := make(map[float64]float64)

	for i := 1; i < len(packets); i++ {
		inter_arrival_time := packets[i].Tx_ts - packets[i-1].Tx_ts
		inter_arrival_times[packets[i].Tx_ts] = inter_arrival_time
	}

	return inter_arrival_times, nil
}

// CalculateJitter (CalculateIPDV) accepts an array of Packets and calculates
// the IPDV for each packet according to
// RFC 3550. It returns the IPDVs in seconds.
func calculateRFC3550Jitter(packets []*parselib.PacketInfo) (map[float64]float64, error) {
	jitters := make(map[float64]float64, len(packets))
	one_way_delays, err := calculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	jitters[packets[0].Tx_ts] = 0

	for i := 1; i < len(packets); i++ {
		diff := calculateDiff(one_way_delays[packets[i-1].Tx_ts], one_way_delays[packets[i].Tx_ts])
		var jitter float64

		if diff == -101010 {
			jitter = -101010 // INVALID
		} else {
			prevJitter := jitters[packets[i-1].Tx_ts]
			if prevJitter == -101010 {
				prevJitter = 0
			}

			jitter = prevJitter + (math.Abs(diff)-prevJitter)/16
		}

		jitters[packets[i].Tx_ts] = jitter
	}

	return jitters, nil
}

// CalculateJitter (CalculateIPDV) accepts an array of Packets and calculates
// the IPDV for each packet according to
// RFC 3393. It returns the IPDVs in seconds.
func calculateRFC3393Jitter(packets []*parselib.PacketInfo) (map[float64]float64, error) {
	jitters := make(map[float64]float64, len(packets))
	one_way_delays, err := calculateOneWayDelay(packets)

	if err != nil {
		return jitters, err
	}

	jitters[packets[0].Rx_ts] = 0

	for i := 1; i < len(packets); i++ {
		diff := calculateDiff(one_way_delays[packets[i-1].Tx_ts], one_way_delays[packets[i].Tx_ts])
		var jitter float64

		if diff == -101010 {
			jitter = -101010 // INVALID
		} else {
			jitter = diff
		}

		jitters[packets[i].Tx_ts] = jitter
	}

	return jitters, nil
}

// CalculateThroughput calculates the raw amount of bits being transmitted
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

		// PacketInfo.Psize unit is bytes - multiply by 8 to get bits
		tputs[currentSecond] += float32(packet.Psize) * 8
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

		numPacketsCurrentSecond++

		if !packet.Found_match {
			numLostPacketsCurrentSecond++
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

		numPacketsCurrentSecond++

		owd, err := packet.OneWayDelay()
		if err != nil || !packet.Found_match || owd < 0 {
			continue
		}
		if float32(owd) <= threshold {
			numPacketsWithinThresholdCurrentSecond++
		}
	}
	availabilities[currentSecond] = numPacketsWithinThresholdCurrentSecond / numPacketsCurrentSecond

	return availabilities, nil
}

// func getAvailabilityCalculators() map[string]func(packets []*parselib.PacketInfo) (map[int64]float32, error) {
// 	thresholds := map[string]float32{
// 		"2ms":   0.002,
// 		"4ms":   0.004,
// 		"8ms":   0.008,
// 		"16ms":  0.016,
// 		"32ms":  0.032,
// 		"64ms":  0.064,
// 		"128ms": 0.128,
// 	}

// 	availabilityFuncs := make(map[string]func(packets []*parselib.PacketInfo) (map[int64]float32, error))

// 	for thresh_str, thresh_val := range thresholds {
// 		foo := func(packets []*parselib.PacketInfo) (map[int64]float32, error) {
// 			return calculateAvailability(packets, thresh_val)
// 		}
// 		availabilityFuncs[thresh_str] = foo
// 	}

// 	return availabilityFuncs
// }

type PerPacketCalculatorMap map[string]func([]*parselib.PacketInfo) (map[float64]float64, error)
type AggregateCalculatorMap map[string]func([]*parselib.PacketInfo) (map[int64]float32, error)

func GetPerPacketCalculatorMap() PerPacketCalculatorMap {
	m := make(PerPacketCalculatorMap)

	m["packet_interarrival_time"] = calculateInterarrivalTime
	m["packet_jitter_weighted"] = calculateRFC3550Jitter
	m["packet_jitter_raw"] = calculateRFC3393Jitter
	m["packet_owd"] = calculateOneWayDelay

	return m
}

func GetAggregateCalculatorMap() AggregateCalculatorMap {
	m := make(AggregateCalculatorMap)

	m["throughput"] = calculateThroughput
	m["packet_loss"] = calculatePacketLoss

	return m
}

func CalculatePerPacketKPIs(calculatorMap PerPacketCalculatorMap, packets []*parselib.PacketInfo) (map[string]map[float64]float64, error) {
	valueMap := make(map[string]map[float64]float64)

	for kpiName, fn := range calculatorMap {
		values, err := fn(packets)

		if err != nil {
			return nil, err
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

	thresholds := map[string]float32{
		"2ms":   0.002,
		"4ms":   0.004,
		"8ms":   0.008,
		"16ms":  0.016,
		"32ms":  0.032,
		"64ms":  0.064,
		"128ms": 0.128,
	}
	for thresh_str, thresh_val := range thresholds {
		availabilities, err := calculateAvailability(packets, thresh_val)
		fmt.Println(thresh_val)
		if err != nil {
			return nil, err
		}
		valueMap["availability_"+thresh_str] = availabilities
	}

	return valueMap, nil
}
