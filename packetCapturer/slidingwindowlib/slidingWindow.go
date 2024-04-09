package slidingwindowlib

import (
	"encoding/csv"
	"math"
	"packetCapturer/csvlib"
	"packetCapturer/samplelib"
	"time"
)

func HandlePacketMatch(
	writer *csv.Writer,
	packet map[string]interface{},
	slidingWindowPacket map[string]interface{},
) {

	txTs := slidingWindowPacket["packet_ts"].(time.Time)
	packetSize := slidingWindowPacket["psize"].(int)

	packetStruct := csvlib.NewPacketInfo(slidingWindowPacket["src_ip"].(string), slidingWindowPacket["dst_ip"].(string), packetSize, packet["psize"].(int), packet["packet_ts"].(time.Time),txTs,true)

	if !txTs.Before(packet["packet_ts"].(time.Time)) {
		packetStruct.Srcip = packet["src_ip"].(string)
		packetStruct.Dstip = packet["dst_ip"].(string)
		packetStruct.Rx_ts =  float64(packet["packet_ts"].(time.Time).UnixNano())/math.Pow10(9)
		packetStruct.Tx_ts = float64(txTs.UnixNano())/math.Pow10(9)
		packetStruct.Psize = packet["psize"].(int)
		//TODO this is probably not the correct way to handle this field
		packetStruct.Encapsulated_psize = packetSize

	}
	packetStruct.WriteToCsv(writer)
}

func EmptySlidingWindow(slidingWindow []map[string]interface{}, writer *csv.Writer, cdf []float32) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		packet := csvlib.NewPacketInfo(p["src_ip"].(string), p["dst_ip"].(string), p["psize"].(int), p["psize"].(int), p["packet_ts"].(time.Time), p["packet_ts"].(time.Time), false)
		if samplelib.Sample(cdf) == 1 {
			packet.WriteToCsv(writer)
			localRowCount += 1
		}
	}
	return localRowCount
}
