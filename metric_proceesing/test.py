log_file_path = "result/PANDAS-Gossip--b1-v20-nv79-prs16/gros-2.nancy.grid5000.fr-log"
import pandas as pd

sar_log_path = log_file_path  # Replace with the actual path to the converted SAR text file

# Read the SAR text file into a DataFrame
sar_data = pd.read_csv(sar_log_path, delimiter=r"\s+")

# Display the first few rows of the DataFrame
print(sar_data.head())
