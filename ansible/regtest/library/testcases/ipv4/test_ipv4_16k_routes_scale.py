#!/usr/bin/python
""" Test/Verify IPV4 16k Routes Scale """

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

import mmap
import shlex

from collections import OrderedDict

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: test_ipv4_16k_routes_scale
author: Platina Systems
short_description: Module to test and verify ipv4 16k routes scale.
description:
    Module to test and verify ipv4 routes and log the same.
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
- name: Verify ipv4 routes scale
  test_ipv4_16k_routes_scale:
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

    if 'route -n' not in cmd:
        HASH_DICT[key] = out

    return out


def verify_ipv4_routes_scale(module):
    """
    Method to verify ipv4 routes scale.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    switch_name = module.params['switch_name']
    spine_list = module.params['spine_list']
    leaf_list = module.params['leaf_list']
    failure_summary = ''
    total_routes, routes_count = 0, 0

    is_spine = True if switch_name in spine_list else False

    if is_spine:
        if spine_list.index(switch_name) == 0:
            octet = 2
        else:
            octet = 4
    else:
        if leaf_list.index(switch_name) == 0:
            octet = 1
        else:
            octet = 3

    linux_routes = open('/var/log/linux_routes.txt')
    quagga_routes = open('/var/log/quagga_routes.txt')
    lrs = mmap.mmap(linux_routes.fileno(), 0, access=mmap.ACCESS_READ)
    qrs = mmap.mmap(quagga_routes.fileno(), 0, access=mmap.ACCESS_READ)

    for i in range(1, 100):
        if total_routes < 16000:
            for j in range(1, 256):
                if total_routes < 16000 and routes_count < 1000:
                    route = '{}.{}.{}.{}'.format(octet, octet, i, j)
                    if lrs.find(route) != -1 or qrs.find(route) != -1:
                        pass
                    else:
                        RESULT_STATUS = False
                        failure_summary += 'On switch {} '.format(switch_name)
                        failure_summary += '{} route is not present\n'.format(route)

                    routes_count += 1
                    total_routes += 1
                else:
                    routes_count = 0
                    break
        else:
            break

    linux_routes.close()
    quagga_routes.close()

    # Get the GOES status info
    goes_status = execute_commands(module, 'goes status')

    if 'not ok' in goes_status.lower():
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'GOES status is NOT OK\n'

    HASH_DICT['result.detail'] = failure_summary


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            spine_list=dict(required=False, type='list', default=[]),
            leaf_list=dict(required=False, type='list', default=[]),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_ipv4_routes_scale(module)

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

