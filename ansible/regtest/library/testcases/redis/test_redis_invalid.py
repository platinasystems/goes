#!/usr/bin/python
""" Test Redis DB with Invalid Values """

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
# along with Ansible.  If not, see <http://www.gnu.org/licenses/>.
#

import shlex

from collections import OrderedDict

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: test_redis_invalid
author: Platina Systems
short_description: Module to test redis db with invalid values.
description:
    Module to perform different tests on redis db with invalid hset values.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    switch_ip:
      description:
        - IP of the switch on which tests will be performed.
      required: False
      type: str
    remote_access:
      description:
        - Specify if we want to access redis db remotely from server emulator.
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
- name: Test Redis db with invalid values
  test_redis_valid:
    switch_name: "{{ inventory_hostname }}"
"""

RETURN = """
hash_dict:
  description: Dictionary containing key value pairs to store in hash.
  returned: always
  type: dict
"""

RESULT_STATUS = True
HASH_DICT = OrderedDict()


def get_cli(module):
    """
    Method to get initial cli string.
    :param module: The Ansible module to fetch input parameters.
    :return: Initial cli string.
    """
    if module.params['remote_access']:
        cli = 'redis-cli -h {} -p 6379 '.format(module.params['switch_ip'])
    else:
        cli = 'redis-cli '

    return cli


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


def execute_and_verify(module, operation, param, set_value):
    """
    Execute hset command for the given parameter and verify the same.
    :param module: The Ansible module to fetch input parameters.
    :param operation: Name of the operation to perform: hget/hset.
    :param param: Name of the parameter.
    :param set_value: Value to set to the parameter.
    """
    global HASH_DICT, RESULT_STATUS
    failure_summary = ''
    switch_name = module.params['switch_name']

    cmd = '{} platina {} '.format(operation, param)

    if operation == 'hset':
        cmd += '{}'.format(set_value)

    cli = get_cli(module) + cmd
    out = run_cli(module, cli)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{0} {1} {2}'.format(switch_name, exec_time, cmd)
    HASH_DICT[key] = out

    # For errors, update the result status to False
    if out is None:
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'output of command {} is None\n'.format(cli)
    elif "error can't set" not in out.lower():
        # If expected error does not get thrown out,
        # update the result status to False
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'command {} should not have executed\n'.format(cli)
        
    return failure_summary


def test_hset_operations_with_invalid_input(module):
    """
    Method to test hset operations on redis with invalid input.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS
    failure_summary = ''

    # Set vnet.ready value to false and
    # verify if error msg gets thrown out
    set_value = 'false'
    parameter = 'vnet.ready'
    failure_summary += execute_and_verify(module, 'hset', parameter, set_value)

    # Set vnet.packet-generator.admin value to false and
    # verify if error msg gets thrown out
    set_value = 'false'
    parameter = 'vnet.packet-generator.admin'
    failure_summary += execute_and_verify(module, 'hset', parameter, set_value)

    HASH_DICT['result.detail'] = failure_summary


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            switch_ip=dict(required=False, type='str'),
            remote_access=dict(required=False, type='bool', default=False),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global RESULT_STATUS, HASH_DICT
    
    # Perform and verify all required tests
    test_hset_operations_with_invalid_input(module)
    
    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}.log'.format(module.params['hash_name'])

    if not module.params['remote_access']:
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

