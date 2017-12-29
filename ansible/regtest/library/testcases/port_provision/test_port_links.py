#!/usr/bin/python
""" Test/Verify Port Links """

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
module: test_port_links
author: Platina Systems
short_description: Module to execute and verify port links.
description:
    Module to execute and verify port links.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    eth_list:
      description:
        - List of eth interfaces described as string.
      required: False
      type: str
      default: ''
    speed:
      description:
        - Speed of the eth interface port.
      required: False
      type: str
    media:
      description:
        - Media of the eth interface port.
      required: False
      type: str
    fec:
      description:
        - Fec of the eth interface port.
      required: False
      type: str
    leaf_server:
      description:
        - Name of the leaf switch on which iperf server is running.
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
- name: Execute and verify port links
  test_port_links:
    switch_name: "{{ inventory_hostname }}"
    eth_list: "2,4,6,8,10,12,14,16"
    speed: "100g"
    media: "copper"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
    log_dir_path: "{{ port_provision_log_dir }}"
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


def verify_port_links(module):
    """
    Method to execute and verify port links.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    speed = module.params['speed']
    media = module.params['media']
    fec = module.params['fec']
    leaf_server = module.params['leaf_server']
    eth_list = module.params['eth_list'].split(',')

    if speed == '100g' or speed == 'auto' or speed == '40g':
        for eth in eth_list:
            # Verify interface media is set to correct value
            cmd = 'goes hget platina vnet.eth-{}-1.media'.format(eth)
            out = execute_commands(module, cmd)
            if media not in out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'interface media is not set to copper '
                failure_summary += 'for the interface eth-{}-1\n'.format(eth)

            # Verify speed of interfaces are set to correct value
            cmd = 'goes hget platina vnet.eth-{}-1.speed'.format(eth)
            out = execute_commands(module, cmd)
            if speed not in out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'speed of the interface '
                failure_summary += 'is not set to {} for '.format(speed)
                failure_summary += 'the interface eth-{}-1\n'.format(eth)

            # Verify fec of interfaces are set to correct value
            cmd = 'goes hget platina vnet.eth-{}-1.fec'.format(eth)
            out = execute_commands(module, cmd)
            if fec not in out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'fec is not set to {} for '.format(fec)
                failure_summary += 'the interface eth-{}-1\n'.format(eth)

            # Bring up the interfaces
            cmd = 'ifconfig eth-{}-1 up'.format(eth)
            execute_commands(module, cmd)

            # Verify if port links are up for eth
            cmd = 'goes hget platina vnet.eth-{}-1.link'.format(eth)
            out = execute_commands(module, cmd)
            if 'true' not in out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'port link is not up '
                failure_summary += 'for the interface eth-{}-1\n'.format(eth)
    elif speed == '10g' or speed == 'auto10g' or speed == '25g':
        third_octet = 0
        if switch_name == leaf_server:
            last_octet = '1'
        else:
            last_octet = '2'

        for eth in eth_list:
            for subport in range(1, 5):
                # Verify interface media is set to correct value
                cmd = 'goes hget platina vnet.eth-{}-{}.media'.format(
                    eth, subport)
                out = execute_commands(module, cmd)
                if media not in out:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'interface media is not set to copper '
                    failure_summary += 'for the interface eth-{}-{}\n'.format(
                        eth, subport)

                # Verify speed of interfaces are set to correct value
                cmd = 'goes hget platina vnet.eth-{}-{}.speed'.format(
                    eth, subport)
                out = execute_commands(module, cmd)
                if out not in speed:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'speed of the interface '
                    failure_summary += 'is not set to {} for '.format(speed)
                    failure_summary += 'the interface eth-{}-{}\n'.format(
                        eth, subport)

                # Verify fec of interfaces are set to correct value
                cmd = 'goes hget platina vnet.eth-{}-{}.fec'.format(
                    eth, subport)
                out = execute_commands(module, cmd)
                if fec not in out:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'fec is not set to {} for '.format(fec)
                    failure_summary += 'the interface eth-{}-{}\n'.format(
                        eth, subport)

        for eth in eth_list:
            for subport in range(1, 5):
                third_octet += 1
                # Assign ip address to these interfaces
                cmd = 'ifconfig eth-{}-{} 192.168.{}.{} '.format(
                    eth, subport, third_octet, last_octet)
                cmd += 'netmask 255.255.255.0'
                execute_commands(module, cmd)

        for eth in eth_list:
            for subport in range(1, 5):
                # Bring up the interfaces
                cmd = 'ifconfig eth-{}-{} up'.format(eth, subport)
                execute_commands(module, cmd)

        time.sleep(20)
        for eth in eth_list:
            for subport in range(1, 5):
                # Verify if port links are up
                cmd = 'goes hget platina vnet.eth-{}-{}.link'.format(
                    eth, subport)
                out = execute_commands(module, cmd)
                if 'true' not in out:
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'port link is not up '
                    failure_summary += 'for the interface eth-{}-{}\n'.format(
                        eth, subport)

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth_list=dict(required=False, type='str', default=''),
            speed=dict(required=False, type='str'),
            media=dict(required=False, type='str'),
            fec=dict(required=False, type='str', default=''),
            leaf_server=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify port link
    verify_port_links(module)

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

