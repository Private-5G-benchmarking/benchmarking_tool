// packetlib provides utility functions for processing and comparing packets
package packetlib

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type ParsedPacket struct {
	SrcIp              string
	DstIp              string
	Psize              int
	Ts                 float64 // Format: seconds with nanosecond precision
	SequenceNr		   string
}

func NewParsedPacket(packet gopacket.Packet, l4_protocol string) *ParsedPacket {
	if packet.ErrorLayer() != nil {
		log.Fatal("Error decoding packet:", packet.ErrorLayer().Error())
	}

	//TODO add some error handling here

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ipPacket, _ := ipLayer.(*layers.IPv4)

	sqNr := ""
		// This needs to take into account a different sq number if using TCP
	if l4_protocol == "tcp" {
		sqNr = strconv.FormatUint(uint64(GetTCPSequenceNumber(packet)), 10)
	} else {
		sqNr = GetSequenceNr(packet)
	}


	return &ParsedPacket{
		SrcIp: ipPacket.SrcIP.String(),
		DstIp: ipPacket.DstIP.String(),
		Ts: ConvertNanosecondsToSeconds(packet.Metadata().Timestamp),
		Psize: packet.Metadata().CaptureLength,
		SequenceNr: sqNr,
	}
}

// extractSubstring removes the first 16 characters of a string
func ExtractSubstring(input string) (string, error) {
	// Check if the input string is long enough
	if len(input) < 16 {
		return "", fmt.Errorf("input string is too short")
	}

	// Extract the substring starting after the first 16 characters
	substring := input[28:]
	return substring, nil
}

// getSequenceNr returns the custom sequence number from the packet. It detects
// whether the payload is encapsulated in GTP-U and returns the sequence
// number accordingly
func GetSequenceNr(packet gopacket.Packet) string {
	// Extract UDP layer
	udpLayer := packet.Layer(layers.LayerTypeUDP)

	sequence_number := ""
	if udpLayer != nil {

		udpPacket, _ := udpLayer.(*layers.UDP)
		payload := udpPacket.Payload

		// Check if the payload is GTP encapsulated
		gtpLayer := packet.Layer(layers.LayerTypeGTPv1U)
		if gtpLayer == nil {
			sequence_number = string(payload[0:8])
		} else {
			gtpPacket, _ := gtpLayer.(*layers.GTPv1U)

			// Access the GTP payload
			gtpPayload := gtpPacket.Payload
			testPacket := gopacket.NewPacket(gtpPayload, layers.LayerTypeIPv4, gopacket.Default)
			testUdpLayer := testPacket.Layer(layers.LayerTypeUDP)
			testUdpPacket, _ := testUdpLayer.(*layers.UDP)
			testpayload := testUdpPacket.Payload
			sequence_number = string(testpayload[0:8])
		}
	}
	return sequence_number
}

func GetTCPSequenceNumber(packet gopacket.Packet) uint32 {
	//Check if it is a basic TCP packet without encapsulation in GTP
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcpPacket, _ := tcpLayer.(*layers.TCP)
		return tcpPacket.Seq
	}

	//Check if it is a TCP packet encapsulated in GTP
	gtpLayer := packet.Layer(layers.LayerTypeGTPv1U)
	if gtpLayer != nil {
		gtpPacket, _ := gtpLayer.(*layers.GTPv1U)

		// Access the GTP payload
		gtpPayload := gtpPacket.Payload
		encapsulatedTCPPacket := gopacket.NewPacket(gtpPayload, layers.LayerTypeIPv4, gopacket.Default)
		encapsulatedTCPLayer := encapsulatedTCPPacket.Layer(layers.LayerTypeTCP)
		encapsulatedTCP := encapsulatedTCPLayer.(*layers.TCP)
		return encapsulatedTCP.Seq
	}

	//TODO: find a better value her
	return 0
}


func ConvertNanosecondsToSeconds(timestamp time.Time) float64 {
	return float64(timestamp.UnixNano()) / math.Pow10(9)
}