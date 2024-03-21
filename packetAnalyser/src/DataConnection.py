import influxdb_client
from influxdb_client.client.write_api import SYNCHRONOUS
from datetime import datetime


class DataConnection:

    def __init__(self, connect_dict, measurement_name):
        self.connect_dict = connect_dict
        self.measurement_name = measurement_name
        self.client = influxdb_client.InfluxDBClient(
            url=connect_dict["url"],
            token=connect_dict["token"],
            org=connect_dict["org"],
        )

    def query(self, query):
        df = self.client.query_api().query_data_frame(
            org=self.connect_dict["org"], query=query
        )
        return df

    def write(self, points):
        self.client.write_api(write_options=SYNCHRONOUS).write(
            bucket=self.connect_dict["bucket"],
            org=self.connect_dict["org"],
            record=points,
        )

    def write_packet_kpis(self, kpis_dict, data):
        points = []
        for i in range(kpis_dict["packet_owd"].shape[0]):
            point = influxdb_client.Point(self.measurement_name).tag(
                "analysis", "analysis"
            )

            for key in kpis_dict.keys():
                if key in ["packet_loss", "availability"]:
                    continue
                point = point.field(key, kpis_dict[key].iloc[i])
            point = point.time(datetime.fromtimestamp(data["rx_ts"].iloc[i]))
            points.append(point)

        self.write(points=points)

    def write_aggregate_kpis(self, kpis_dict):
        aggregate_kpi_names = ["packet_loss"]  # add throughput once it works
        aggregate_kpis = {key: kpis_dict[key] for key in aggregate_kpi_names}
        availabilities = kpis_dict["availability"]

        point = influxdb_client.Point(f"{self.measurement_name}_aggregate").tag(
            "analysis", "analysis"
        )

        for key, value in aggregate_kpis.items():
            point = point.field(key, value)

        for key, value in availabilities.items():
            point = point.field(f"availability_{key}", value)

        point = point.time(datetime.now())

        self.write(points=point)
