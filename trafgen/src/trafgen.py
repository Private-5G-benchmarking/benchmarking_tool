import stl_path
from trex.stl.api import *

import argparse
from datetime import datetime

ETH_HDR_SZ = 18 # B
IP_HDR_SZ = 20 # B
UDP_HDR_SZ = 8 # B

IPs = {
    "TG": "192.168.2.111",
    "RPI_47": "172.30.0.47",
}

class SimpleLogger:
    _loglevel = None
    def __init__(self, loglevel="debug"):
        self._loglevel = SimpleLogger.encode_log_level(loglevel)

    def log(self, msg: str, level="info", end="\n"):
        if self._loglevel <= SimpleLogger.encode_log_level(level):
            print(msg, end=end)

    def set_loglevel(self, loglevel: str):
        if loglevel != None:
            self._loglevel = SimpleLogger.encode_log_level(loglevel)

    @staticmethod
    def encode_log_level(loglevel):
        if loglevel == "debug":
            return 1
        if loglevel == "info":
            return 2
        if loglevel == "warn":
            return 3
        if loglevel == "error":
            return 4

logger = SimpleLogger(loglevel="info")

class STLS1(object):
    def create_stream(self, **kwargs):
        source_ip, \
        destination_ip, \
        source_port, \
        destination_port, \
        start_delay, \
        payload_size \
        = \
            kwargs.get("source_ip"), \
            kwargs.get("destination_ip"), \
            kwargs.get("source_port"), \
            kwargs.get("destination_port"), \
            kwargs.get("start_delay"), \
            kwargs.get("payload_size")
        
        if payload_size < 16:
            raise ValueError(f"Property 'payload_size' must be at least 16. Received ${payload_size}")
        
        base_pkt = Ether() / IP(src=source_ip, dst=destination_ip) / UDP(dport=destination_port, sport=source_port) / ("x" * payload_size)
        mode = STLTXCont()

        flow_var = STLVmFlowVar(name="seqno", min_value=0, max_value=67108864, op="inc") # takes 4 characters, 8 bytes?
        wr_flow_var = STLVmWrFlowVar(fv_name="seqno", pkt_offset=(ETH_HDR_SZ + IP_HDR_SZ + UDP_HDR_SZ - 4))

        vm = STLScVmRaw([flow_var, wr_flow_var], cache_size=255)

        pkt = STLPktBuilder(pkt=base_pkt, vm=vm)

        return STLStream(name="s0", packet=pkt, mode=mode, isg=start_delay)

    def get_streams(self, **kwargs):
        return [ self.create_stream(**kwargs) ]
    
    
def print_stream(stream):
    """
    Logs information about a TRex stream.
    The function logs the name of the stream and the packet it produces.
    Also logs the name of the next stream, if it exists.

    :param stream: STLStream

    :returns None
    """
    name = stream.get_name()
    next = stream.get_next()

    logger.log(f"Stream name: {name}", level="debug")
    logger.log("Stream packet:", level="debug")
    pkt = stream.to_code()
    logger.log(pkt, level="debug")
    

    if next is None:
        return
    
    logger.log(f"Next stream name {next.get_name()}...", level="debug")

def simple_client(**kwargs):
    mult, \
    duration, \
    = \
        kwargs.get("mult"), \
        kwargs.get("duration"), \
    
    c = STLClient()
    
    try:
        streams = STLS1().get_streams(**kwargs)
        ports = [0]
        c.connect()
        
        if c.is_connected():
            logger.log("client connected to TRex server\n")
        else:
            return
        
        c.reset(ports=ports)

        logger.log(f"Adding streams to ports {ports}:", level="debug")
        for stream in streams:
            print_stream(stream)

        logger.log("\n")

        c.add_streams(streams, ports=ports)
        c.clear_stats()

        starttime = datetime.now()

        logger.log("starting traffic generation...")
        c.start(ports=ports, mult=mult, duration=duration)
        c.wait_on_traffic(ports=ports)
        logger.log(f"traffic finished generating in {datetime.now() - starttime}")

        warn = c.get_warnings()

        if warn:
            logger.log(warn, level="warn")
    except Exception as e:
        logger.log(e, level="error")
    finally:
        c.disconnect()
        logger.log("client disconnected from TRex server")

if __name__ == "__main__":
    # ================ SETUP ARGPARSER ================
    parser = argparse.ArgumentParser(description="Automation script for TRex stateless")
    parser.add_argument("-m", "--mult", dest="mult", default="100pps", type=str)
    parser.add_argument("-d", "--duration", dest="duration", default=15, type=int)
    parser.add_argument("--srcip", dest="source_ip", default=IPs["TG"], type=str)
    parser.add_argument("--dstip", dest="destination_ip", default=IPs["RPI_47"], type=str)
    parser.add_argument("--srcport", dest="source_port", default=8000, type=int)
    parser.add_argument("--dstport", dest="destination_port", default=9000, type=int)
    parser.add_argument("--payload_size", dest="payload_size", default=16, type=int)
    parser.add_argument("--startdelay", dest="start_delay", default=0.0, type=float)
    parser.add_argument("--loglevel", dest="loglevel", default="info", type=str, choices=["debug", "info", "warn", "error"])
    parser.add_argument("--mode", dest="tx_mode", default="continuous", choices=["continuous", "single_burst", "multi_burst"])

    args = parser.parse_args()

    # ================ SETUP LOGGER ================

    logger.set_loglevel(args.loglevel)

    # ================ FINAL VALIDATION AND RUN ================

    if args.tx_mode != "continuous":
        logger.log("ERROR: script only supports tx_mode 'continuous'", level="error")
    else:
        simple_client(
            mult=args.mult,
            duration=args.duration,
            source_ip=args.source_ip,
            destination_ip=args.destination_ip,
            source_port=args.source_port,
            destination_port=args.destination_port,
            start_delay=args.start_delay,
            payload_size=args.payload_size
        )