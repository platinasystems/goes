#!/usr/bin/python
""" Verify Iperf Traffic For Dummy Interface """

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
module: verify_iperf_traffic_for_dummy
author: Platina Systems
short_description: Module to verify iperf traffic for dummy interface.
description:
    Module to verify iperf traffic at client.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    switch_list:
      description:
        - List of adjacent(leaf/spine) switches.
      required: False
      type: list
    packet_size_list:
      description:
        - List of packet sizes described as string.
      required: False
      type: str
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
- name: Verify iperf traffic for dummy interface
  verify_iperf_traffic_for_dummy:
    switch_name: "{{ inventory_hostname }}"
    switch_list: "{{ groups['spine'] }}"
    packet_size_list: "100,1500,5000,12000"
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


def verify_traffic(module):
    """
    Method to verify iperf traffic for dummy interface.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    port = 5000
    failure_summary = ''
    switch_name = module.params['switch_name']
    switch_list = module.params['switch_list']
    packet_size_list = module.params['packet_size_list'].split(',')

    switch_list.remove(switch_name)
    neighbor_switch = switch_list[0]

    self_ip = '192.168.{}.1'.format(switch_name[-2::])
    neighbor_ip = '192.168.{}.1'.format(neighbor_switch[-2::])

    # Verify ping
    packet_count = '3'
    ping_cmd = 'ping -w 3 -c {} -I {} {}'.format(packet_count,
                                                 self_ip, neighbor_ip)
    ping_out = execute_commands(module, ping_cmd)
    if '{} received'.format(packet_count) not in ping_out:
        RESULT_STATUS = False
        failure_summary += 'From switch {}, '.format(switch_name)
        failure_summary += '{} is not getting pinged\n'.format(neighbor_switch)

    # Initiate iperf client and verify traffic
    for size in packet_size_list:
        port += 1
        traffic_cmd = 'iperf -c {} -t 5 -M {} -p {} -P 1 -B {}'.format(
            neighbor_ip, size, port, self_ip
        )
        traffic_out = execute_commands(module, traffic_cmd)

        if ('Transfer' not in traffic_out and 'Bandwidth' not in traffic_out and
                'Bytes' not in traffic_out and 'bits/sec' not in traffic_out):
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'iperf traffic cannot be verified for '
            failure_summary += 'packet size length {} bytes\n'.format(size)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            switch_list=dict(required=False, type='list'),
            packet_size_list=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify iperf traffic
    verify_traffic(module)

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

