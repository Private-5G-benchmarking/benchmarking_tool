package influxlib

import (
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"packetCapturer/packetlib"
)

// WriteInfluxDBPoint accepts a writeAPI and packet information and writes
// said values to InfluxDB using the writeAPI
func WriteInfluxDBPoint(
	writeAPI api.WriteAPI,
	srcIP string,
	dstIP string,
	txTs time.Time,
	rxTs time.Time,
	packetSize int,
	found_match bool,
	measurementName string,
) {
	point := influxdb2.NewPointWithMeasurement(measurementName).
		AddField("src_ip", srcIP).
		AddField("dst_ip", dstIP).
		AddField("tx_ts", txTs).
		AddField("rx_ts", rxTs).
		AddField("psize", packetSize).
		AddField("found_match", found_match).
		SetTime(rxTs)

	writeAPI.WritePoint(point)
}

// processPacketToInfluxPoint creates a map with keys [src_ip, dst_ip,
// packet_ts, sequence_nr, psize] all based on the input packet
func ProcessPacketToInfluxPoint(packet gopacket.Packet, traffic_type string) map[string]interface{} {
	parsedPacket := make(map[string]interface{})
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ipPacket, _ := ipLayer.(*layers.IPv4)
	srcIP := ipPacket.SrcIP.String()
	dstIP := ipPacket.DstIP.String()
	sqNr := ""
		// This needs to take into account a different sq number if using TCP
		if traffic_type == "tcp" {
			sqNr = strconv.FormatUint(uint64(packetlib.GetTCPSequenceNumber(packet)), 10)
		} else {
			sqNr = packetlib.GetSequenceNr(packet)
	
		}
	timestamp := packet.Metadata().CaptureInfo.Timestamp
	packetSize := packet.Metadata().CaptureInfo.CaptureLength

	parsedPacket["src_ip"] = srcIP
	parsedPacket["dst_ip"] = dstIP
	parsedPacket["packet_ts"] = timestamp
	parsedPacket["sequence_nr"] = sqNr
	parsedPacket["psize"] = packetSize

	return parsedPacket
}
