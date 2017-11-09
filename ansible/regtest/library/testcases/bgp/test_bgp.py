#!/usr/bin/python
""" Test/Verify BGP Config """

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
module: test_bgp
author: Platina Systems
short_description: Module to test and verify bgp configurations.
description:
    Module to test and verify bgp configurations and log the same.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    testcase_name:
      description:
        - Name of the test case.
      required: False
      type: str
    check_neighbors:
      description:
        - Boolean value to determine which command to execute.
      required: False
      type: bool
      default: False
    neighbor1_ip:
      description:
        - IP of 1st bgp neighbor.
      required: False
      type: str
    neighbor1_as:
      description:
        - AS number of 1st bgp neighbor.
      required: False
      type: str
    neighbor2_ip:
      description:
        - IP of 2nd bgp neighbor.
      required: False
      type: str
    neighbor2_as:
      description:
        - AS number of 2nd bgp neighbor.
      required: False
      type: str
    network1:
      description:
        - Network address of 1st switch.
      required: False
      type: str
    network2:
      description:
        - Network address of 2nd switch.
      required: False
      type: str
    network3:
      description:
        - Network address of 3rd switch.
      required: False
      type: str
    network4:
      description:
        - Network address of 4th switch.
      required: False
      type: str
    log_dir_path:
      description:
        - Path to log directory where logs will be stored.
      required: False
      type: str
"""

EXAMPLES = """
- name: Verify bgp config
  test_bgp:
    switch_name: "{{ inventory_hostname }}"
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
    key = '{}  {}'.format(exec_time, cmd)
    HASH_DICT[key] = out

    return out


def verify_bgp_neighbors(module):
    """
    Method to verify if bgp neighbor relationship got established or not.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    if module.params['check_neighbors']:
        # Get bgp neighbors relationships
        out = execute_commands(module, "vtysh -c 'sh ip bgp neighbors'")

        # For errors, update the result status to False
        if out is None or 'error' in out:
            RESULT_STATUS = False

        neighbor1_ip = module.params['neighbor1_ip']
        neighbor1_as = module.params['neighbor1_as']
        neighbor2_ip = module.params['neighbor2_ip']
        neighbor2_as = module.params['neighbor2_as']

        if (neighbor1_ip not in out or neighbor1_as not in out or
                neighbor2_ip not in out or neighbor2_as not in out or
                'BGP state = Established' not in out):
            RESULT_STATUS = False
    else:
        # Get bgp config
        out = execute_commands(module, "vtysh -c 'sh ip bgp'")

        # For errors, update the result status to False
        if out is None or 'error' in out:
            RESULT_STATUS = False

        network1 = module.params['network1']
        network2 = module.params['network2']
        network3 = module.params['network3']
        network4 = module.params['network4']

        number_of_prefixes = 0

        if network1:
            number_of_prefixes += 1
            if network1 not in out:
                RESULT_STATUS = False

        if network2:
            number_of_prefixes += 1
            if network2 not in out:
                RESULT_STATUS = False

        if network3:
            number_of_prefixes += 1
            if network3 not in out:
                RESULT_STATUS = False

        if network4:
            number_of_prefixes += 1
            if network4 not in out:
                RESULT_STATUS = False

        if 'Total number of prefixes {}'.format(number_of_prefixes) not in out:
            RESULT_STATUS = False


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            testcase_name=dict(required=False, type='str'),
            check_neighbors=dict(required=False, type='bool', default=False),
            neighbor1_ip=dict(required=False, type='str'),
            neighbor1_as=dict(required=False, type='str'),
            neighbor2_ip=dict(required=False, type='str'),
            neighbor2_as=dict(required=False, type='str'),
            network1=dict(required=False, type='str', default=''),
            network2=dict(required=False, type='str', default=''),
            network3=dict(required=False, type='str', default=''),
            network4=dict(required=False, type='str', default=''),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    # Get the switch name
    switch_name = module.params['switch_name']

    # Get the testcase name
    testcase_name = module.params['testcase_name']

    # Get the start time
    start_time = run_cli(module, 'date +%Y%m%d%T')

    # Create a hash name
    hash_name = switch_name + '-' + testcase_name + '-' + start_time

    # Verify bgp neighbors
    verify_bgp_neighbors(module)

    # Get the end time
    end_time = run_cli(module, 'date +%Y%m%d%T')

    # Calculate the entire test result
    test_result = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}_'.format(testcase_name) + start_time + '.log'
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

