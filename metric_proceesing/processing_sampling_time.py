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

def get_result(data):
    result = []
    i = 1

    tmp = i
    while i < len(data)-1:
        if data[i+1]["Block"] != None:
            
            if not(data[i]["Block"] == data[tmp]["Block"]):
                if int(data[i]["TimeStamp"]) - int(data[tmp]["TimeStamp"]) > 0:
                    result.append(int(data[i]["TimeStamp"]) - int(data[tmp]["TimeStamp"]))
                    tmp = i
        i += 1
    result.append(int(data[i]["TimeStamp"]) - int(data[tmp]["TimeStamp"]))
    return result

def list_directories(directory_path):
    directories = [entry for entry in os.listdir(directory_path) if os.path.isdir(os.path.join(directory_path, entry))]
    return directories

def calculate_average(lst):
    if not lst:
        return None  # Handle the case when the list is empty to avoid division by zero.

    total = sum(lst)
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
            tmp_result = get_result(data)
            if file.split("-")[0] == "nonvalidator":
                result[2].append(calculate_average(tmp_result))
            elif file.split("-")[0] == "validator":
                result[1].append(calculate_average(tmp_result))
            else:
                result[0].append(calculate_average(tmp_result))
        name = directory.split(":")[-1]
        data = [name, str(calculate_average(result[0])), str(calculate_average(result[1])), str(calculate_average(result[2]))]
        append_to_csv( os.path.join(directory_path, "result.csv"), data)
    