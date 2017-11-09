#!/usr/bin/python
""" Test/Verify OSPF Load Balancing """

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
module: test_ospf_loadbalancing
author: Platina Systems
short_description: Module to test and verify ospf loadbalancing.
description:
    Module to test and verify ospf configurations and log the same.
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
- name: Verify ospf loadbalancing
  test_ospf_traffic:
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


def verify_ospf_load_balancing(module):
    """
    Method to verify ospf load balancing.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    routes_to_check = []
    loopback_ip = None

    switch_name = module.params['switch_name']
    leaf_list = module.params['leaf_list']
    config_file = module.params['config_file'].splitlines()
    is_leaf = True if switch_name in leaf_list else False

    # Get the current/running configurations
    execute_commands(module, "vtysh -c 'sh running-config'")

    # Update eth network interfaces
    for line in config_file:
        line = line.strip()
        if line.startswith('network'):
            switch_id = switch_name[-2::]
            address = line.split()[1]
            address = address.split('/')[0]
            octets = address.split('.')
            if octets[2] == switch_id and is_leaf:
                octets[3] = '1'
                ip = '.'.join(octets)
                loopback_ip = ip
                cmd = 'ifconfig lo {} netmask 255.255.255.0'.format(ip)
            else:
                octets[3] = switch_id
                ip = '.'.join(octets)
                routes_to_check.append(ip)
                eth = 'eth-{}-1'.format(octets[2])
                cmd = 'ifconfig {} {} netmask 255.255.255.0'.format(eth, ip)

            # Run ifconfig command
            execute_commands(module, cmd)

    # Restart and check Quagga status
    execute_commands(module, 'service quagga restart')
    time.sleep(35)
    execute_commands(module, 'service quagga status')

    if is_leaf:
        switch_ids = [leaf[-2::] for leaf in leaf_list]
        loopback_ip_octets = loopback_ip.split('.')
        switch_ids.remove(loopback_ip_octets[2])
        loopback_ip_octets[2] = switch_ids.pop()
        loopback_ip = '.'.join(loopback_ip_octets)

        # Get all ospf ip routes
        cmd = "vtysh -c 'sh ip route {}'".format(loopback_ip)
        ospf_routes = execute_commands(module, cmd)

        for route in routes_to_check:
            route_octets = route.split('.')
            route_octets.pop()
            route_ip = '.'.join(route_octets)
            if route_ip not in ospf_routes:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'output of command {} '.format(cmd)
                failure_summary += 'did not show correct routes\n'
                break

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            config_file=dict(required=False, type='str', default=''),
            leaf_list=dict(required=False, type='list', default=[]),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    verify_ospf_load_balancing(module)

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

