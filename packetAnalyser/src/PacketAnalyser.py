import pandas as pd
import numpy as np
from typing import Dict

"""
POINT FORMAT
srcip, dstip, psize, encapsulated_psize, rx_ts, tx_ts
"""


class PacketAnalyser:
    BYTE_SIZE = 8

    def __init__(self):
        self.kpi_calculators = {
            # "throughput": PacketAnalyser.calculate_throughput,
            "packet_loss": PacketAnalyser.calculate_packet_loss,
            "packet_owd": PacketAnalyser.calculate_packet_owd,
            "packet_interarrival_time": PacketAnalyser.calculate_packet_interarrival_time,
            "packet_jitter": PacketAnalyser.calculate_packet_jitter,
            "availability": PacketAnalyser.calculate_availability,
        }

    @staticmethod
    def calculate_throughput(df: pd.DataFrame, **kwargs) -> float:
        """
        Calculates throughput in bit/sec

        :param df: pandas DataFrame of all timeseries points

        :return float: througput in bits/sec
        """
        start_ts, end_ts = df["rx_ts"].min(), df["rx_ts"].max()
        total_bytes = df["psize"].sum()

        tput_byte = total_bytes / (end_ts - start_ts)
        tput_bit = tput_byte * PacketAnalyser.BYTE_SIZE

        return tput_bit

    @staticmethod
    def calculate_packet_loss(df: pd.DataFrame, **kwargs) -> float:
        """
        Calculates packet loss in percent. Divides the number of missing rx
        packets by the total number of transmitted packets.

        :param df: pandas DataFrame of all timeseries points

        :return float: packet loss in percent
        """
        total_pkts = df.shape[0]
        lost_pkts = df["found_match"].sum()

        pkt_loss = 1 - lost_pkts / total_pkts
        pkt_loss_pct = pkt_loss * 100

        return pkt_loss_pct

    @staticmethod
    def calculate_packet_owd(df: pd.DataFrame, **kwargs) -> pd.Series:
        """
        Calculates the one-way delay time for all points.

        :param df: pandas DataFrame of all timeseries points

        :return pd.Series: one-way delay per packet
        """
        owds = df["rx_ts"] - df["tx_ts"]

        owds[df["found_match"] == False] = np.nan

        return owds

    @staticmethod
    def calculate_packet_jitter(
        df: pd.DataFrame, bootstrap=True, method=1, **kwargs
    ) -> pd.Series:
        """
        Calculates jitter according to one of two methods. Method 1 corresponds
        to RFC3350 specification of jitter, method 2 corresponds to RFC3393
        speficication of jitter.
        (1):
            D(i, j) = (R_j - S_j) - (R_i - S_i)
            J(i) = J(i-1) + (|D(i-i, i)| - J(i-1))/16
        OR
        (2)
            J(i) = (Rx_i - Ts_i) - (Rx_i-1 - Tx_i-1)

        Both of the equations yield singleton statistics. Lone statistics are
        not interesting, but must be seen in a collection of singletons.
        Method (2) requires sampling of J(i)'s in non-overlapping intervals,
        i.e. defining n intervals in the measurement lifespan and calculating
        n ipdv-statistics, one for each interval. The method accounts for noise
        in a selection function F, which might filter out packets with
        transmit times deemed to large to be realistic.

        :param df: pandas DataFrame of all timeseries points

        :return pd.Series: jitter (ipdv) of each packet
        """
        if method == 1:
            one_way_delay = df["rx_ts"] - df["tx_ts"]

            def d(i, j):
                return one_way_delay.iloc[j] - one_way_delay.iloc[i]

            ipdv_measurements = pd.Series(
                [pd.NaT for _ in range(one_way_delay.shape[0])]
            )
            first_value = (
                one_way_delay.iloc[0] if bootstrap else pd.NaT
            )  # TODO: THIS IS INCORRECT
            ipdv_measurements.iloc[0] = first_value
            for i in range(1, one_way_delay.shape[0]):
                ipdv = (
                    ipdv_measurements.iloc[i - 1]
                    + (abs(d(i - 1, i)) - ipdv_measurements.iloc[i - 1]) / 16
                )
                ipdv_measurements.iloc[i] = ipdv

            return ipdv_measurements

    @staticmethod
    def calculate_packet_interarrival_time(
        df: pd.DataFrame, tx=True, **kwargs
    ) -> pd.Series:
        """
        Calculates the interarrival time of packets in df.

        :param df: pandas DataFrame of all timeseries points
        :param tx: boolean value, if True, return iat of transmit timestamps.
        If False, return iat of receive timestamps.

        :return pd.Series: iat per-packet (NaN for first packet)
        """
        if tx:
            column = "tx_ts"
        else:
            column = "rx_ts"

        return df[column].diff()

    @staticmethod
    def calculate_availability(df: pd.DataFrame, thresholds: Dict, **kwargs) -> Dict:
        """
        Calculates availability of the packets based on a number of thresholds
        in the thresholds-dictionary. The availability is the number of packets
        not exceeding the limit from the thresholds.

        :param df: pandas DataFrame of all timeseries points
        :param thresholds: dictionary of format Dict[str, float] where the key
        is the alias for the threshold, e.g. '2ms', and the value is the
        corresponding floating point value in seconds.

        :return Dict: dictionary with the same keys as thresholds. The values
        are the rate of packets from df satisfying the given threshold value.
        """

        def calculator(series: pd.Series, lim: float) -> float:
            length = series.shape[0]
            within_limit = series[series <= lim].count()

            return within_limit / length

        availabilities = {}
        one_way_delays = df["rx_ts"] - df["tx_ts"]
        for key in thresholds.keys():
            availabilities[key] = calculator(one_way_delays, thresholds[key])

        return availabilities
