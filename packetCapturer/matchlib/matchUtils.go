// packetlib provides utility functions for processing and comparing packets
package matchlib

import "packetCapturer/packetlib"


func IsPacketMatchSequenceNr(parsedPacket *packetlib.ParsedPacket, p *packetlib.ParsedPacket) bool {
	//This is extracted into its own function to make it easier later
	return p.SequenceNr == parsedPacket.SequenceNr 
}