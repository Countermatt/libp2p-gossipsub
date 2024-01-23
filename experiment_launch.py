import logging
import enoslib as en
import os
import datetime
import subprocess

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

#Convert time in second in hh:mm:ss
def seconds_to_hh_mm_ss(seconds):
    hours = seconds // 3600
    minutes = (seconds % 3600) // 60
    seconds = seconds % 60
    return f"{hours:02d}:{minutes:02d}:{seconds:02d}"

#Experiment node partition between Grid5000 machine
def node_partition(nb_cluster_machine, network_size, nb_builder, prop_validator):
    partition = [[0, 0, 0] for i in range(nb_cluster_machine)]
    nb_validator = int(network_size*prop_validator)
    nb_regular = network_size - nb_validator - nb_builder
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
        index +=1
    return partition
 
def main():

    #========== Parameters ==========
    #Grid5000 parameters
    login = "mapigaglio" #Grid5000 login
    site = "nancy" #Grid5000 Site See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    cluster = "gros" #Gride5000 Cluster name See: https://www.grid5000.fr/w/Status and https://www.grid5000.fr/w/Hardware
    job_name = "PANDAS"

    #Node launch script path
    dir_path = os.path.dirname(os.path.realpath(__file__)) #Get current directory path
    launch_script = dir_path +"/" + "run.sh"

    #Experiment parameters

    parcel_size_list = [512]
    network_size_list = [1000]
    nb_run = 1

    k = 0
    nb_expe = len(network_size_list)*len(parcel_size_list)*nb_run
    nb_cluster_machine = 3 #Number of machine booked on the cluster
    prop_validator = 0.20
    exp_duration = 30  #In seconds
    batch_experiment_name = "PANDAS-Gossip-"
    #Network parameters 
    """
    delay = "10%"
    rate = "1gbit"
    loss = "0%"
    symmetric=True
    """
    walltime_in_s = 300+(exp_duration+30)*nb_expe
    #========== Create and validate Grid5000 and network emulation configurations ==========
    #Log to Grid5000 and check connection
    en.init_logging(level=logging.INFO)
    en.check()
    #network = en.G5kNetworkConf(type="prod", roles=["experiment_network"], site=site)
    network = en.G5kNetworkConf(type="kavlan", roles=["experiment_network"], site=site)
    conf = (
        en.G5kConf.from_settings(job_name=job_name, walltime= seconds_to_hh_mm_ss(walltime_in_s))
        #en.G5kConf.from_settings(job_name=job_name, walltime="01:00:00")

        .add_network_conf(network)
        .add_machine(roles=["experiment"], cluster=cluster, nodes=nb_cluster_machine, primary_network=network) #Add experiment nodes
        .finalize()
    )

    #Validate Grid5000 configuration
    provider = en.G5k(conf)
    test = 0
    while test < 10:
        try:
            roles, networks = provider.init(force_deploy=False)
            roles = en.sync_info(roles, networks)
            test += 10
        except:
            test += 1


    #========== Grid5000 network emulation configuration ==========
    #network parameters
    """
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
    """

    #========== Deploy Experiment ==========
    #Send launch script to Grid5000 site frontend
    #execute_ssh_command(launch_script, login, site)
    k = 0

    for batch in range(nb_run):
        for network_size in network_size_list:
            partition = node_partition(nb_cluster_machine, network_size, 1, prop_validator)
            run_name = batch_experiment_name + str(batch)
            for parcel_size in parcel_size_list:
                i = 0
                experiment_name = run_name+"-b1-v"+str(int(network_size*prop_validator))+"-nv"+str(network_size-int(network_size*prop_validator)-1)+"-prs"+str(parcel_size)
                en.run_command(f"mkdir /home/{login}/results/{experiment_name}", roles=roles["experiment"][0])

                for x in roles["experiment"]:
                    if i < len(roles["experiment"]) - 1:
                        with en.actions(roles=x, on_error_continue=True, background=True) as p:
                            builder, validator, regular = partition[i]
                            p.shell(f"/home/{login}/run.sh {exp_duration} {experiment_name} {builder} {validator} {regular} {login} {parcel_size} ")
                            i += 1
                    else:
                        with en.actions(roles=x, on_error_continue=True, background=False) as p:
                            builder, validator, regular = partition[i]
                            p.shell(f"/home/{login}/run.sh {exp_duration} {experiment_name} {builder} {validator} {regular} {login} {parcel_size} ")
                k += 1
                print("Experiment:",k,"/",nb_expe)

        #========== Wait job and and release grid5000 ressources ==========
        # netem.destroy()

if __name__ == "__main__":
    main()
