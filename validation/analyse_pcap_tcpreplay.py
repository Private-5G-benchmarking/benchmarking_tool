import pandas as pd
from scapy.all import rdpcap

def pcap_to_dataframe(pcap_file):
    """
    Function to read packets from a pcap file and convert them into a DataFrame.
    """
    packets = rdpcap(pcap_file)
    
    # Initialize lists to store data
    timestamps = []
    src_ips = []
    dst_ips = []
    
    # Extract timestamp, source IP, and destination IP from each packet
    for packet in packets:
        timestamps.append(packet.time)
        src_ips.append(packet[1].src)  # Assuming IP is at layer 1 (index 1)
        dst_ips.append(packet[1].dst)  # Assuming IP is at layer 1 (index 1)
    
    # Create DataFrame
    df = pd.DataFrame({
        'Timestamp': timestamps,
        'SrcIP': src_ips,
        'DstIP': dst_ips
    })

    df['iat'] = df['Timestamp'].diff()
    
    return df

# Usage example
pcap_file = "/home/shared/validation_files/effect_tcpreplay/tcpreplay_validation_capture_1.pcapng"
df = pcap_to_dataframe(pcap_file)
print(df.head(20))
