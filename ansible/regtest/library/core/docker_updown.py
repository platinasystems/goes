#!/usr/bin/python
""" Docker Up Down """

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

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: docker_updown
author: Platina Systems
short_description: Module to bring up and bring down docker containers.
description:
    Module to bring up and bring down docker containers.
options:
    config_file:
      description:
        - Config details of docker container.
      required: False
      type: str
    state:
      description:
        - String describing if docker container has to be brought up or down.
      required: False
      type: str
      choices: ['up', 'down']
"""

EXAMPLES = """
- name: Bring up docker container
  docker_updown:
    config_file: "{{ lookup('file', '../../group_vars/{{ inventory_hostname }}/{{ item }}') }}"
    state: 'up'
"""

RETURN = """
msg:
  description: String describing docker container state.
  returned: always
  type: str
"""


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
        return out
    elif err:
        return err
    else:
        return None


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            config_file=dict(required=False, type='str'),
            state=dict(required=False, type='str', choices=['up', 'down']),
        )
    )

    d_move = '~/./docker_move.sh'
    container_name = None
    eth_list = []
    config_file = module.params['config_file'].splitlines()

    for line in config_file:
        if 'container_name' in line:
            container_name = line.split()[1]
        elif 'interface' in line:
            eth = line.split()[1]
            eth_list.append(eth)

    container_id = container_name[1::]
    dummy_id = int(container_id) - 1

    if module.params['state'] == 'up':
        # Add dummy interface and bring it up
        cmd = 'ip link add dummy{} type dummy 2> /dev/null'.format(dummy_id)
        run_cli(module, cmd)

        # Bring up dummy interface
        cmd = '{} up {} dummy{} 192.168.{}.1/32'.format(d_move, container_name,
                                                        dummy_id, container_id)
        run_cli(module, cmd)

        # Bring up given interfaces in the docker container
        for eth in eth_list:
            cmd = '{} up {} eth-{}-1 10.0.{}.32/24'.format(d_move, container_name,
                                                           eth, eth)
            run_cli(module, cmd)
    else:
        # Bring down all interfaces in the docker container
        for eth in eth_list:
            cmd = '{} down {} eth-{}-1'.format(d_move, container_name, eth)
            run_cli(module, cmd)

        # Bring down dummy interface
        cmd = '{} down {} dummy{}'.format(d_move, container_name, dummy_id)
        run_cli(module, cmd)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg='Module executed successfully'
    )

if __name__ == '__main__':
    main()

