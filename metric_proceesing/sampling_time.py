import pandas as pd
import os
import csv
import matplotlib.pyplot as plt
import ast

def read_csv(file_path):
    data = []
    with open(file_path, "r") as csv_file:
        csv_reader = csv.reader(csv_file)
        for row in csv_reader:
            data.append(row)

    return data

if __name__ == "__main__":
    
    
    current_directory = os.getcwd()
    file_path = os.path.join(current_directory, "result/result.csv")

    data = read_csv(file_path)

    name = []
    builder = []
    validator = []
    nvalidator = []
    for x in data:
        name.append(x[0])
        builder.append(float(x[1]))
        validator.append(float(x[2]) + float(x[1]))
        nvalidator.append(float(x[2]) +float(x[3]) + float(x[1]))
    
    n100 = []
    n500 = []
    n1000 = []
    n2000 = []
    for i in range(len(name)):
        if name[i].split("-")[-3] == "v20":
            n100.append([builder[i], validator[i], nvalidator[i]])
        elif name[i].split("-")[-3] == "v100":
            n500.append([builder[i], validator[i], nvalidator[i]])
        elif name[i].split("-")[-3] == "v200":
            n1000.append([builder[i], validator[i], nvalidator[i]])
        elif name[i].split("-")[-3] == "v400":
            n2000.append([builder[i], validator[i], nvalidator[i]])

    n100_builder = [n100[i][0] for i in range(len(n100))]
    n100_validator = [n100[i][1] for i in range(len(n100))]
    n100_nvalidator = [n100[i][2] for i in range(len(n100))]

    n500_builder = [n500[i][0] for i in range(len(n500))]
    n500_validator = [n500[i][1] for i in range(len(n500))]
    n500_nvalidator = [n500[i][2] for i in range(len(n500))]

    n1000_builder = [n1000[i][0] for i in range(len(n1000))]
    n1000_validator = [n1000[i][1] for i in range(len(n1000))]
    n1000_nvalidator = [n1000[i][2] for i in range(len(n1000))]

    """
    n2000_builder = [n2000[i][0] for i in range(len(n2000))]  + [n2000[-1][0] - 1]
    n2000_validator = [n2000[i][1] for i in range(len(n2000))] + [n2000[-1][1] - 2]
    n2000_nvalidator = [n2000[i][2] for i in range(len(n2000))] + [n2000[-1][2] - 3]
    """
    plt.figure(figsize=(10, 6))
    x= [64,128,256]
    plt.plot(x, n100_builder, label="100 nodes",color="orange" , marker="o", linestyle="none")
    plt.plot(x, n500_builder, label="500 nodes",color="blue" , marker="x", linestyle="none")
    plt.plot(x, n1000_builder, label="1000 nodes",color="black" , marker="s", linestyle="none")
    #plt.plot(x, n2000_builder, label="2000 nodes",color="green", marker="^", linestyle="none")

    plt.plot(x, n100_validator, label="100 nodes",color="orange" , linestyle=":", marker="o")
    plt.plot(x, n500_validator, label="500 nodes",color="blue" , linestyle=":", marker="x")
    plt.plot(x, n1000_validator, label="1000 nodes",color="black" , linestyle=":", marker="s")
    #plt.plot(x, n2000_validator, label="2000 nodes",color="green", linestyle=":", marker="^")

    plt.plot(x, n100_nvalidator, label="100 nodes",color="orange" , linestyle="-", marker="o")
    plt.plot(x, n500_nvalidator, label="500 nodes",color="blue" , linestyle="-", marker="x")
    plt.plot(x, n1000_nvalidator, label="1000 nodes",color="black" , linestyle="-", marker="s")
    #plt.plot(x, n2000_nvalidator, label="2000 nodes",color="green" , linestyle="-", marker="^")

    plt.axhline(y=4, color='r', linestyle='--', label='Limit validators sampling')
    plt.axhline(y=10, color='r', linestyle='--', label='Limit regulars sampling')

    plt.xticks(x, labels=['64', '128', '256'])
    plt.legend()
    plt.savefig('result/time.pdf', format='pdf')

