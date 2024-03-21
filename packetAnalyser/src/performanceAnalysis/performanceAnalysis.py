import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import argparse

def get_test_df():
    test_size = 50
    mu, sigma = 25000, 1000

    rng = np.random.default_rng()
    packet_nums = rng.normal(mu, sigma, size=test_size)
    packet_nums = [round(num) for num in packet_nums]

    def packet_num_to_proc_delay(num_packets: int, rng, scaleup: float) -> float:
        scale = 0.0000025 * scaleup
        stochastic = rng.normal(1, 0.2)

        return (num_packets * scale) * stochastic

    def packet_nums_to_proc_delays(packet_nums: int, rng, scaleup: float):
        return [packet_num_to_proc_delay(num_packets, rng, scaleup) for num_packets in packet_nums]
    
    y_variables = [
             ("packet_loss", 1),
             ("packet_owd", 1.2),
             ("packet_interarrival_time", 1.14),
             ("packet_jitter", 4.3),
             ("availability", 3),
             ("query", 20)
            ]
    
    df_dict = {variable[0]: packet_nums_to_proc_delays(packet_nums, rng, variable[1]) for variable in y_variables}
    df_dict["packets"] = packet_nums

    df = pd.DataFrame(data=df_dict)

    return df


def get_log_df(filename: str):
    return pd.read_csv(filename)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Performance log analysis script")
    parser.add_argument("-t", "--testmode", dest="testmode", default=False, type=bool)
    parser.add_argument("-f", "--filename", dest="filename", default="resources/performanceLog.csv", type=str)

    args = parser.parse_args()

    df = get_test_df() if args.testmode else get_log_df(args.filename)

    fig, ((ax1, ax2), (ax3, ax4), (ax5, ax6)) = plt.subplots(nrows=3, ncols=2)

    plt.subplots_adjust(left=0.1, right=0.9, top=0.9, bottom=0.1, wspace=1, hspace=1)

    ax1.scatter(df["packets"], df["packet_loss"])
    ax1.set_title("packet_loss")
    ax1.set_ylabel("processing time (s)")
    ax1.set_xlabel("processing size (#packets)")
    
    ax2.scatter(df["packets"], df["packet_jitter"])
    ax2.set_title("packet_jitter")
    ax2.set_ylabel("processing time (s)")
    ax2.set_xlabel("processing size (#packets)")
    
    ax3.scatter(df["packets"], df["packet_owd"])
    ax3.set_title("packet_owd")
    ax3.set_ylabel("processing time (s)")
    ax3.set_xlabel("processing size (#packets)")
    
    ax4.scatter(df["packets"], df["packet_interarrival_time"])
    ax4.set_title("packet_interarrival_time")
    ax4.set_ylabel("processing time (s)")
    ax4.set_xlabel("processing size (#packets)")
    
    ax5.scatter(df["packets"], df["availability"])
    ax5.set_title("availability")
    ax5.set_ylabel("processing time (s)")
    ax5.set_xlabel("processing size (#packets)")
    
    ax6.scatter(df["packets"], df["query"])
    ax6.set_title("query")
    ax6.set_ylabel("processing time (s)")
    ax6.set_xlabel("query size (#packets)")

    fig_filename = "resources/performanceAnalysis.png"
    fig.savefig(fig_filename)