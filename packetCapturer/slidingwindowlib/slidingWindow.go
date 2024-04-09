package slidingwindowlib

import (
	"encoding/csv"
	"packetCapturer/csvlib"
	"packetCapturer/samplelib"
	"time"
)

func HandlePacketMatch(
	writer *csv.Writer,
	packet map[string]interface{},
	slidingWindowPacket map[string]interface{},
	measurementName string,
) {

	txTs := slidingWindowPacket["packet_ts"].(time.Time)
	packetSize := slidingWindowPacket["psize"].(int)

	packetStruct := csvlib.PacketInfo{
		Srcip: slidingWindowPacket["src_ip"].(string),
		Dstip: slidingWindowPacket["dst_ip"].(string),
		Psize: packetSize,
		Encapsulated_psize: packet["psize"].(int), 
		Rx_ts: float64(packet["packet_ts"].(time.Time).UnixNano()),
		Tx_ts: float64(txTs.UnixNano()),
		Found_match: true,
	}

	if !txTs.Before(packet["packet_ts"].(time.Time)) {
		packetStruct.Srcip = packet["src_ip"].(string)
		packetStruct.Dstip = packet["dst_ip"].(string)
		packetStruct.Rx_ts =  float64(packet["packet_ts"].(time.Time).UnixNano())
		packetStruct.Tx_ts = float64(txTs.UnixNano())
		packetStruct.Psize = packet["psize"].(int)
		//TODO this is probably not the correct way to handle this field
		packetStruct.Encapsulated_psize = packetSize

	}
	packetStruct.WriteToCsv(writer, measurementName)
}

func EmptySlidingWindow(slidingWindow []map[string]interface{}, writer *csv.Writer, cdf []float32, dest_measurement string) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		packet := csvlib.PacketInfo{
			Srcip: p["src_ip"].(string),
			Dstip: p["dst_ip"].(string),
			Psize: p["psize"].(int),
			Encapsulated_psize:  p["psize"].(int), 
			Rx_ts: float64(p["packet_ts"].(time.Time).UnixNano()),
			Tx_ts: float64(p["packet_ts"].(time.Time).UnixNano()),
			Found_match: false,
		}
		if samplelib.Sample(cdf) == 1 {
			packet.WriteToCsv(writer, dest_measurement)
			localRowCount += 1
		}
	}
	return localRowCount
}
