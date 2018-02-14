#!/usr/bin/python
""" Test On Transit Vlan Configurations """

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
import time

from collections import OrderedDict

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: test_on_transit_vlan_configuration
author: Platina Systems
short_description: Module to verify on transit vlan configurations.
description:
    Module to test on transit vlan configurations.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    spine_list:
      description:
        - List of all spine switches.
      required: False
      type: list
      default: []
    leaf_list:
      description:
        - List of all leaf switches.
      required: False
      type: list
      default: []
    eth_list:
      description:
        - List of eth interfaces.
      required: False
      type: list
      default: []
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
- name: Verify on transit vlan configurations
  test_on_transit_vlan_configurations:
    switch_name: "{{ inventory_hostname }}"
    leaf_list: "{{ groups['leaf'] }}"
    spine_list: "{{ groups['spine'] }}"
    eth_list: "3,21"
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


def verify_vlan_configurations(module):
    """
    Method to verify vlan configurations.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    spine_list = module.params['spine_list']
    leaf_list = module.params['leaf_list']
    eth_list = module.params['eth_list'].split(',')
    third_octet = switch_name[-2::]

    is_leaf = True if switch_name in leaf_list else False

    # Bring down interfaces that are connected to packet generator
    if is_leaf:
        for eth in [x for x in range(1, 33) if x % 2 == 0]:
            execute_commands(module, 'ifconfig eth-{}-1 down'.format(eth))

    # Configure vlan interfaces and assign ip to it
    for eth in eth_list:
        cmd = 'ip link add link eth-{}-1 name eth-{}-1.1 type vlan id {}'.format(
            eth, eth, eth
        )
        execute_commands(module, cmd)
        execute_commands(module, 'ifconfig eth-{}-1.1 192.168.{}.{}/24'.format(
            eth, eth, third_octet
        ))

    # Verify vlan interfaces got created with ip assigned to them
    for eth in eth_list:
        ip_out = execute_commands(module, 'ifconfig eth-{}-1.1'.format(eth))
        if ip_out:
            if '192.168.{}.{}'.format(eth, third_octet) not in ip_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'failed to configure vlan on interface '
                failure_summary += 'eth-{}-1.1\n'.format(eth)

    # Restart quagga
    execute_commands(module, 'service quagga restart')
    time.sleep(15)

    # Get neighbor switch info
    if is_leaf:
        index = leaf_list.index(switch_name)
        leaf_list.remove(switch_name)
        neighbor = leaf_list[0]
    else:
        index = spine_list.index(switch_name)
        spine_list.remove(switch_name)
        neighbor = spine_list[0]

    # Initiate ping between neighbors and verify tcpdump output
    for eth in eth_list:
        if index == 0:
            ping_cmd = 'ping -c 15 -I 192.168.{}.{} 192.168.{}.{}'.format(
                eth, third_octet, eth, neighbor[-2::]
            )
            execute_commands(module, ping_cmd)
        else:
            cmd = 'tcpdump -c 15 -net -i eth-{}-1'.format(eth)
            tcpdump_out = execute_commands(module, cmd)

            if tcpdump_out:
                if ('802.1Q (0x8100)' not in tcpdump_out or
                        'vlan {}'.format(eth) not in tcpdump_out):
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'there are no vlan tagged packets '
                    failure_summary += 'captured in tcpdump for eth-{}-1\n'.format(eth)
            else:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'failed to capture tcpdump output\n'

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            spine_list=dict(required=False, type='list', default=[]),
            leaf_list=dict(required=False, type='list', default=[]),
            eth_list=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_vlan_configurations(module)

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

