#!/usr/bin/python
""" Test Redis Stats """

#
# This file is part of Ansible
#
# Ansible is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Ansible is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Ansible. If not, see <http://www.gnu.org/licenses/>.
#

import shlex

from collections import OrderedDict

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: test_redis_stats
author: Platina Systems
short_description: Module to verify redis stats.
description:
    Module to test iperf traffic at client and verify rx and tx packets count.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    spine0_eth1_ip:
      description:
        - IP address of eth-1-1 interface of spine0 switch.
      required: False
      type: str
      default: ''
    spine1_eth1_ip:
      description:
        - IP address of eth-1-1 interface of spine1 switch.
      required: False
      type: str
      default: ''
    hash_name:
      description:
        - Name of the hash in which to store the result in redis.
      required: False
      type: str
    log_dir_path:
      description:
        - Path to log directory where logs will be stored.
      required: False
      type: str
"""

EXAMPLES = """
- name: Verify redis stats
  test_redis_stats:
    switch_name: "{{ inventory_hostname }}"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
    log_dir_path: "{{ port_provision_log_dir }}"
"""

RETURN = """
hash_dict:
  description: Dictionary containing key value pairs to store in hash.
  returned: always
  type: dict
"""

RESULT_STATUS = True
HASH_DICT = OrderedDict()


def run_cli(module, cli):
    """
    Method to execute the cli command on the target node(s) and
    returns the output.
    :param module: The Ansible module to fetch input parameters.
    :param cli: The complete cli string to be executed on the target node(s).
    :return: Output/Error or None depending upon the response from cli.
    """
    cli = shlex.split(cli)
    rc, out, err = module.run_command(cli)

    if out:
        return out.rstrip()
    elif err:
        return err.rstrip()
    else:
        return None


def execute_commands(module, cmd):
    """
    Method to execute given commands and return the output.
    :param module: The Ansible module to fetch input parameters.
    :param cmd: Command to execute.
    :return: Output of the commands.
    """
    global HASH_DICT

    out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def test_traffic(module, eth, third_octet, port):
    """
    Method to test iperf traffic at client.
    :param module: The Ansible module to fetch input parameters.
    :param eth: Interface number.
    :param third_octet: Third octet of iperf client ip address.
    :param port: Port number on which iperf server is running.
    :return: Failure summary if any.
    """
    global RESULT_STATUS
    failure_summary = ''
    switch_name = module.params['switch_name']

    cmd = 'iperf -c 10.0.{}.{} -t 2 -p {} -P 1'.format(
        eth, third_octet, port)
    traffic_out = execute_commands(module, cmd)

    if ('Transfer' not in traffic_out and 'Bandwidth' not in traffic_out and
            'Bytes' not in traffic_out and 'bits/sec' not in traffic_out):
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'iperf traffic cannot be verified for '
        failure_summary += 'eth-{}-1 using command {}\n'.format(eth, cmd)

    return failure_summary


def verify_redis_stats(module):
    """
    Method to verify traffic and rx and tx packets count.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    spine0_eth1_ip = module.params['spine0_eth1_ip']
    spine1_eth1_ip = module.params['spine1_eth1_ip']

    spine0_eth_list = [x for x in range(1, 17) if x % 2 == 1]
    spine1_eth_list = [x for x in range(17, 32) if x % 2 == 1]

    spine0_third_octet = spine0_eth1_ip.split('.')[3]
    spine1_third_octet = spine1_eth1_ip.split('.')[3]

    port = 5000
    for eth in spine0_eth_list:
        port += 1
        failure_summary += test_traffic(module, eth, spine0_third_octet, port)

    port = 5000
    for eth in spine1_eth_list:
        port += 1
        failure_summary += test_traffic(module, eth, spine1_third_octet, port)

    # Get RX & TX packets stats from redis & compare it with front panel stats
    for eth in spine0_eth_list + spine1_eth_list:
        redis_rx_count, redis_tx_count = 0, 0
        front_rx_count, front_tx_count = 0, 0

        cmd = 'goes vnet show ha eth-{}-1'.format(eth)
        ha_out = execute_commands(module, cmd)

        if ha_out:
            ha_out = ha_out.lower()
            for line in ha_out:
                line = line.strip()
                if 'port rx packets' in line:
                    line = line.split('port rx packets')
                    redis_rx_count = line[1].strip()
                    redis_rx_count = int(redis_rx_count)

                if 'port tx packets' in line:
                    line = line.split('port tx packets')
                    redis_tx_count = line[1].strip()
                    redis_tx_count = int(redis_tx_count)

            config_out = execute_commands(module, 'ifconfig eth-{}-1'.format(
                eth))

            if config_out:
                config_out = config_out.lower()
                for line in config_out:
                    line = line.strip()
                    if 'rx packets' in line:
                        details = line.split('rx packets:')[1]
                        front_rx_count = details.split()[0]
                        front_rx_count = int(front_rx_count)

                    if 'tx packets' in line:
                        details = line.split('tx packets:')[1]
                        front_tx_count = details.split()[0]
                        front_tx_count = int(front_tx_count)

            if (redis_rx_count != front_rx_count or
                    redis_tx_count != front_tx_count):
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'rx/tx packets counts on front panel stats '
                failure_summary += 'are not matching with redis stats\n'
        else:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'rx and tx packets count cannot be verified '
            failure_summary += 'since output of command {} is None'.format(cmd)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            spine0_eth1_ip=dict(required=False, type='str', default=''),
            spine1_eth1_ip=dict(required=False, type='str', default=''),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_redis_stats(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}.log'.format(module.params['hash_name'])
    log_file = open(log_file_path, 'a')
    for key, value in HASH_DICT.iteritems():
        log_file.write(key)
        log_file.write('\n')
        log_file.write(str(value))
        log_file.write('\n')
        log_file.write('\n')

    log_file.close()

    # Exit the module and return the required JSON.
    module.exit_json(
        hash_dict=HASH_DICT,
        log_file_path=log_file_path
    )

if __name__ == '__main__':
    main()

