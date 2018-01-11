#!/usr/bin/python
""" Test/Verify BIRD Peering State Propagation """

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
module: test_bird_peering_state_propagation
author: Platina Systems
short_description: Module to test and verify bird configuration.
description:
    Module to test and verify bird configurations and log the same.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    package_name:
      description:
        - Name of the package installed (e.g. quagga/frr/bird).
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
        - Comma separated string of eth interfaces to bring down/up.
      required: False
      type: str
    propagate_switch:
      description:
        - Name of the switch on which we need to propagate it's state.
      required: False
      type: str
    is_convergence:
      description:
        - Flag to indicate if we need to verify bgp convergence
      required: False
      type: bool
      default: False
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
- name: Verify bird peering state propagation
  test_bird_peering_state_propagation:
    switch_name: "{{ inventory_hostname }}"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
    log_dir_path: "{{ log_dir_path }}"
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

    if 'service' in cmd and 'restart' in cmd:
        out = None
    else:
        out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_routes(module, dummy_interfaces_list):
    """
    Method to verify if dummy interface routes are present or not.
    :param module: The Ansible module to fetch input parameters.
    :param dummy_interfaces_list: List of dummy interfaces to verify.
    :return: Failure summary if any.
    """
    global RESULT_STATUS
    failure_summary = ''
    switch_name = module.params['switch_name']

    cmd = 'birdc show route'
    all_routes = execute_commands(module, cmd)

    for ip in dummy_interfaces_list:
        if ip not in all_routes:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'dummy interface {} '.format(ip)
            failure_summary += 'is not present in the output '
            failure_summary += 'of command {}\n'.format(cmd)

    return failure_summary


def verify_bird_peering_state_propagation(module):
    """
    Method to verify bird peering state propagation.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    dummy_interfaces_list = []
    switch_name = module.params['switch_name']
    package_name = module.params['package_name']
    spine_list = module.params['spine_list']
    leaf_list = module.params['leaf_list']
    propagate_switch = module.params['propagate_switch']
    is_convergence = module.params['is_convergence']

    # Get the current/running configurations
    execute_commands(module, 'cat /etc/bird/bird.conf')

    # Restart and check package status
    execute_commands(module, 'service {} restart'.format(package_name))
    execute_commands(module, 'service {} status'.format(package_name))

    # Get the list of dummy interfaces to verify
    switches_list = spine_list + leaf_list
    for switch in switches_list:
        ip = '192.168.{}.0'.format(switch[-2::])
        dummy_interfaces_list.append(ip)

    if not is_convergence:
        eth_list = module.params['eth_list'].split(',')
        # Verify required routes are present or not
        failure_summary += verify_routes(module, dummy_interfaces_list)

        # Bring down interfaces on propagate switch
        if switch_name == propagate_switch:
            for eth in eth_list:
                execute_commands(module, 'ifconfig eth-{}-1 down'.format(eth))

        # Wait for 200 sec in order for route to become unreachable
        time.sleep(200)

        # Verify required routes are present or not
        if switch_name != propagate_switch:
            temp_list = dummy_interfaces_list
            temp_list.remove('192.168.{}.0'.format(propagate_switch[-2::]))
            failure_summary += verify_routes(module, temp_list)

        # Bring up interfaces on propagate switch
        if switch_name == propagate_switch:
            for eth in eth_list:
                execute_commands(module, 'ifconfig eth-{}-1 up'.format(eth))

        # Wait for 12 sec for BGP to send/receive state update message
        time.sleep(12)

        # Verify required routes are present or not
        if switch_name != propagate_switch:
            failure_summary += verify_routes(module, dummy_interfaces_list)
    else:
        # Verify required routes are present or not
        if not propagate_switch:
            failure_summary += verify_routes(module, dummy_interfaces_list)
        else:
            dummy_interfaces_list.remove('192.168.{}.0'.format(
                propagate_switch[-2::]))
            failure_summary += verify_routes(module, dummy_interfaces_list)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            package_name=dict(required=False, type='str'),
            spine_list=dict(required=False, type='list', default=[]),
            leaf_list=dict(required=False, type='list', default=[]),
            eth_list=dict(required=False, type='str'),
            propagate_switch=dict(required=False, type='str', default=''),
            is_convergence=dict(required=False, type='bool', default=False),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_bird_peering_state_propagation(module)

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

