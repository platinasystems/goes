#!/usr/bin/python
""" Test Redis DB with Valid Values """

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
module: test_redis_valid
author: Platina Systems
short_description: Module to test redis db with valid values.
description:
    Module to perform different tests on redis db with valid hset values.
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
    log_dir_path:
      description:
        - Path to log directory where logs will be stored.
      required: False
      type: str
"""

EXAMPLES = """
- name: Test Redis db with valid values
  test_redis_valid:
    switch_name: "{{ inventory_hostname }}"
    switch_ip: "{{ ansible_ssh_host }}"
"""

RETURN = """
hash_name:
  description: Name of the hash in which to store the test result.
  returned: always
  type: str
start_time:
  description: Start time of test.
  returned: always
  type: str
end_time:
  description: End time of test.
  returned: always
  type: str
test_result:
  description: Passed/Failed.
  returned: always
  type: str
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
    Execute hset/hget command for the given parameter and verify the same.
    :param module: The Ansible module to fetch input parameters.
    :param operation: Name of the operation to perform: hget/hset.
    :param param: Name of the parameter.
    :param set_value: Value to set to the parameter.
    """
    global HASH_DICT, RESULT_STATUS

    cmd = '{} platina {} '.format(operation, param)

    if operation == 'hset':
        cmd += '{}'.format(set_value)

    cli = get_cli(module) + cmd
    out = run_cli(module, cli)

    # Store command prefixed with exec time as key and
    # command output as value in the hash dictionary
    exec_time = run_cli(module, 'date +%Y%m%d%T')
    key = '{}  {}'.format(exec_time, cmd)
    HASH_DICT[key] = out

    # For errors, update the result status to False
    if out is None or 'error' in out:
        RESULT_STATUS = False

    # Update the result status to False if set and get values do not match
    if operation == 'hget':
        if set_value != out:
            RESULT_STATUS = False


def test_hget_hset_operations(module):
    """
    Method to test basic hget and hset operations on redis.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS

    # Set vnet.pollInterval value to 2
    set_value = '2'
    parameter = 'vnet.pollInterval'
    execute_and_verify(module, 'hset', parameter, set_value)

    # Verify vnet.pollInterval value using hget command
    execute_and_verify(module, 'hget', parameter, set_value)

    # Set vnet.meth-2.speed value to auto
    set_value = 'auto'
    parameter = 'vnet.meth-2.speed'
    execute_and_verify(module, 'hset', parameter, set_value)

    # Verify vnet.meth-2.speed value using hget command
    execute_and_verify(module, 'hget', parameter, set_value)

    # Set vnet.fe1-pipe3-loopback.speed value to auto
    set_value = 'auto'
    parameter = 'vnet.fe1-pipe3-loopback.speed'
    execute_and_verify(module, 'hset', parameter, set_value)

    # Verify vnet.fe1-pipe3-loopback.speed value using hget command
    execute_and_verify(module, 'hget', parameter, set_value)


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            switch_ip=dict(required=False, type='str'),
            remote_access=dict(required=False, type='bool', default=False),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global RESULT_STATUS, HASH_DICT
    remote_access = module.params['remote_access']

    # Get the switch name
    switch_name = module.params['switch_name']

    # Get the start time
    start_time = run_cli(module, 'date +%Y%m%d%T')

    # Create a hash name
    test_name = 'redis_regression_03_remote' if remote_access else 'redis_regression_03'
    hash_name = switch_name + '-' + test_name + '-' + start_time

    # Perform and verify all required tests
    test_hget_hset_operations(module)

    # Get the end time
    end_time = run_cli(module, 'date +%Y%m%d%T')

    # Calculate the entire test result
    test_result = 'Passed' if RESULT_STATUS else 'Failed'

    if not remote_access:
        # Create a log file
        log_file_path = module.params['log_dir_path']
        log_file_path += '/regression03_' + start_time + '.log'
        log_file = open(log_file_path, 'w')
        for key, value in HASH_DICT.iteritems():
            log_file.write(key)
            log_file.write('\n')
            log_file.write(value)
            log_file.write('\n')
            log_file.write('\n')
    
        log_file.close()

    # Exit the module and return the required JSON.
    module.exit_json(
        hash_name=hash_name,
        start_time=start_time,
        end_time=end_time,
        test_result=test_result,
        hash_dict=HASH_DICT
    )

if __name__ == '__main__':
    main()

