package slidingwindowlib

import (
	"encoding/csv"
	"packetCapturer/csvlib"
	"packetCapturer/samplelib"
	"time"
)

func HandlePacketMatch(
	writer *csv.Writer,
	writeHandler func(
		writer *csv.Writer,
		packetStruct csvlib.Packet,
		measurementName string,
	),
	packet map[string]interface{},
	slidingWindowPacket map[string]interface{},
	measurementName string,
) {

	txTs := slidingWindowPacket["packet_ts"].(time.Time)
	packetSize := slidingWindowPacket["psize"].(int)

	packetStruct := csvlib.Packet{
		Srcip: slidingWindowPacket["src_ip"].(string),
		Dstip: slidingWindowPacket["dst_ip"].(string),
		Psize: packetSize,
		Encapsulated_psize: packetSize, 
		Rx_ts: float64(packet["packet_ts"].(time.Time).Unix()),
		Tx_ts: float64(txTs.Unix()),
		Found_match: true,
	}

	if !txTs.Before(packet["packet_ts"].(time.Time)) {
		packetStruct.Srcip = packet["src_ip"].(string)
		packetStruct.Dstip = packet["dst_ip"].(string)
		packetStruct.Rx_ts =  float64(packet["packet_ts"].(time.Time).Unix())
		packetStruct.Tx_ts = float64(txTs.Unix())
		packetStruct.Psize = packet["psize"].(int)
		//TODO this is probably not the correct way to handle this field
		packetStruct.Encapsulated_psize = packet["psize"].(int)

	}

	writeHandler(writer, packetStruct, measurementName)
}

func EmptySlidingWindow(slidingWindow []map[string]interface{}, writer *csv.Writer, cdf []float32, dest_measurement string) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		packet := csvlib.Packet{
			Srcip: p["src_ip"].(string),
			Dstip: p["dst_ip"].(string),
			Psize: p["psize"].(int),
			Encapsulated_psize:  p["psize"].(int), 
			Rx_ts: float64(p["packet_ts"].(time.Time).Unix()),
			Tx_ts: float64(p["packet_ts"].(time.Time).Unix()),
			Found_match: false,
		}
		if samplelib.Sample(cdf) == 1 {
			csvlib.WriteParsedPacketToCsv(writer, packet, dest_measurement)
			localRowCount += 1
		}
	}
	return localRowCount
}
