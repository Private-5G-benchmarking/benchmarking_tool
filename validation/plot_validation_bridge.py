import csv
import plotly.graph_objs as go

def read_csv(filename):
    data = {}
    with open(filename, 'r') as file:
        reader = csv.reader(file)
        next(reader)
        for row in reader:
            key = row[0]
            value1 = float(row[1])  
            value2 = float(row[2])
            nr_lost_packets = value1-value2

            if data.get(key) is not None:
                data[key]["values"].append(nr_lost_packets)
                data[key]["avg"] = sum(data[key]["values"]) / len(data[key]["values"])
            else:
                data[key] = {"avg": nr_lost_packets, "values": [nr_lost_packets]}

    return data

def plot_histogram(data):
    keys = list(data.keys())
    print(keys)
    lost_packets = [entry["avg"] for entry in data.values()]
    print(lost_packets)

    fig = go.Figure()
    fig.add_trace(go.Bar(y=lost_packets, x=keys, orientation='v'))

    fig.update_layout(
        title='Histogram showing number of lost packets between ingress and egress of bridge',
        xaxis_title='Number of lost packets',
        yaxis_title='Keys',
        barmode='overlay'
    )

    fig.show()

if __name__ == "__main__":
    filename = "/home/shared/validation_backups/sensitivity_analysis_bridge/results_16.csv"  # Change this to your CSV file's name
    data = read_csv(filename)
    plot_histogram(data)
