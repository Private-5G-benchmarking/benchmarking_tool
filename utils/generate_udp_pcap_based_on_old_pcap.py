import sys
import binascii
from scapy.all import *
import argparse

new_ether_src = "b4:96:91:3b:5c:3b"
new_ether_dst = "00:1e:42:5c:07:c0"
# original_src_ip = "192.168.86.12"
# original_dst_ip = "192.168.86.36"
new_src_ip = "192.168.2.100"
new_dst_ip = "172.30.0.47"
# new_src_port = 9999
# new_dst_port = 7001

parser = argparse.ArgumentParser(prog="Pcap generator", description="This program is used to generate a pcap which can be replayed as udp traffic using tcpreplay")

parser.add_argument("-i", dest="input_file")
parser.add_argument("-o", dest="output_file")
parser.add_argument("-old_src", dest="old_src_ip")
parser.add_argument("-old_dst", dest="old_dst_ip")
parser.add_argument("-new_sport", dest="new_sport", type=int)
parser.add_argument("-new_dport", dest="new_dport", type=int)

args = parser.parse_args()

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
    input_file, output_file, old_src_ip, old_dst_ip, new_sport, new_dport = kwargs.get("input_file"), kwargs.get("output_file"), kwargs.get("old_src_ip"), kwargs.get("old_dst_ip"), kwargs.get("new_sport"), kwargs.get("new_dport")

    capture = input_file
    pcap = rdpcap(capture)



    packet_counter = 0
    previous_payload = b"00000000"

    for pkt in pcap:
        try:
            if IP in pkt and pkt[IP].src == old_src_ip and pkt[IP].dst == old_dst_ip:

                l2 = Ether()
                l3 = IP(src=new_src_ip, dst=new_dst_ip)
                l4 = UDP(sport=new_sport, dport=new_dport)

                new_payload = generate_payload(previous_payload)
                previous_payload = new_payload
                l4.add_payload(new_payload)

                newPkt = l2/l3/l4
                write(newPkt, output_file)

                packet_counter += 1
        
        except Exception as e:
            print(e)

    print(f"Finished generating new pcap at {output_file} with a total of {packet_counter} packets generated based on {input_file}.")


generate_new_pcap(input_file=args.input_file, output_file=args.output_file, old_src_ip=args.old_src_ip, old_dst_ip=args.old_dst_ip, new_sport=args.new_sport, new_dport=args.new_dport)