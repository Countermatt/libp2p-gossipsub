import logging
import enoslib as en
import os
import datetime

def add_time(original_time, hours=0, minutes=0, seconds=0):
    time_delta = datetime.timedelta(hours=hours, minutes=minutes, seconds=seconds)
    new_time = original_time + time_delta
    return new_time

def convert_seconds_to_time(seconds):
    hours, remainder = divmod(seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    return hours, minutes, seconds
dir_path = os.path.dirname(os.path.realpath(__file__))

en.init_logging(level=logging.INFO)
en.check()

#Change to your Grid5000 user
login = "mapigaglio"
site = "nancy"
cluster = "gros"

#Node type configuration
nb_real_node = 10
nb_node = 360
nb_builder = 2
nb_validator = 18
nb_node_per_cpu = nb_node//nb_real_node

arguments = 20  # nb_second


#conf each node
node_conf = [[0 for x in range(3)] for y in range(nb_real_node)]

k=0
tmp = nb_builder
while tmp > 0:
    if k == nb_real_node:
        k = 0
    node_conf[k][0] += 1
    k += 1
    tmp -= 1

k=0
tmp = nb_validator
while tmp > 0:
    if k == nb_real_node:
        k = 0
    node_conf[k][1] += 1
    k += 1
    tmp -= 1

k=0
tmp =  nb_node - nb_validator - nb_builder
while tmp > 0:
    if k == nb_real_node:
        k = 0
    node_conf[k][2] += 1
    k += 1
    tmp -= 1

network = en.G5kNetworkConf(type="prod", roles=["experiment_network"], site=site)

conf = (
    en.G5kConf.from_settings(job_name="Louvain-job-1", walltime="0:02:00")
    .add_network_conf(network)
    #.add_machine(roles=["experiment"], cluster="gros", nodes=nb_node-1, primary_network=network)
    .add_machine(roles=["first"], cluster=cluster, nodes=nb_real_node, primary_network=network)
    .finalize()
)

# This will validate the configuration, but not reserve resources yet
provider = en.G5k(conf)
roles, networks = provider.init(force_deploy=True)
roles = en.sync_info(roles, networks)

#Network emulation
netem = en.NetemHTB()
(
    netem.add_constraints(
        src=roles["first"],
        dest=roles["first"],
        delay="70ms",
        rate="1gbit",
        symmetric=True,)
)

netem.deploy()
netem.validate()

i = 0
for x in roles["first"]:
    with en.actions(roles=x, on_error_continue=True, background=True) as p:
        p.shell("/home/" + login + "/run.sh " + str(arguments) + " /home/" + login + "/result/" + " " + str(node_conf[i][0]) + " " + str(node_conf[i][1]) + " " + str(node_conf[i][2]))
    i += 1

x = datetime.datetime.now()
h,m,s = convert_seconds_to_time(arguments)
print("Begin at: ",x)
print("Expected to finish at: ",add_time(x,h,m,s))

# Release all Grid'5000 resources
#netem.destroy()
#provider.destroy()