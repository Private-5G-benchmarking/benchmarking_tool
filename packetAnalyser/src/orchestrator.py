from DataConnection import DataConnection
from PacketAnalyser import PacketAnalyser
from performanceAnalysis.PerformanceLogger import PerformanceLogger
from datetime import datetime, timedelta
from time import perf_counter
import argparse


def timed(func, perform_timing):
    def wrapper(*args, **kwargs):
        if not perform_timing:
            return {"call_result": func(*args, **kwargs)}

        t1 = perf_counter()
        result = func(*args, **kwargs)
        t2 = perf_counter()

        return {"call_result": result, "time": t2 - t1}

    return wrapper


class PacketAnalysisOrchestrator:
    """
    Class orchestrating (1) querying of experiment data from influx,
    (2) calculates KPIs based on the queried data, and (3) writes
    the calculated KPIs into new tables in influx.

    (2) expects points with the following fields:
        [rx_ts, tx_ts, psize]

    (3) writes the calculated KPIs into two tables: <kpi_table_name>, which
    is given in the command line argument -d/--destinationtable, and
    <kpi_table_name>_aggregate.

    <kpi_table_name> fields:
        [packet_owd, packet_jitter, packet_interarrival_time]

    <kpi_table_name>_aggregate fields:
        [
            packet_loss,
            throughput,
            availability_2ms,
            availability_4ms,
            availability_8ms,
            availability_16ms,
            availability_32ms,
            availability_64ms
        ] # throughput not tested and is currently not calculated
    """

    def __init__(self, connect_dict, measurement_name, perform_logging=False):
        self.DataConnection = DataConnection(connect_dict, measurement_name)
        self.measurement_name = measurement_name
        self.PacketAnalyser = PacketAnalyser()
        self._perform_logging = perform_logging
        self.performance_logger = PerformanceLogger("performanceLog.csv")

    def run_analysis(self, table_name, skip_write):
        query = f'from(bucket: "{con["bucket"]}") |> range(start: 0) |> filter(fn: (r) => r._measurement == "{table_name}" ) |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")'

        print("Querying data...")
        timed_query_fn = timed(self.DataConnection.query, perform_timing=self._perform_logging)
        timed_query = timed_query_fn(query)
        data = timed_query["call_result"]
        print("Finished querying data...")

        self.convert_ts_string_to_seconds(data, ["rx_ts", "tx_ts"])
        kpis_dict = self.PacketAnalyser.kpi_calculators

        calculated_kpis = self.calculate_kpis(kpi_dict=kpis_dict, data=data.sort_values(by="tx_ts"))

        if self._perform_logging:
            self.performance_logger.set_row_point("packets", data.shape[0])
            self.performance_logger.set_row_point("query", timed_query["time"])
            self.performance_logger.write_logs_to_file()

        if skip_write:
            return

        print("Writing data...")
        self.DataConnection.write_packet_kpis(calculated_kpis, data)
        self.DataConnection.write_aggregate_kpis(calculated_kpis)
        print("Finished writing data...")

    def calculate_kpis(self, kpi_dict, data):
        calculated_kpis = {}
        availability_thresholds = {
            "2ms": 0.002,
            "4ms": 0.004,
            "8ms": 0.008,
            "16ms": 0.016,
            "32ms": 0.032,
            "64ms": 0.064,
        }

        for kpi_key in kpi_dict.keys():
            print(f"Calculating {kpi_key}...")

            calculator = kpi_dict[kpi_key]
            timed_calculator = timed(calculator, perform_timing=self._perform_logging)
            result = timed_calculator(data, thresholds=availability_thresholds)

            calculated_kpis[kpi_key] = result["call_result"]

            print(f"Finished calculating {kpi_key}...")

            if self._perform_logging:
                self.performance_logger.set_row_point(kpi_key, result["time"])

        return calculated_kpis

    def parse_timestamp(self, timestamp_str):
        # Split the timestamp into seconds and nanoseconds parts
        seconds_str, nanoseconds_str = timestamp_str[:-1].split(".")

        # Parse seconds part
        timestamp = datetime.strptime(seconds_str, "%Y-%m-%dT%H:%M:%S")

        # Add nanoseconds part
        nanoseconds = int(nanoseconds_str.ljust(9, "0"))  # Pad to 9 digits
        timestamp += timedelta(microseconds=nanoseconds // 1000)

        return timestamp

    def convert_ts_string_to_seconds(self, data, cols):
        for col_name in cols:
            data[col_name] = data[col_name].apply(
                lambda x: self.parse_timestamp(x).timestamp()
            )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Packet analysis script")
    parser.add_argument(
        "-s", "--sourcetable", dest="sourcetable", default="packet_info2", type=str
    )
    parser.add_argument(
        "-d",
        "--destinationtable",
        dest="destinationtable",
        default="slettmeg",
        type=str,
    )
    parser.add_argument("--skipwrite", dest="skipwrite", default=False, type=bool)
    parser.add_argument(
        "-b", "--bucket", dest="bucket", default="5gbenchmarking", type=str
    )
    parser.add_argument("--skip-log", dest="skiplog", action=argparse.BooleanOptionalAction)

    args = parser.parse_args()

    con = {
        "bucket": args.bucket,
        "org": "5gbenchmarking",
        "token": "OnjSj1CE5Feqwdb1c7w1SPj2EJVV6yWpHHUe93HkfKyVeBo4TN5BrcfVezKJ6sUk50XPVyvPVH1ljSv4JaypzQ==",
        "url": "http://localhost:8086",
    }

    orch = PacketAnalysisOrchestrator(
        connect_dict=con,
        measurement_name=args.destinationtable,
        perform_logging=not args.skiplog,
    )
    orch.run_analysis(table_name=args.sourcetable, skip_write=args.skipwrite)
