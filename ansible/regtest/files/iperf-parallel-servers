#!/bin/bash
# Run multiple parallel instances of iperf servers

# Assumes the port numbers used by the servers start at 5001 and increase
# e.g. 5001, 5002, 5003, ...
# If you want something different, then change the following parameter value
# to be: firstport - 1
base_port=5000

# Command line input: number of servers
# E.g. 5
num_servers=$1
shift

# Command line input: base report file name
# E.g. report
report_base=$1
shift

# Optional command line input: other iperf options
# E.g. -u
iperf_options="$*"

# Run iperf multiple times
for i in `seq 1 $num_servers`; do

	# Set server port
	server_port=$(($base_port+$i));

	# Report file includes server port
	report_file=${report_base}-${server_port}.txt

	# Run iperf
	iperf -s -p $server_port $iperf_options &> $report_file &

done
