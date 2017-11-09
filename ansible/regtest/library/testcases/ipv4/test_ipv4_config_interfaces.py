#!/usr/bin/python
""" Test/Verify IPV4 Configuration Interfaces """

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
module: test_ipv4_config_interfaces
author: Platina Systems
short_description: Module to test and verify ipv4 interfaces
description:
    Module to test and verify ipv4 interfaces and log the same.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    eth_port:
      description:
        - Name of the eth port on which ip has been configured.
      required: False
      type: str
    spine_interface_list:
      description:
        - Comma separated list of all spine eth interface addresses.
      required: False
      type: str
    leaf_interface_list:
      description:
        - Comma separated list of all leaf eth interface addresses.
      required: False
      type: str
    spine_list:
      description:
        - List of all spine switches
      required: False
      type: list
      default: []
    leaf_list:
      description:
        - List of all leaf switches
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
- name: Verify ipv4 config interfaces
  test_ipv4_config_interfaces:
    switch_name: "{{ inventory_hostname }}"
    eth_port: 'eth-9-1'
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

    out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_ipv4_config_interfaces(module):
    """
    Method to verify ipv4 configuration interfaces.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    eth_port = module.params['eth_port']
    spine_list = module.params['spine_list']
    leaf_list = module.params['leaf_list']
    spine_interface_list = module.params['spine_interface_list'].split(',')
    leaf_interface_list = module.params['leaf_interface_list'].split(',')

    # Get the current/running configurations
    execute_commands(module, 'cat /etc/network/interfaces')

    # Restart and check GOES status
    key = '{} goes restart'.format(switch_name)
    HASH_DICT[key] = None
    execute_commands(module, 'goes status')

    is_spine = True if switch_name in spine_list else False

    if is_spine:
        index = spine_list.index(switch_name)
        ping_ip = leaf_interface_list[index]
    else:
        index = leaf_list.index(switch_name)
        ping_ip = spine_interface_list[index]

    cmd = "ping -I {} {} -c 3".format(eth_port, ping_ip)
    out = execute_commands(module, cmd)

    if '64 bytes from {}'.format(ping_ip) not in out:
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += '{} command failed\n'.format(cmd)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth_port=dict(required=False, type='str'),
            spine_interface_list=dict(required=False, type='str'),
            leaf_interface_list=dict(required=False, type='str'),
            spine_list=dict(required=False, type='list', default=[]),
            leaf_list=dict(required=False, type='list', default=[]),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_ipv4_config_interfaces(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}_'.format(module.params['hash_name']) + '.log'
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
        hash_dict=HASH_DICT
    )

if __name__ == '__main__':
    main()

