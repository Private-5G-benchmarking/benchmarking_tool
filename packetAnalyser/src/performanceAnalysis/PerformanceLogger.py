import pandas as pd
from typing import Dict, List
from datetime import datetime
import os


class PerformanceLogger:
    """
    Logger that writes internal perofmance metrics to a specific (csv)-file.
    Stores performance values in a 'row',  which is a key-value object.
    where each key-value pair is a 'point'.
    """

    row = {}
    resources_folder = "resources"

    def __init__(self, filename: str):
        self._filename = f"{self.resources_folder}/{filename}"

    def set_row_point(self, point_name: str, value):
        """
        Sets a point in the row of the class. Performs no validation, expects
        implementer to behave nicely and verify the type of the value.

        :param point_name: the name of the point to set in the row
        :param value: the value this point shall take. Typically a numeric value
        """
        self.row[point_name] = value

    def get_row_point(self, point_name: str):
        """
        :param point_name: the name of the point in the row

        :returns None if point_name is not in the row, otherwise returns the
        value
        """
        if point_name not in self.row.keys():
            return None
        return self.row[point_name]

    def write_logs_to_file(self):
        """
        Writes the produced row to file. If file exists, validates points of
        row against column names in destination file. If no file exists,
        a new file is created with the row as the only row.
        Writes a timestamp for writing to the row.
        """
        self.set_row_point("ts", datetime.now())
        self._add_row(self.row)

    def _add_row(self, row_dict: Dict):
        row_df = pd.DataFrame(data={key: [value] for key, value in row_dict.items()})

        try:
            df = pd.read_csv(self._filename)
            self._validate_row(row_dict, df.columns)

            df = pd.concat([df, row_df])
        except FileNotFoundError:
            print(f"No csv-file named {self._filename} found. Creating new file")
            df = row_df

            if not os.path.exists(self.resources_folder):
                self._create_resources_folder()

        df.to_csv(self._filename, index=False)

    def _validate_row(self, row_dict: Dict, columns: List[str]):
        row_keys = row_dict.keys()

        if len(row_keys) != len(columns):
            raise Exception(
                f"Number of keys in row dictionary doesn't match number of keys in dataframe"
            )

        for key in row_keys:
            if key not in columns:
                raise Exception(f"Key {key} not in dataframe")
            
    def _create_resources_folder(self):
        os.makedirs(self.resources_folder)
