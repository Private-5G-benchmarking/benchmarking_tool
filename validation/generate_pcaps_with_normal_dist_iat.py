import scapy.all as scapy
import numpy as np
import time

def generate_udp_packet(dstIP, dstPort, payload_size):
    """
    Function to generate a UDP packet with given payload size.
    """
    l2 = scapy.Ether()
    l3 = scapy.IP(dst=dstIP)
    l4 = scapy.UDP(dport=dstPort)
    # l4.add_payload(b'\x00\x21\x32\x31\x23\x21\x31\x23\x12\x32\x13\x21\x23\x12\x31\x23')
    return l2/l3/l4

def generate_packet_stream(packet_count, payload_size, mu, sigma, dstIP, dstPort):
    """
    Function to generate a stream of UDP packets with specified inter-arrival time.
    """
    packet_stream = []
    previous_time = time.time()
    for i in range(packet_count):
        packet = generate_udp_packet(dstIP, dstPort, payload_size)
        packet.time = previous_time + np.random.normal(mu, sigma)
        previous_time = packet.time
        packet_stream.append(packet)
    return packet_stream

def write_to_pcap(packet_stream, output_file):
    """
    Function to write a packet stream to a pcap file.
    """
    scapy.wrpcap(output_file, packet_stream)

def main():
    packet_count = 10000  # Number of packets to generate
    payload_size = 16   # Payload size of each packet
    mu = 0.0001         # Mean of inter-arrival time in seconds
    sigma = 0.00001     # Standard deviation of inter-arrival time in seconds
    output_file = "/home/shared/validation_files/effect_tcpreplay/replay_3.pcapng"
    dstIP = "172.30.0.47"
    dstPort = 9999

    # Generate packet stream
    packet_stream = generate_packet_stream(packet_count, payload_size, mu, sigma, dstIP, dstPort)

    # Write packet stream to pcap file
    write_to_pcap(packet_stream, output_file)
    print(f"Packet stream with {packet_count} packets written to {output_file}.")

if __name__ == "__main__":
    main()
