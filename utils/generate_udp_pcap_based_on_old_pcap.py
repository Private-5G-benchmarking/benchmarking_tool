import sys
import binascii
from scapy.all import *
import argparse

# new_ether_src = "b4:96:91:3b:5c:3b"
# new_ether_dst = "00:1e:42:5c:07:c0"
# # original_src_ip = "192.168.86.12"
# # original_dst_ip = "192.168.86.36"
# new_src_ip = "192.168.2.100"
# new_dst_ip = "172.30.0.47"

def generate_payload(previous_payload):
    """
    Function that is used to generate the custom udp payload that is used
    to match packets later.

    :param previous_payload: the payload that was used previously
    
    :returns new_payload: a new payload that is the previous payload incremented
    by 1 
    """

    payload_number = int.from_bytes(previous_payload, 'big')
    payload_number += 1
    new_payload = payload_number.to_bytes((payload_number.bit_length() + 7) // 8, "big")

    return new_payload

def write(pkt, output_file):
    """
    Function that is used to write a packet to a pcap.

    :param pkt: The packet that will be written to the pcap
    :param output_file: The file path that will be written to

    :returns None
    """
    wrpcap(output_file, pkt, append=True)


def generate_new_pcap(**kwargs):
    """
    :param input_file: The input file that will be used for generation.
    :param output_file: The generated file.

    :returns None
    """
    input_file, output_file, old_src_ip, old_dst_ip, new_src_ip, new_dst_ip, new_sport, new_dport = kwargs.get("input_file"), kwargs.get("output_file"), kwargs.get("old_src_ip"), kwargs.get("old_dst_ip"), kwargs.get("new_src_ip"), kwargs.get("new_dst_ip"), kwargs.get("new_sport"), kwargs.get("new_dport")

    capture = input_file
    pcap = rdpcap(capture)

    packet_counter = 0
    previous_payload = b"00000000"

    for pkt in pcap:
        try:
            if IP in pkt and pkt[IP].src == old_src_ip and pkt[IP].dst == old_dst_ip:
                #Done to avoid packets that do not at least contain L4.
                if UDP not in pkt and TCP not in pkt: 
                    continue
                if UDP in pkt:
                    top_layer = (bytes(pkt[UDP].payload))
                elif TCP in pkt:
                    top_layer = (bytes(pkt[TCP].payload))

                l2 = Ether()
                l3 = IP(src=new_src_ip, dst=new_dst_ip)
                l4 = UDP(sport=new_sport, dport=new_dport)

                new_payload = generate_payload(previous_payload)
                previous_payload = new_payload
                combined_payload = new_payload + top_layer[9:]
                l4.add_payload(combined_payload)

                newPkt = l2/l3/l4

                #Done to keep the original capture time for replays
                newPkt.time = pkt.time
                write(newPkt, output_file)

                packet_counter += 1
        
        except Exception as e:
            print(e)

    print(f"Finished generating new pcap at {output_file} with a total of {packet_counter} packets generated based on {input_file}.")



if __name__ == "__main__":

    parser = argparse.ArgumentParser(prog="Pcap generator", description="This program is used to generate a pcap which can be replayed as udp traffic using tcpreplay")

    parser.add_argument("-i", dest="input_file")
    parser.add_argument("-o", dest="output_file")
    parser.add_argument("-old_src_ip", dest="old_src_ip")
    parser.add_argument("-old_dst_ip", dest="old_dst_ip")
    parser.add_argument("-new_src_ip", dest="new_src_ip")
    parser.add_argument("-new_dst_ip", dest="new_dst_ip")
    parser.add_argument("-new_sport", dest="new_sport", type=int)
    parser.add_argument("-new_dport", dest="new_dport", type=int)

    args = parser.parse_args()

    generate_new_pcap(input_file=args.input_file, output_file=args.output_file, old_src_ip=args.old_src_ip, old_dst_ip=args.old_dst_ip, new_src_ip=args.new_src_ip, new_dst_ip=args.new_dst_ip, new_sport=args.new_sport, new_dport=args.new_dport)