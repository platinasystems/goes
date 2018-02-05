#!/usr/bin/python
""" Add/Remove Advertised Route """

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
module: add_remove_route
author: Platina Systems
short_description: Module to add/remove advertised route.
description:
    Module to add/remove advertised route for BGP.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    remove:
      description:
        - Flag to indicate if we want to add a route or remove it.
      required: False
      type: bool
      default: False
"""

EXAMPLES = """
- name: Add advertised route
  add_remove_route:
    switch_name: "{{ inventory_hostname }}"
    remove: False

- name: Remove advertised route
  add_remove_route:
    switch_name: "{{ inventory_hostname }}"
    remove: True
"""

RETURN = """
hash_dict:
  description: Dictionary containing key value pairs to store in hash.
  returned: always
  type: dict
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
            remove=dict(required=False, type='bool', default=False),
        )
    )

    switch_name = module.params['switch_name']
    ip = '192.168.{}.1/32'.format(switch_name[-2::])

    if module.params['remove']:
        cmd = 'gobgp global rib -a ipv4 del {}'.format(ip)
        state = 'removed'
    else:
        cmd = 'gobgp global rib -a ipv4 add {}'.format(ip)
        state = 'added'

    run_cli(module, cmd)

    msg = 'On switch {}, {} route {}'.format(switch_name, state, ip)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=msg
    )


if __name__ == '__main__':
    main()

