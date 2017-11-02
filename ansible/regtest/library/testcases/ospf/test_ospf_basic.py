#!/usr/bin/python
""" Test/Verify OSPF Config """

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
module: test_ospf_basic
author: Platina Systems
short_description: Module to test and verify ospf configurations.
description:
    Module to test and verify ospf configurations and log the same.
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
    spine_ips:
      description:
        - List of all spine switches ip separated by comma.
      required: False
      type: str
    leaf_ips:
      description:
        - List of all leaf switches ip separated by comma.
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
    check_route:
      description:
        - Flag to decide if ospf routes to checked or not.
      required: False
      type: bool
      default: False
"""

EXAMPLES = """
- name: Verify ospf config
  test_ospf_basic:
    switch_name: "{{ inventory_hostname }}"
    spine_list: "{{ groups['spine'] }}"
    spine_ips: "{{ groups['spine'] | map('extract', hostvars, ['ansible_ssh_host']) | join(',') }}"
    leaf_ips: "{{ groups['leaf'] | map('extract', hostvars, ['ansible_ssh_host']) | join(',') }}"
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

    if 'service quagga restart' in cmd:
        out = None
    else:
        out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_ospf_neighbors(module):
    """
    Method to verify if ospf neighbor relationship got established or not.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    check_route = module.params['check_route']

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Restart and check Quagga status
    execute_commands(module, 'service quagga restart')
    execute_commands(module, 'service quagga status')

    if check_route:
        # Get ospf routes
        cmd = "vtysh -c 'show ip route'"
        out = execute_commands(module, cmd)

        # For errors, update the result status to False
        if out is None or 'error' in out:
            RESULT_STATUS = False
            failure_summary += 'On Switch {} '.format(switch_name)
            failure_summary += 'output of command {} is None\n'.format(cmd)
    else:
        # Get ospf neighbors relationships
        cmd = "vtysh -c 'show ip ospf neighbor'"
        out = execute_commands(module, cmd)

        # For errors, update the result status to False
        if out is None or 'error' in out:
            RESULT_STATUS = False
            failure_summary += 'On Switch {} '.format(switch_name)
            failure_summary += 'output of command {} is None\n'.format(cmd)
        else:
            spine_list = module.params['spine_list']

            if switch_name in spine_list:
                neighbour_ips = module.params['leaf_ips']
            else:
                neighbour_ips = module.params['spine_ips']

            for ip in neighbour_ips.split(','):
                if ip not in out:
                    RESULT_STATUS = False
                    failure_summary += 'On Switch {} '.format(switch_name)
                    failure_summary += 'ospf neighbor {} is not present '.format(ip)
                    failure_summary += 'in the output of command {}\n'.format(cmd)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            spine_list=dict(required=False, type='list', default=[]),
            spine_ips=dict(required=False, type='str'),
            leaf_ips=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
            check_route=dict(required=False, type='bool', default=False),
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify ospf neighbors
    verify_ospf_neighbors(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}_'.format(module.params['hash_name']) + '.log'
    log_file = open(log_file_path, 'w')
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

