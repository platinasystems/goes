#!/usr/bin/python
""" Add Dummy Interface """

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
module: add_dummy_interface
author: Platina Systems
short_description: Module to add dummy interface.
description:
    Module to add dummy interface.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
"""

EXAMPLES = """
- name: Add dummy interface
  add_dummy_interface:
    switch_name: "{{ inventory_hostname }}"
"""

RETURN = """
msg:
  description: String describing which dummy interface ip got assigned.
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
        )
    )

    # Get the switch name
    switch_name = module.params['switch_name']

    # Add dummy0 interface
    run_cli(module, 'ip link add dummy0 type dummy')

    # Assign ip to this created dummy0 interface
    ip = '192.168.{}.1'.format(switch_name[-2::])
    cmd = 'ifconfig dummy0 {} netmask 255.255.255.0'.format(ip)
    run_cli(module, cmd)
    
    msg = 'Added dummy0 interface with ip {} to {}'.format(ip, switch_name)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=msg
    )

if __name__ == '__main__':
    main()

