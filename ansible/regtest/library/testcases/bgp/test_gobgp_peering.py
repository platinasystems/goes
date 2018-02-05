#!/usr/bin/python
""" Test/Verify GOBGP PEERING """

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
module: test_gobgp_peering
author: Platina Systems
short_description: Module to test and verify gobgp configurations.
description:
    Module to test and verify gobgp configurations and log the same.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    config_file:
      description:
        - GOBGP config which have been added.
      required: False
      type: str
    package_name:
      description:
        - Name of the package installed (e.g. gobgpd).
      required: False
      type: str
    check_ping:
      description:
        - Flag to indicate if ping should be tested or not.
      required: False
      type: bool
      default: False
    if_down:
      description:
        - Flag to indicate if interface down/up test case should be executed.
      required: False
      type: bool
      default: False
    eth_list:
      description:
        - Comma separated string of eth interfaces.
      required: False
      type: str
      default: ''
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
- name: Verify gobgp peering authentication
  test_gobgp_peering:
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


def verify_neighbor_relationship(module):
    """
    Method to verify if bgp neighbor relation is established or not.
    :param module: The Ansible module to fetch input parameters.
    :return: Failure summary if any
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    neighbor_count = 0
    switch_name = module.params['switch_name']
    check_ping = module.params['check_ping']
    config_file = module.params['config_file'].splitlines()
    self_ip = '192.168.{}.1'.format(switch_name[-2::])

    # Get gobgp neighbors
    cmd = 'gobgp nei'
    gobgp_out = execute_commands(module, cmd)

    if gobgp_out:
        for line in config_file:
            line = line.strip()
            if 'neighbor-address' in line:
                neighbor_count += 1
                neighbor_ip = line.split().pop()
                neighbor_ip = neighbor_ip.replace('"', '')
                if neighbor_ip not in gobgp_out:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'bgp neighbor {} '.format(neighbor_ip)
                    failure_summary += 'is not present in the output of '
                    failure_summary += 'command {}\n'.format(cmd)

                if check_ping:
                    packet_count = '3'
                    ping_cmd = 'ping -w 5 -c {} -I {} {}'.format(
                        packet_count, self_ip, neighbor_ip)
                    ping_out = execute_commands(module, ping_cmd)
                    if '{} received'.format(packet_count) not in ping_out:
                        RESULT_STATUS = False
                        failure_summary += 'From switch {} '.format(switch_name)
                        failure_summary += 'neighbor ip {} '.format(neighbor_ip)
                        failure_summary += 'is not getting pinged\n'

            if 'peer-as' in line:
                remote_as = line.split().pop()
                if remote_as not in gobgp_out:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'remote-as {} '.format(remote_as)
                    failure_summary += 'is not present in the output of '
                    failure_summary += 'command {}\n'.format(cmd)

        gobgp_out = gobgp_out.lower()
        if gobgp_out.count('establ') != neighbor_count:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'bgp state is not established for neighbors\n'
    else:
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'bgp neighbors cannot be verified since '
        failure_summary += 'output of command {} is None'.format(cmd)

    return failure_summary


def verify_gobgp_peering(module):
    """
    Method to verify gobgp peering.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    package_name = module.params['package_name']
    check_ping = module.params['check_ping']
    if_down = module.params['if_down']
    eth_list = module.params['eth_list']
    leaf_list = module.params['leaf_list']

    if eth_list:
        eth_list = eth_list.split(',')

    # Get the gobgp config
    execute_commands(module, 'cat /etc/gobgp/gobgpd.conf')

    # Restart and check package status
    execute_commands(module, 'service {} restart'.format(package_name))
    execute_commands(module, 'service {} status'.format(package_name))

    # Advertise the routes
    if check_ping or if_down:
        add_route_cmd = 'gobgp global rib -a ipv4 add 192.168.{}.1/32'.format(
            switch_name[-2::])
        execute_commands(module, add_route_cmd)
        time.sleep(2)

    # Verify bgp neighbor relationship
    failure_summary += verify_neighbor_relationship(module)

    if if_down:
        if switch_name in leaf_list:
            # Bring down the interfaces
            for eth in eth_list:
                down_cmd = 'ifconfig eth-{}-1 down'.format(eth)
                execute_commands(module, down_cmd)

        # Wait for 5 secs
        time.sleep(5)

        # Verify bgp neighbor relationship
        failure_summary += verify_neighbor_relationship(module)

        if switch_name in leaf_list:
            # Bring up the interfaces
            for eth in eth_list:
                down_cmd = 'ifconfig eth-{}-1 up'.format(eth)
                execute_commands(module, down_cmd)

        # Wait for 5 secs
        time.sleep(5)

        # Verify bgp neighbor relationship
        failure_summary += verify_neighbor_relationship(module)

    # Store the failure summary in hash
    HASH_DICT['result.detail'] = failure_summary

    # # Delete advertised routes
    if check_ping or if_down:
        time.sleep(7)
        cmd = 'gobgp global rib -a ipv4 del 192.168.{}.1/32'.format(
            switch_name[-2::])
        execute_commands(module, cmd)

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            config_file=dict(required=False, type='str'),
            package_name=dict(required=False, type='str'),
            check_ping=dict(required=False, type='bool', default=False),
            if_down=dict(required=False, type='bool', default=False),
            eth_list=dict(required=False, type='str', defaut=''),
            leaf_list=dict(required=False, type='list', default=[]),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_gobgp_peering(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}.log'.format(module.params['hash_name'])
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
        hash_dict=HASH_DICT,
        log_file_path=log_file_path
    )

if __name__ == '__main__':
    main()

