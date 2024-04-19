package slidingwindowlib

import (
	"encoding/csv"
	"packetCapturer/csvlib"
	"packetCapturer/packetlib"
	"packetCapturer/samplelib"
)

func HandlePacketMatch(
	writer *csv.Writer,
	packet *packetlib.ParsedPacket,
	slidingWindowPacket *packetlib.ParsedPacket,
) {

	packetInfo := csvlib.NewPacketInfo(slidingWindowPacket.SrcIp, slidingWindowPacket.DstIp, slidingWindowPacket.Psize, packet.Psize, packet.Ts, slidingWindowPacket.Ts, true)

	if slidingWindowPacket.Ts >= packet.Ts {
		packetInfo.Srcip = packet.SrcIp
		packetInfo.Dstip = packet.DstIp
		packetInfo.Tx_ts = packet.Ts
		packetInfo.Rx_ts = slidingWindowPacket.Ts
		packetInfo.Psize = packet.Psize
		packetInfo.Encapsulated_psize = slidingWindowPacket.Psize
	} 

	packetInfo.WriteToCsv(writer)
}

func EmptySlidingWindow(slidingWindow []*packetlib.ParsedPacket, writer *csv.Writer, cdf []float32) int {
	localRowCount := 0

	for _, p := range slidingWindow {
		packet := csvlib.NewPacketInfo(p.SrcIp, p.DstIp, p.Psize, p.Psize, p.Ts, p.Ts, false)
		if samplelib.Sample(cdf) == 1 {
			packet.WriteToCsv(writer)
			localRowCount += 1
		}
	}
	return localRowCount
}
