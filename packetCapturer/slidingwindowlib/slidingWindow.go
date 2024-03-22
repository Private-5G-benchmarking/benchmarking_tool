package slidingwindowlib

import (
	"packetCapturer/influxlib"
	"packetCapturer/samplelib"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

func HandlePacketMatch(
	writeHandler func(
		writeAPI api.WriteAPI,
		srcIP string,
		dstIP string,
		txTs time.Time,
		rxTs time.Time,
		packetSize int,
		found_match bool,
		measurementName string,
	),
	writeAPI api.WriteAPI,
	packet map[string]interface{},
	slidingWindowPacket map[string]interface{},
	mearurementName string,
) {
	txTs := slidingWindowPacket["packet_ts"].(time.Time)
	packetSize := slidingWindowPacket["psize"].(int)

	if txTs.Before(packet["packet_ts"].(time.Time)) {
		writeHandler(writeAPI, slidingWindowPacket["src_ip"].(string), slidingWindowPacket["dst_ip"].(string), txTs, packet["packet_ts"].(time.Time), packetSize, true, mearurementName)
	} else {
		writeHandler(writeAPI, packet["src_ip"].(string), packet["dst_ip"].(string), packet["packet_ts"].(time.Time), txTs, packet["psize"].(int), true, mearurementName)
	}
}

func EmptySlidingWindow(slidingWindow []map[string]interface{}, writeAPI api.WriteAPI, cdf []float32, dest_measurement string) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		// Type assertions for map values
		srcIP := p["src_ip"].(string)
		dstIP := p["dst_ip"].(string)
		packet_ts := p["packet_ts"].(time.Time)
		psize := p["psize"].(int)
		if samplelib.Sample(cdf) == 1 {
			influxlib.WriteInfluxDBPoint(writeAPI, srcIP, dstIP, packet_ts, packet_ts, psize, false, dest_measurement)
			localRowCount += 1
		}
	}
	return localRowCount
}
