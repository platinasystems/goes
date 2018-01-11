#!/usr/bin/python
""" Test/Verify Single Port Provisioning on Packet Generator """

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
module: test_single_port_provision_packet_generator
author: Platina Systems
short_description: Module to execute and verify port provisioning on packet
generator.
description:
    Module to execute and verify port provisioning on packet generator.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    ce:
      description:
        - Name of ce ports described as string.
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
    create_cint_file:
      description:
        - Name of create cint file.
      required: False
      type: str
    delete_cint_file:
      description:
        - Name of delete cint file.
      required: False
      type: str
    reset_config:
      description:
        - Flag to indicate if config needs to be reset.
      required: False
      type: bool
      default: False
    hash_name:
      description:
        - Name of the hash in which to store the result in redis.
      required: False
      type: str
"""

EXAMPLES = """
- name: Execute and verify port provisioning
  test_port_provision_packet_generator:
    switch_name: "{{ inventory_hostname }}"
    ce: "27"
    speed: "100g"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
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
    create_cint_file = module.params['create_cint_file']
    delete_cint_file = module.params['delete_cint_file']
    ce = module.params['ce']
    initial_cli = "python /home/platina/bin/bcm.py"

    if not module.params['reset_config']:
        if speed == '100g':
            # Install optic media and set auto negotiation mode to no
            cmd = '{} port ce{} speed=100000 an=0 if={}'.format(
                initial_cli, ce, media.lower())
            execute_commands(module, cmd)

        # Verify if port is up
        cmd = "{} 'ps ce{}'".format(initial_cli, ce)
        ps_out = execute_commands(module, cmd)
        ps_out = ps_out.lower()

        if 'up' not in ps_out:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'ce{} port is not up\n'.format(ce)

        # Execute cint configuration script
        cmd = "{} 'cint {}'".format(initial_cli, create_cint_file)
        execute_commands(module, cmd)

        # Generate traffic
        cmd = 'python /home/platina/bin/autoTest_pktDrop_by_pktSizes.py '
        cmd += '30 31 1 10'
        execute_commands(module, cmd)
    else:
        cmd = "{} 'port ce{} en=f'".format(initial_cli, ce)
        execute_commands(module, cmd)

        cmd = "{} 'port ce{} lanes 4'".format(initial_cli, ce)
        execute_commands(module, cmd)

        cmd = "{} 'port ce{} speed=100000 an=0 if=cr4'".format(initial_cli, ce)
        execute_commands(module, cmd)

        # Delete cint configuration
        cmd = "{} 'cint {}'".format(initial_cli, delete_cint_file)
        execute_commands(module, cmd)

        cmd = "{} 'port ce{} ena=t'".format(initial_cli, ce)
        execute_commands(module, cmd)

    HASH_DICT['result.detail'] = failure_summary


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            ce=dict(required=False, type='str', default=''),
            speed=dict(required=False, type='str'),
            media=dict(required=False, type='str'),
            fec=dict(required=False, type='str', default=''),
            create_cint_file=dict(required=False, type='str'),
            delete_cint_file=dict(required=False, type='str'),
            reset_config=dict(required=False, type='bool', default=False),
            hash_name=dict(required=False, type='str')
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify port_provisioning
    verify_single_port_provisioning(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Exit the module and return the required JSON.
    module.exit_json(
        hash_dict=HASH_DICT
    )

if __name__ == '__main__':
    main()

