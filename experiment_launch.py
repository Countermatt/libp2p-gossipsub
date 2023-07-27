import logging
import enoslib as en
import os
import datetime
import subprocess
import time
#Upload launch script to site frontend
def execute_ssh_command(launch_script, login, site):
    # SSH command in the format 'ssh <username>@<hostname> "<command>"'
    ssh_command = f'scp {launch_script} {login}@access.grid5000.fr:{site}'

    try:
        # Execute the SSH command
        result = subprocess.run(ssh_command, shell=True, capture_output=True, text=True)
        print("Script send to frontend")
        # Check if the command was successful
        if result.returncode == 0:
            # Print the output
            print(result.stdout)
        else:
            # Print the error message
            print(result.stderr)

    except subprocess.CalledProcessError as e:
        print(f"Error occurred while executing SSH command: {e}")

#Get timestamp after end of experiment
def add_time(original_time, hours=0, minutes=0, seconds=0):
    time_delta = datetime.timedelta(hours=hours, minutes=minutes, seconds=seconds)
    new_time = original_time + time_delta
    return new_time

def convert_seconds_to_time(seconds):
    hours, remainder = divmod(seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    return hours, minutes, seconds

def seconds_to_hh_mm_ss(seconds):
    hours = seconds // 3600
    minutes = (seconds % 3600) // 60
    seconds = seconds % 60
    return f"{hours:02d}:{minutes:02d}:{seconds:02d}"

#Experiment node partition between Grid5000 machine
def node_partition(nb_cluster_machine, nb_builder, nb_validator, nb_regular):
    partition = [[0, 0, 0] for i in range(nb_cluster_machine)]

    index = 0
    while nb_builder > 0 or nb_validator > 0 or nb_regular > 0:
        if index == len(partition):
            index  = 0
        if nb_builder > 0:
            partition[index][0] += 1
            nb_builder -= 1
        elif nb_validator > 0:
            partition[index][1] += 1
            nb_validator -= 1            
        elif nb_regular > 0:
            partition[index][2] += 1
            nb_regular -= 1      
        index += 1
    return partition
 
def main():

    #========== Parameters ==========
    #Grid5000 parameters
    login = "mapigaglio" #Grid5000 login
    site = "nancy" #Grid5000 Site See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    cluster = "grisou" #Gride5000 Cluster name See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    job_name = "PANDAS"

    #Node launch script path
    dir_path = os.path.dirname(os.path.realpath(__file__)) #Get current directory path
    launch_script = dir_path +"/" + "run.sh"

    #Experiment parameters
    nb_cluster_machine = 10 #Number of machine booked on the cluster
    nb_experiment_node = 160 #Number of nodes running for the experiment
    nb_builder = 1
    nb_validator = 40
    nb_regular = nb_experiment_node - nb_builder - nb_validator
    exp_duration = 120  #In seconds
    experiment_name = "PANDAS"
    current_datetime = datetime.datetime.now()
    experiment_name += current_datetime.strftime("%Y-%m-%d-%H:%M:%S") 
    #Network parameters
    delay = "10%"
    rate = "1gbit"
    loss = "0%"
    symmetric=True


    #========== Experiment nodes partition on cluster machines ==========
    partition = node_partition(nb_cluster_machine, nb_builder, nb_validator, nb_regular)


    #========== Create and validate Grid5000 and network emulation configurations ==========
    #Log to Grid5000 and check connection
    en.init_logging(level=logging.INFO)
    en.check()
    network = en.G5kNetworkConf(type="prod", roles=["experiment_network"], site=site)
    Job_walltime = seconds_to_hh_mm_ss(exp_duration + 120)
    conf = (
        en.G5kConf.from_settings(job_name=job_name, walltime=Job_walltime)
        .add_network_conf(network)
        .add_machine(roles=["experiment"], cluster=cluster, nodes=nb_cluster_machine, primary_network=network) #Add experiment nodes
        .finalize()
    )

    #Validate Grid5000 configuration
    start = datetime.datetime.now() #Timestamp grid5000 job start
    provider = en.G5k(conf)
    roles, networks = provider.init(force_deploy=True)
    roles = en.sync_info(roles, networks)


    #========== Grid5000 network emulation configuration ==========
    #network parameters
    netem = en.NetemHTB()
    (
        netem.add_constraints(
            src=roles["experiment"],
            dest=roles["experiment"],
            delay=delay,
            rate=rate,
            loss=loss,
            symmetric=symmetric,
        )
    )
    
    #Deploy network emulation
    netem.deploy()
    netem.validate()


    #========== Deploy Experiment ==========
    #Send launch script to Grid5000 site frontend
    execute_ssh_command(launch_script, login, site)
    i = 0
    for x in roles["experiment"]:
        with en.actions(roles=x, on_error_continue=True, background=True) as p:
            builder, validator, regular = partition[i]
            p.shell(f"/home/{login}/run.sh {exp_duration} {experiment_name} {builder} {validator} {regular} {login}")
            i += 1

    start = datetime.datetime.now() #Timestamp grid5000 job start

    #========== Wait job and and release grid5000 ressources ==========
    #Print experiment duration
    h,m,s = convert_seconds_to_time(exp_duration)
    print("Begin at: ",start)
    print("Expected to finish at: ",add_time(start,h,m,s + 10))
    time.sleep(exp_duration + 20)

    #Release all Grid'5000 resources
    netem.destroy()
    provider.destroy()

if __name__ == "__main__":
    main()