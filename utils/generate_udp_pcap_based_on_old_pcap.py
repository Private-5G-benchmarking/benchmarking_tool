import sys
import binascii
from scapy.all import *

def generate_payload(previous_payload):
    payload_number = int.from_bytes(previous_payload, 'big')
    payload_number += 1
    new_payload = payload_number.to_bytes((payload_number.bit_length() + 7) // 8, "big")
    return new_payload

def write(pkt):
    wrpcap("waqas-cleaned-udp2.pcapng", pkt, append=True)

capture = "/home/sebastfu/WDM-AWIN_WiresharkLogs2024_03_08.pcapng"
pcap = rdpcap(capture)

new_ether_src = "b4:96:91:3b:5c:3b"
new_ether_dst = "00:1e:42:5c:07:c0"
original_src_ip = "192.168.86.12"
original_dst_ip = "192.168.86.36"
new_src_ip = "192.168.2.100"
new_dst_ip = "172.30.0.47"
new_src_port = 9999
new_dst_port = 7001

previous_payload = b"00000000"

for pkt in pcap:
    try:
        if IP in pkt and pkt[IP].src == original_src_ip and pkt[IP].dst == original_dst_ip:
            l2 = Ether()
            l3 = IP(src=new_src_ip, dst=new_dst_ip)
            l4 = UDP(sport=new_src_port, dport=new_dst_port)
            new_payload = generate_payload(previous_payload)
            previous_payload = new_payload
            l4.add_payload(new_payload)
            newPkt = l2/l3/l4
            write(newPkt)
    
    except Exception as e:
        print(e)
print("done")
