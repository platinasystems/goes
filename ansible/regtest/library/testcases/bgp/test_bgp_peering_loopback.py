#!/usr/bin/python
""" Test/Verify BGP Peering Loopback """

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
module: test_bgp_peering_loopback
author: Platina Systems
short_description: Module to test and verify bgp loopback config.
description:
    Module to test and verify bgp configurations and log the same.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    config_file:
      description:
        - OSPF configurations added in Quagga.conf file.
      required: False
      type: str
    cmd_file:
      description:
        - Route add commands to be executed on switch.
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
- name: Verify bgp peering loopback
  test_bgp_peering_loopback:
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

    out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_bgp_peering_loopback(module):
    """
    Method to verify bgp peering loopback config.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    config_file = module.params['config_file'].splitlines()
    cmd_file = module.params['cmd_file'].splitlines()

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Run commands mentioned in cmd file to assign loopback ip
    # and add static routes
    for cmd in cmd_file:
        execute_commands(module, cmd)

    # Restart and check Quagga status
    execute_commands(module, 'service quagga restart')
    time.sleep(35)
    execute_commands(module, 'service quagga status')

    # Get all bgp routes
    cmd = "vtysh -c 'sh ip bgp neighbors'"
    bgp_out = execute_commands(module, cmd)

    for line in config_file:
        line = line.strip()
        if 'neighbor' in line and 'remote-as' in line:
            config = line.split()
            neighbor_ip = config[1]
            remote_as = config[3]
            if neighbor_ip not in bgp_out or remote_as not in bgp_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'bgp neighbor {} '.format(neighbor_ip)
                failure_summary += 'is not present in the output of '
                failure_summary += 'command {}\n'.format(cmd)

            if 'BGP state = Established' not in bgp_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'bgp state of neighbor {} '.format(neighbor_ip)
                failure_summary += 'is not Established in the output of '
                failure_summary += 'command {}\n'.format(cmd)

    # Remove added static routes and revert the loopback ip
    for line in cmd_file:
        if 'ifconfig' in line:
            cmd = 'ifconfig lo 127.0.0.1 netmask 255.0.0.0'
        else:
            cmd = line.replace('add', 'del')

        execute_commands(module, cmd)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            config_file=dict(required=False, type='str', default=''),
            cmd_file=dict(required=False, type='str', default=''),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_bgp_peering_loopback(module)

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

