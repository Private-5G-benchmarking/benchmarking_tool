import pandas as pd
import numpy as np

ROW_COUNT = 200000 
COLUMNS = ['Srcip', 'Dstip', 'Psize', 'Encapsulated_psize', 'Rx_ts', 'Tx_ts', 'Found_match']
SRCIP, DSTIP = '0.0.0.0', '1.1.1.1'
PSIZE, ENCAPSULATED_PSIZE = 58, 102



def generate_df(columns, row_count, tx_mu, tx_sigma, owd_mu, owd_sigma, not_loss_prob):
    rows = []
    tx_normal_dist = np.random.normal(loc=tx_mu, scale=tx_sigma, size=(row_count))
    owd_normal_dist = np.random.normal(loc=owd_mu, scale=owd_sigma, size=(row_count))

    for i in range(row_count):
        tx_ts = Tx_ts(i) + tx_normal_dist[i]
        rx_ts = tx_ts + owd_normal_dist[i]
        found_match = 'true' if Found_match(not_loss_prob) else 'false'
        rows.append([SRCIP, DSTIP, PSIZE, ENCAPSULATED_PSIZE, rx_ts, tx_ts, found_match])

    return pd.DataFrame(columns=columns, data=rows)


def Tx_ts(i):
    return i * 0.001 + 1 # i multiplied by iat=0.001

def Found_match(prob):
    return np.random.rand() < prob

if __name__ == '__main__':
    FILE_PATH = "/home/shared/validation_files/csvs/"
    csv_configs = {
        '1': {
            'columns': COLUMNS,
            'row_count': ROW_COUNT,
            'tx_mu': 0,
            'tx_sigma': 0.0006,
            'owd_mu': 0.012,
            'owd_sigma': 0.005,
            'not_loss_prob': 0.99
        },
        '2': {
            'columns': COLUMNS,
            'row_count': ROW_COUNT,
            'tx_mu': 0,
            'tx_sigma': 0.0004,
            'owd_mu': 0.032,
            'owd_sigma': 0.025,
            'not_loss_prob': 0.95
        },
        '3': {
            'columns': COLUMNS,
            'row_count': ROW_COUNT,
            'tx_mu': 0,
            'tx_sigma': 0.0008,
            'owd_mu': 0.02,
            'owd_sigma': 0.05,
            'not_loss_prob': 0.8
        },
    }

    for config_name, config in csv_configs.items():
        df = generate_df(
            columns=config['columns'],
            row_count=config['row_count'],
            tx_mu=config['tx_mu'],
            tx_sigma=config['tx_sigma'],
            owd_mu=config['owd_mu'],
            owd_sigma=config['owd_sigma'],
            not_loss_prob=config['not_loss_prob'],
        )
        print("Created df for ", config_name)
        df.to_csv(f"{FILE_PATH}/analysis_{config_name}.csv", index=False)
        print("Created csv for ", config_name)
