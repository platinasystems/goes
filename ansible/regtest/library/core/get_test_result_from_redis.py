#!/usr/bin/python
""" Get test result from redis db """

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
# along with Ansible.  If not, see <http://www.gnu.org/licenses/>.
#

import shlex

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: get_test_result_from_redis
author: Platina Systems
short_description: Module to get test result from redis db.
description:
    Module to get test result from redis db using hash on server emulator.
options:
    hash_name:
      description:
        - Name of the hash in which to store the result.
      required: False
      type: str
"""

EXAMPLES = """
- name: Get test result from redis db
  store_result_in_redis:
   hash_name: "{{ valid_out.hash_name }}"
"""

RETURN = """
result_status:
  description: Passed/Failed.
  returned: always
  type: str
result_detail:
  description: String describing details of the test result.
  returned: always
  type: str
"""


def get_cli():
    """
    Method to get the initial cli string.
    :return: Initial cli string.
    """
    return 'redis-cli -p 9090 '


def run_cli(module, cli):
    """
    Method to execute the cli command on the target node(s) and returns the
    output.
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
            hash_name=dict(required=False, type='str'),
        )
    )

    hash_name = module.params['hash_name']

    cli = get_cli()
    cli += 'hget {0} {1}'.format(hash_name, 'result.status')
    out = run_cli(module, cli)
    status = 'Failed' if 'Failed' in out else 'Passed'

    cli = get_cli()
    cli += 'hget {0} {1}'.format(hash_name, 'result.detail')
    out = run_cli(module, cli)
    detail = 'None' if out.isspace() else out

    # Exit the module and return the required JSON.
    module.exit_json(
        result_status=status,
        result_detail=detail
    )

if __name__ == '__main__':
    main()

