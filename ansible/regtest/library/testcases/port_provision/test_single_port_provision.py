#!/usr/bin/python
""" Test/Verify Single Port Provisioning """

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
module: test_single_port_provision
author: Platina Systems
short_description: Module to execute and verify port provisioning.
description:
    Module to execute and verify port provisioning.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    eth:
      description:
        - Single eth interface.
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
    verify_links:
      description:
        - Flag to indicate if port links need to be verified.
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
- name: Execute and verify port provisioning
  test_port_provision:
    switch_name: "{{ inventory_hostname }}"
    eth: "28"
    speed: "100g"
    media: "copper"
    fec: "none"
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


def verify_single_port_provisioning(module):
    """
    Method to execute and verify port provisioning.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    speed = module.params['speed']
    media = module.params['media']
    fec = module.params['fec']
    verify_links = module.params['verify_links']
    eth = module.params['eth']

    if speed == '100g':
        if verify_links:
            # Verify if port links are up for eth
            cmd = 'goes hget platina vnet.eth-{}-1.link'.format(eth)
            link_out = execute_commands(module, cmd)
            if 'true' not in link_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'port link is not up for '
                failure_summary += 'eth-{}-1 interface\n'.format(eth)
        else:
            # Verify optic media
            cmd = 'goes hget platina qsfp.compliance'
            optic_out = execute_commands(module, cmd)

            verify_str = 'port-{}.qsfp.compliance: extended '.format(eth)
            if media == 'CWDM4':
                verify_str += '100G CWDM4'
            elif media == 'SR4':
                verify_str += '100GBASE-SR4'

            if verify_str not in optic_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'eth-{}-1 interface does not '.format(eth)
                failure_summary += 'have correct optic installed\n'

            # Verify fec is set to correct value
            cmd = 'goes hget platina vnet.eth-{}-1.fec'.format(eth)
            fec_out = execute_commands(module, cmd)

            if fec not in fec_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'fec is not set to {} '.format(fec)
                failure_summary += 'for eth-{}-1 interface\n'.format(eth)

            # Verify speed of interfaces are set to correct value
            cmd = 'goes hget platina vnet.eth-{}-1.speed'.format(eth)
            speed_out = execute_commands(module, cmd)
            if speed not in speed_out:
                RESULT_STATUS = False
                failure_summary += 'On switch {} '.format(switch_name)
                failure_summary += 'speed of the interface '
                failure_summary += 'is not set to {} for '.format(speed)
                failure_summary += 'the eth-{}-1 interface\n'.format(eth)

            # Bring up the interface
            execute_commands(module, 'ifconfig eth-{}-1 up'.format(eth))

    HASH_DICT['result.detail'] = failure_summary

    # Get the GOES status info
    execute_commands(module, 'goes status')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth=dict(required=False, type='str', default=''),
            speed=dict(required=False, type='str'),
            media=dict(required=False, type='str'),
            fec=dict(required=False, type='str', default=''),
            verify_links=dict(required=False, type='bool', default=False),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify single port_provisioning
    verify_single_port_provisioning(module)

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

