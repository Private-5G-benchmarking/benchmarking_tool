import csv
import plotly.graph_objs as go

def read_csv(filename):
    data = {}
    with open(filename, 'r') as file:
        reader = csv.reader(file)
        next(reader)
        for row in reader:
            key = row[0]
            ingres = float(row[1])  
            egres = float(row[2])
            nr_lost_packets = ingres-egres

            if data.get(key) is not None:
                data[key]["values"].append(nr_lost_packets)
                data[key]["avg_ost"] = sum(data[key]["values"]) / len(data[key]["values"])
                data[key]["nr_ingres"].append(ingres)
                data[key]["nr_egres"].append(egres)
            else:
                data[key] = {"avg_lost": nr_lost_packets, "values": [nr_lost_packets], "nr_ingres":[ingres], "nr_egres": [egres]}

    return data

def plot_histogram(data):
    keys = list(data.keys())
    lost_packets = [entry["avg_lost"] for entry in data.values()]

    fig = go.Figure()
    fig.add_trace(go.Bar(y=lost_packets, x=keys, orientation='v'))

    fig.update_layout(
        title='Histogram showing number of lost packets between ingress and egress of bridge',
        xaxis_title='Number of lost packets',
        yaxis_title='PPS',
        barmode='overlay'
    )

    fig.show()

def plot_histogram_percent_dropped(data):
    keys = list(data.keys())
    for item in data.items():
        print(f"Nr offered packets are {int(item[0])*60}")
        print(f"nr lost packets are {item[1]['avg_lost']}")
        print(item[1]["avg_lost"]/(int(item[0])*60))
    percent_dropped = [item[1]["avg_lost"]/(int(item[0])*60) for item in data.items()]
    print(percent_dropped)
    fig = go.Figure()
    fig.add_trace(go.Bar(y=percent_dropped, x=keys, orientation='v'))

    fig.update_layout(
        title='Histogram showing number of lost packets between ingress and egress of bridge',
        xaxis_title='Percentage of lost packets',
        yaxis_title='PPS',
        barmode='overlay'
    )

    fig.show()

if __name__ == "__main__":
    filename = "/home/shared/validation_backups/sensitivity_analysis_bridge/results_16.csv" 
    data = read_csv(filename)
    # plot_histogram(data)
    plot_histogram_percent_dropped(data)
