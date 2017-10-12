#!/usr/bin/python
""" Test BMC Processor's Redis DB """

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

from ansible.module_utils.basic import AnsibleModule


DOCUMENTATION = """
---
module: test_bmc_redis
author: Platina Systems
short_description: Module to test redis db on BMC processor.
description:
    Module to perform different tests on redis db running on BMC processor.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    bmc_redis_ip:
      description:
        - IP to access BMC processor's redis db.
      required: False
      type: str
"""

EXAMPLES = """
- name: Test BMC Redis db
  test_redis_valid:
    switch_name: "{{ inventory_hostname }}"
    bmc_redis_ip: "{{ ansible_ssh_host }}"
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
HASH_DICT = {}


def get_cli(module):
    """
    Method to get initial cli string.
    :param module: The Ansible module to fetch input parameters.
    :return: Initial cli string.
    """
    return 'redis-cli -h {} '.format(module.params['bmc_redis_ip'])


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


def get_ipv6_address(module):
    """
    Method to get ipv6 address from base mac address
    :param module: The Ansible module to fetch input parameters.
    :return: IPV6 local address as string
    """
    cli = get_cli(module)
    cli += 'hget platina eeprom.BaseEthernetAddress'
    mac = run_cli(module, cli)

    parts = mac.split(":")
    parts.insert(3, "ff")
    parts.insert(4, "fe")
    parts[0] = "%x" % (int(parts[0], 16) ^ 2)

    ipv6parts = []
    for i in range(0, len(parts), 2):
        ipv6parts.append("".join(parts[i:i + 2]))

    ipv6 = "fe80::%s" % (":".join(ipv6parts))

    return ipv6


def execute_and_verify(module, operation, param):
    """
    Execute hget command for the given parameter and verify the same.
    :param module: The Ansible module to fetch input parameters.
    :param operation: Name of the operation to perform: hget/hset.
    :param param: Name of the parameter.
    """
    global HASH_DICT, RESULT_STATUS

    cmd = '{} platina {} '.format(operation, param)

    cli = get_cli(module) + cmd
    out = run_cli(module, cli)

    # Store command as key and command output as value in the hash dictionary
    HASH_DICT[cmd] = out

    # For errors, update the result status to False
    if out is None or 'error' in out:
        RESULT_STATUS = False


def test_hget_operations(module):
    """
    Method to test basic hget operations on bmc redis db.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS

    ipv6 = get_ipv6_address(module)
    ipv6 += '%eth0'

    parameter = 'temp {}'.format(ipv6)
    execute_and_verify(module, 'hget', parameter)

    parameter = 'status {}'.format(ipv6)
    execute_and_verify(module, 'hget', parameter)

    parameter = 'fan_tray {}'.format(ipv6)
    execute_and_verify(module, 'hget', parameter)

    parameter = 'psu {}'.format(ipv6)
    execute_and_verify(module, 'hget', parameter)

    parameter = 'vmon {}'.format(ipv6)
    execute_and_verify(module, 'hget', parameter)


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            bmc_redis_ip=dict(required=False, type='str'),
        )
    )

    global RESULT_STATUS, HASH_DICT

    # Get the switch name
    switch_name = module.params['switch_name']

    # Get the start time
    start_time = run_cli(module, 'date +%Y%m%d%T')

    # Create a hash name
    hash_name = switch_name + '-' + 'test_bmc_redis_db' + '-' + start_time

    # Perform and verify all required tests
    test_hget_operations(module)

    # Get the end time
    end_time = run_cli(module, 'date +%Y%m%d%T')

    # Calculate the entire test result
    test_result = 'Passed' if RESULT_STATUS else 'Failed'

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

