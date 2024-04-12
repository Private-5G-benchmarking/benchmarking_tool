// packetlib provides utility functions for processing and comparing packets
package matchlib


func IsPacketMatchSequenceNr(parsedPacket map[string]interface{}, p map[string]interface{}) bool {
	//This is extracted into its own function to make it easier later
	return p["sequence_nr"] == parsedPacket["sequence_nr"] 
}