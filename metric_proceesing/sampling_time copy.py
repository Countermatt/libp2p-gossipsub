import pandas as pd
import os
import csv
import matplotlib.pyplot as plt
import ast

def calculate_average(lst):
    if not lst:
        return 0  # Handle the case when the list is empty to avoid division by zero.
    if len(lst)>0:
        total = sum(lst)
    else:
        return 0
    average = total / len(lst)
    return average

def read_csv(file_path):
    data = []
    with open(file_path, "r") as csv_file:
        csv_reader = csv.reader(csv_file)
        for row in csv_reader:
            data.append(row)

    return data

if __name__ == "__main__":
    
    csv.field_size_limit(104857000000006)
    current_directory = os.getcwd()
    file_path = os.path.join(current_directory, "result/result.csv")
    file_path2 = os.path.join(current_directory, "result/result3.csv")

    data2 = read_csv(file_path2)
    data2 = sorted(data2, key=lambda x: x[0])
    name = []
    builder = []
    validator = []
    nvalidator = []   
    data = []
    for x in data2:
        if len(ast.literal_eval(x[3])) > 0:
            if x[0].split("-")[-3] == "v20":
                if x[0].split("-")[-1][-3] == "s":
                    #data.append(["0100-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["0100-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
                else: 
                    #data.append(["0100-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["0100-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
            elif x[0].split("-")[-3] == "v100":
                if x[0].split("-")[-1][-3] == "s":
                    #data.append(["0500-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["0500-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
                else:
                    #data.append(["0500-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["0500-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
            elif x[0].split("-")[-3] == "v200":
                if x[0].split("-")[-1][-3] == "s":
                    #data.append(["1000-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["1000-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
                else:
                    #data.append(["1000-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["1000-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
            else:
                if x[0].split("-")[-1][-3] == "s":
                    #data.append(["2000-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["2000-0"+x[0].split("-")[-1][-2:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
                else:
                    #data.append(["2000-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[3]), calculate_average(ast.literal_eval(x[2]))])
                    data.append(["2000-"+x[0].split("-")[-1][-3:],ast.literal_eval(x[2]), calculate_average(ast.literal_eval(x[2]))])
        builder.append(float(x[1]))
    data = sorted(data, key=lambda x: x[0])

    y = []
    i = 0
    for x in data:
        #tmp = [h + x[2] + 5 for h in x[1]]
        tmp = [h + 2 for h in x[1]]
        y.append(tmp)
    name = [x[0] for x in data]
    print(validator)
    # Create a violin plot using Matplotlib
    plt.figure(figsize=(8, 6))
    plt.boxplot(y)

    plt.xlabel('# experiment')
    plt.ylabel('Time in second')
    plt.title('validator node time to sample')
    plt.xticks(ticks=range(0, len(name)), labels=name, rotation=45)
    plt.axhline(y=4, color='r', linestyle='--', label='Limit validators sampling')
    #plt.axhline(y=10, color='r', linestyle='--', label='Limit regulars sampling')
    plt.savefig('result/distri_validator.pdf', format='pdf')
