import csv
import glob
import os
from math import *
def read_csv(file_path):
    data = []
    with open(file_path, 'r') as file:
        reader = csv.DictReader(file)
        for row in reader:
            data.append(row)
    return data

def get_result(data, parcel_size):
    result = []

    tmp_result = []
    tmp_block = []
    for row in data:
        if int(row["Block"]) in tmp_block:
            tmp_result[tmp_block.index(int(row["Block"]))].append(int(row["TimeStamp"]))
        else:
            tmp_block.append(int(row["Block"]))
            tmp_result.append([int(row["TimeStamp"])])

    for x in tmp_result:
        if len(x) >= 512*2/parcel_size:
            sort_list = sorted(x)
            #result.append(x[int((512*2)/parcel_size) - 1] - x[0])
            result.append(max(sort_list[ - 1]/1000 - sort_list[0]/1000, (x[int((512*2)/parcel_size) - 1]/1000 - x[0]/1000)))

            result.sort()
    return result

def get_result2(data, parcel_size):
    high = 0
    lower = 9999999999999999999
    for x in data:
        if int(x["TimeStamp"])> high:
            high = int(x["TimeStamp"])
        if int(x["TimeStamp"])< lower:
            lower = int(x["TimeStamp"])
    result = high - lower
    result = result/len(data)
    return result*parcel_size*512

def list_directories(directory_path):
    directories = [entry for entry in os.listdir(directory_path) if os.path.isdir(os.path.join(directory_path, entry))]
    return directories

def calculate_average(lst):
    if not lst:
        return None  # Handle the case when the list is empty to avoid division by zero.
    if len(lst)>0:
        total = sum(lst)
    else:
        return 0
    average = total / len(lst)
    return average

def append_to_csv(file_path, data):
    with open(file_path, 'a', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(data)

if __name__ == "__main__":
    
    current_directory = os.getcwd()
    directory_path = os.path.join(current_directory, "result")
    directory_list = list_directories(directory_path)

    
    for directory in directory_list:
        print(directory)
        path = os.path.join(directory_path, directory)
        
        files = [f for f in os.listdir(path) if os.path.isfile(os.path.join(path, f)) and f.split("-")[-1] == "MessageLog.csv"]
        result = [[] for i in range(3)]
        result2 = [[] for i in range(3)]
        for file in files:
            path_file = os.path.join(path, file)
            data = read_csv(path_file)
            if len(data) > 10:
                tmp_result = get_result(data, int(directory.split("-")[-1].split("s")[-1]))
                tmp_2 = get_result2(data, int(directory.split("-")[-1].split("s")[-1]))
                if file.split("-")[0] == "nonvalidator":
                        #print("a")
                    result[2] += tmp_result
                    result2[2].append(tmp_2)
                elif file.split("-")[0] == "validator":
                        #print("b")
                    result[1] += tmp_result
                    result2[1].append(tmp_2)
                else:
                        #print("c")
                    result[0] = calculate_average(tmp_result)
                    result2[0] = tmp_2
        name = directory.split(":")[-1]
        data = [name, str(result[0]), str(result[1]), str(result[2])]
        append_to_csv( os.path.join(directory_path, "result3.csv"), data)
        data2 = [name, str(result2[0]), str(calculate_average(result2[1])), str(calculate_average(result2[2]))]
        append_to_csv( os.path.join(directory_path, "result2.csv"), data2)
    