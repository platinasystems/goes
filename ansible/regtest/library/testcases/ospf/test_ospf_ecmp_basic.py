#!/usr/bin/python
""" Test/Verify OSPF ECMP Basic """

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
module: test_ospf_ecmp_basic
author: Platina Systems
short_description: Module to test and verify ospf ecmp basic config.
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
        - Comma separated string of eth interfaces.
      required: False
      type: str
    cost:
      description:
        - Value of cost configured on eth interfaces.
      required: False
      type: str
    config_file:
      description:
        - OSPF configurations added in ospfd.conf file.
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
- name: Verify ospf ecmp basic
  test_ospf_ecmp_basic:
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

    if 'service quagga restart' in cmd or 'ifconfig' in cmd:
        out = None
    else:
        out = run_cli(module, cmd)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(module.params['switch_name'], exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_ospf_ecmp_basic(module):
    """
    Method to verify ospf ecmp basic config.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    cost = module.params['cost']

    # Assign loopback ip
    lo_cmd = 'ifconfig lo 192.168.{}.1 netmask 255.255.255.0'.format(
        switch_name[-2::]
    )
    execute_commands(module, lo_cmd)

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Restart and check Quagga status
    execute_commands(module, 'service quagga restart')
    execute_commands(module, 'service quagga status')

    # Verify configured cost on interfaces
    for eth in module.params['eth_list'].split(','):
        cmd = "vtysh -c 'sh ip ospf interface eth-{}-1'".format(eth)
        ospf_out = execute_commands(module, cmd)

        if 'Cost: {}'.format(cost) not in ospf_out:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'cost {} is not showing up for '.format(cost)
            failure_summary += 'eth-{}-1 interface '.format(eth)
            failure_summary += 'in output of command {}\n'.format(cmd)

    # Get all ospf routes
    cmd = "vtysh -c 'sh ip route ospf'"
    routes_out = execute_commands(module, cmd)

    # Verify configured cost in ospf routes
    for line in module.params['config_file'].splitlines():
        line = line.strip()
        if 'network' in line and 'area' in line and '192' not in line:
            config = line.split()
            network = config[1]

            for route in routes_out.splitlines():
                if network in route:
                    if '/{}'.format(cost) not in route:
                        RESULT_STATUS = False
                        failure_summary += 'On switch {} '.format(switch_name)
                        failure_summary += 'cost {} is not showing up '.format(cost)
                        failure_summary += 'for route {} '.format(route)
                        failure_summary += 'in output of command {}\n'.format(cmd)

    # Revert back the loopback ip
    cmd = 'ifconfig lo 127.0.0.1 netmask 255.0.0.0'
    execute_commands(module, cmd)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth_list=dict(required=False, type='str'),
            cost=dict(required=False, type='str'),
            config_file=dict(required=False, type='str', default=''),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_ospf_ecmp_basic(module)

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

