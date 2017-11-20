#!/usr/bin/python
""" Assign Loopback IP """

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
module: assign_loopback
author: Platina Systems
short_description: Module to assign loopback ip or ip to eth interface.
description:
    Module to assign loopback ip or ip to eth interface given.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    eth:
      description:
        - Name of the eth interface to which ip need to be assigned.
      required: False
      type: str
    revert:
      description:
        - Flag to indicate if ip needs to reverted or not.
      required: False
      type: bool
      default: False
"""

EXAMPLES = """
- name: Assign loopback ip
  assign_loopback:
    switch_name: "{{ inventory_hostname }}"

- name: Assign ip to eth-11-1 interface
  assign_loopback:
    switch_name: "{{ inventory_hostname }}"
    eth: "11"
"""

RETURN = """
msg:
  description: String describing which loopback ip got assigned.
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
        return out.rstrip()
    elif err:
        return err.rstrip()
    else:
        return None


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            eth=dict(required=False, type='str', default=''),
            revert=dict(required=False, type='bool', default=False)
        )
    )

    switch_name = module.params['switch_name']
    eth = module.params['eth']
    switch_id = switch_name[-2::]

    if eth:
        if module.params['revert']:
            ip = '10.0.{}.{}'.format(eth, switch_id)
        else:
            ip = '192.168.{}.{}'.format(eth, switch_id)

        cmd = 'ifconfig eth-{}-1 {} netmask 255.255.255.0'.format(eth, ip)
        run_cli(module, cmd)
        msg = 'Assigned ip {} to eth-{}-1 interface on switch {}'.format(
            ip, eth, switch_name)
    else:
        ip = '192.168.{}.1'.format(switch_name[-2::])
        cmd = 'ifconfig lo {} netmask 255.255.255.0'.format(ip)
        run_cli(module, cmd)
        msg = 'Assigned loopback ip {} to {}'.format(ip, switch_name)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=msg
    )

if __name__ == '__main__':
    main()

