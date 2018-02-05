#!/usr/bin/python
""" Test/Verify BGP State Propagation """

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
module: test_bgp_state_propagation
author: Platina Systems
short_description: Module to test and verify bgp state propagation.
description:
    Module to test and verify bgp configurations and log the same.
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
    eth_list:
      description:
        - Comma separated string of eth interfaces to bring down/up.
      required: False
      type: str
    propagate_switch:
      description:
        - Name of the switch to propagate state.
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
- name: Verify quagga bgp state propagation
  test_bgp_state_propagation:
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

    if 'service' in cmd or 'dummy' in cmd or 'restart' in cmd:
        out = None
    else:
        out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_bgp_routes(module, route_present):
    """
    Method to verify bgp routes
    :param module: The Ansible module to fetch input parameters.
    :param route_present: Flag to indicate if route needs to be present or not.
    :return: Failure summary if any.
    """
    global RESULT_STATUS
    failure_summary = ''
    switch_name = module.params['switch_name']
    propagate_switch = module.params['propagate_switch']

    # Get all ip routes
    cmd = "vtysh -c 'sh ip route'"
    out = execute_commands(module, cmd)

    if out:
        network = '192.168.{}.1'.format(propagate_switch[-2::])
        route = 'B>* {}'.format(network)

        if route_present:
            if route not in out:
                RESULT_STATUS = False
                failure_summary += 'On Switch {} bgp route '.format(switch_name)
                failure_summary += '{} is not present '.format(route)
                failure_summary += 'in the output of command {}\n'.format(cmd)
        else:
            if route in out:
                RESULT_STATUS = False
                failure_summary += 'On Switch {} bgp route '.format(switch_name)
                failure_summary += '{} is present '.format(route)
                failure_summary += 'in the output of command {} '.format(cmd)
                failure_summary += 'even after shutting down this route\n'
    else:
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'bgp routes cannot be verified '
        failure_summary += 'because output of command {} '.format(cmd)
        failure_summary += 'is None'

    return failure_summary


def verify_quagga_bgp_state_propagation(module):
    """
    Method to verify quagga bgp state propagation.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    package_name = module.params['package_name']
    propagate_switch = module.params['propagate_switch']
    eth_list = module.params['eth_list'].split(',')

    if switch_name == propagate_switch:
        # Add dummy0 interface
        execute_commands(module, 'ip link add dummy0 type dummy')

        # Assign ip to this created dummy0 interface
        cmd = 'ifconfig dummy0 192.168.{}.1 netmask 255.255.255.255'.format(
            switch_name[-2::]
        )
        execute_commands(module, cmd)

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Restart and check package status
    execute_commands(module, 'service {} restart'.format(package_name))
    execute_commands(module, 'service {} status'.format(package_name))

    # Verify bgp routes
    if switch_name != propagate_switch:
        failure_summary += verify_bgp_routes(module, True)

    # Bring down few interfaces on propagate switch
    if switch_name == propagate_switch:
        for eth in eth_list:
            eth = eth.strip()
            cmd = 'ifconfig eth-{}-1 down'.format(eth)
            execute_commands(module, cmd)

    # Wait 200 secs for routes to become unreachable
    time.sleep(200)

    # Verify bgp routes
    if switch_name != propagate_switch:
        failure_summary += verify_bgp_routes(module, False)

    # Bring up interfaces on propagate switch
    if switch_name == propagate_switch:
        for eth in eth_list:
            eth = eth.strip()
            cmd = 'ifconfig eth-{}-1 up'.format(eth)
            execute_commands(module, cmd)

    # Wait 60 secs for routes to become unreachable
    time.sleep(60)

    # Verify bgp routes
    if switch_name != propagate_switch:
        failure_summary += verify_bgp_routes(module, True)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            propagate_switch=dict(required=False, type='str'),
            eth_list=dict(required=False, type='str'),
            package_name=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_quagga_bgp_state_propagation(module)

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

