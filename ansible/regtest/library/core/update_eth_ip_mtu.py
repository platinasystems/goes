#!/usr/bin/python
""" Update eth IP and MTU """

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
module: update_eth_ip_mtu
author: Platina Systems
short_description: Module to update eth interface ip and mtu size.
description:
    Module to update eth interface ip and mtu size.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    config_file:
      description:
        - Config file describing eth ip and mtu details.
      required: False
      type: str
    revert:
      description:
        - Flag to indicate if we want to revert eth ip and mtu to default.
      required: False
      type: bool
      default: False
"""

EXAMPLES = """
- name: Update eth interface ip and mtu
  update_eth_ip_mtu:
    switch_name: "{{ inventory_hostname }}"
    revert: False
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
            config_file=dict(required=False, type='str', default=''),
            revert=dict(required=False, type='bool', default=False),
        )
    )

    switch_name = module.params['switch_name']
    config_file = module.params['config_file'].splitlines()
    msg = ''

    for line in config_file:
        if module.params['revert']:
            if 'mtu' in line:
                line = line.split()
                line[3] = '9216'
                line = ' '.join(line)
            elif 'netmask' in line:
                line = line.split()
                eth = line[1].split('-')[1]
                line[2] = '10.0.{}.{}'.format(eth, switch_name[-2::])
                line = ' '.join(line)

        run_cli(module, line)
        msg += "On switch {}, executed '{}'".format(switch_name, line)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=msg
    )

if __name__ == '__main__':
    main()

