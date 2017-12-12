#!/usr/bin/python
""" Test/Verify OSPF Hello & Dead Intervals """

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
module: test_ospf_intervals
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
    eth_list:
      description:
        - Comma separated string of eth interfaces to bring down/up.
      required: False
      type: str
    hello_timer:
      description:
        - Value of hello timer interval.
      required: False
      type: str
    dead_timer:
      description:
        - Value of dead timer interval.
      required: False
      type: str
    package_name:
      description:
        - Name of the package installed (e.g. quagga/frr/bird).
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
- name: Verify ospf hello and dead time intervals
  test_ospf_intervals:
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


def verify_ospf_intervals(module):
    """
    Method to verify ospf intervals.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    package_name = module.params['package_name']
    hello_timer = module.params['hello_timer']
    dead_timer = module.params['dead_timer']
    eth_list = module.params['eth_list'].split(',')

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Restart and check package status
    execute_commands(module, 'service {} restart'.format(package_name))
    execute_commands(module, 'service {} status'.format(package_name))

    # Get and verify ospf neighbor relationships
    cmd = "vtysh -c 'sh ip ospf neighbor'"
    all_neighbors = execute_commands(module, cmd)

    for eth in eth_list:
        if 'eth-{}-1'.format(eth) not in all_neighbors:
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'ospf neighbor is not showing up '
            failure_summary += 'for eth-{}-1 interface\n'.format(eth)

    # Verify hello and dead time intervals
    for eth in eth_list:
        cmd = "vtysh -c 'sh ip ospf interface eth-{}-1'".format(eth)
        out = execute_commands(module, cmd)

        if 'Hello {}'.format(hello_timer) not in out:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'hello timer interval is not configured '
            failure_summary += 'for eth-{}-1 interface\n'.format(eth)

        if 'Dead {}'.format(dead_timer) not in out:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'dead timer interval is not configured '
            failure_summary += 'for eth-{}-1 interface\n'.format(eth)

        # Bring down the interface
        down_cmd = 'ifconfig eth-{}-1 down'.format(eth)
        execute_commands(module, down_cmd)

        # Wait for dead time interval
        time.sleep(int(dead_timer))

        # Now, ospf neighbor for this down interface should not be displayed
        cmd = "vtysh -c 'sh ip ospf neighbor'"
        all_neighbors = execute_commands(module, cmd)

        if 'eth-{}-1'.format(eth) in all_neighbors:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'ospf neighbor is showing up '
            failure_summary += 'for eth-{}-1 interface '.format(eth)
            failure_summary += 'even after bringing down this interface\n'

        # Bring up the interface which we had brought down
        up_cmd = 'ifconfig eth-{}-1 up'.format(eth)
        execute_commands(module, up_cmd)

        # Wait for hello time interval
        time.sleep(int(hello_timer))

        # Now, ospf neighbor for this up interface should be displayed
        cmd = "vtysh -c 'sh ip ospf neighbor'"
        all_neighbors = execute_commands(module, cmd)

        if 'eth-{}-1'.format(eth) not in all_neighbors:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'ospf neighbor is not showing up '
            failure_summary += 'for eth-{}-1 interface '.format(eth)
            failure_summary += 'even after bringing up this interface\n'

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth_list=dict(required=False, type='str'),
            hello_timer=dict(required=False, type='str'),
            dead_timer=dict(required=False, type='str'),
            package_name=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_ospf_intervals(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}.log'.format(module.params['hash_name'])
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
        hash_dict=HASH_DICT,
        log_file_path=log_file_path
    )

if __name__ == '__main__':
    main()

