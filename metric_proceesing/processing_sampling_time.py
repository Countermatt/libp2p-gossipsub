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
    for row in data:
        if int(row["# samples"]) >= 512*2/parcel_size:
            result.append(int(row["duration"]))
    if len(result)> 0:
        return result
    return None

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
        for file in files:
            path_file = os.path.join(path, file)
            data = read_csv(path_file)
            if len(data) > 10:
                tmp_result = get_result(data, int(directory.split("-")[-1].split("s")[-1]))
                if file.split("-")[0] == "nonvalidator":
                    if tmp_result != None:
                        result[2] += tmp_result
                elif file.split("-")[0] == "validator":
                    if tmp_result != None:
                        result[1] += tmp_result
                else:
                    if tmp_result != None:
                        result[0] = calculate_average(tmp_result)
        name = directory.split(":")[-1]
        data = [name, str(result[0]), str(calculate_average(result[1])), str(calculate_average(result[2]))]
        append_to_csv( os.path.join(directory_path, "result.csv"), data)
    