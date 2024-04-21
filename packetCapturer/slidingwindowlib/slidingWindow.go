package slidingwindowlib

import (
	"encoding/csv"
	"packetCapturer/csvlib"
	"packetCapturer/matchlib"
	"packetCapturer/packetlib"
	"packetCapturer/samplelib"
)

type SlidingWindow struct {
	Window		[]*packetlib.ParsedPacket
	WindowSize	int
}

func (slidingWindow *SlidingWindow) HandleNewPacket(newPacket *packetlib.ParsedPacket, cdf []float32, writer *csv.Writer) {
	matchFound := false
	for index, p := range slidingWindow.Window {
		if matchlib.IsPacketMatchSequenceNr(newPacket, p) {
			sample := samplelib.Sample(cdf)
			if sample == 1 {
				slidingWindow.HandlePacketMatch(writer, newPacket, p)
			}

			slidingWindow.RemoveFromWindow(index)
			matchFound = true
			break
		}
	}
	if !matchFound {
		// fmt.Println("added to window")
		slidingWindow.AddToWindow(newPacket)
	}

	if slidingWindow.IsWindowFull() {
		exitingElement := slidingWindow.Window[0]
		exitingPacket := csvlib.NewPacketInfo(exitingElement.SrcIp, exitingElement.DstIp, exitingElement.Psize,exitingElement.Psize,exitingElement.Ts,exitingElement.Ts, false)

		if samplelib.Sample(cdf) == 1 {
			exitingPacket.WriteToCsv(writer)
		}
		// slidingWindow.Window = slidingWindow.Window[:1]
		slidingWindow.RemoveFromWindow(0)
	}

}

func (slidingWindow SlidingWindow) IsWindowFull() bool {
	return len(slidingWindow.Window) >= slidingWindow.WindowSize 
}

func (slidingWindow SlidingWindow) HandlePacketMatch(
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

func (slidingWindow SlidingWindow) EmptySlidingWindow(writer *csv.Writer, cdf []float32) {

	for _, p := range slidingWindow.Window {
		packet := csvlib.NewPacketInfo(p.SrcIp, p.DstIp, p.Psize, p.Psize, p.Ts, p.Ts, false)
		if samplelib.Sample(cdf) == 1 {
			packet.WriteToCsv(writer)
		}
	}
}

func (slidingWindow *SlidingWindow) RemoveFromWindow(indexToRemove int) {
	// Ensure the index is within the valid range
	if indexToRemove < 0 || indexToRemove >= len(slidingWindow.Window) {
		return
	}

	// Use append to create a new slice excluding the map at the specified index
	slidingWindow.Window = append(slidingWindow.Window[:indexToRemove], slidingWindow.Window[indexToRemove+1:]...)
}

func (slidingWindow *SlidingWindow) AddToWindow(parsedPacket *packetlib.ParsedPacket) {
	slidingWindow.Window = append(slidingWindow.Window, parsedPacket)
}
