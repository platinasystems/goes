#!/usr/bin/python
""" Store test result in redis db """

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
module: store_result_in_redis
author: Platina Systems
short_description: Module to store test result in redis db.
description:
    Module to store test result in redis db using hash on server emulator.
options:
    hash_name:
      description:
        - Name of the hash in which to store the result.
      required: False
      type: str
    start_time:
      description:
        - Time at which test started executing.
      required: False
      type: str
    end_time:
      description:
        - Time at which test execution ended.
      required: False
      type: str
    hash_dict:
      description:
        - Dict containing key value pairs to store in hash.
      required: False
      type: dict
"""

EXAMPLES = """
- name: Store test result in redis db
  store_result_in_redis:
   hash_name: "{{ valid_out.hash_name }}"
   start_time: "{{ valid_out.start_time }}"
   end_time: "{{ valid_out.end_time }}"
   hash_dict: "{{ valid_out.hash_dict }}"
"""

RETURN = """
msg:
  description: String describing that test result got stored in hash.
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


def store_in_hash(module, hash_name, key, value):
    """
    Method to store test result in a hash on server emulator redis db
    :param module: The Ansible module to fetch input parameters.
    :param hash_name: Name of the hash.
    :param key: Key name in the hash.
    :param value: Value for the key.
    """
    if (key == 'result.detail' or key == 'result.status' or
            key == 'result.raw'):
        cli = get_cli()
        cli += 'hget {0} {1}'.format(hash_name, key)
        existing_value = run_cli(module, cli)

        value = existing_value + value
        if key == 'result.status':
            value = 'Failed' if 'Failed' in value else 'Passed'

        cli = get_cli()
        cli += 'hset {0} "{1}" "{2}"'.format(hash_name, key, value)
        run_cli(module, cli)
    elif '.time' in key:
        cli = get_cli()
        cli += 'hset {0} "{1}" "{2}"'.format(hash_name, key, value)
        run_cli(module, cli)


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            hash_name=dict(required=False, type='str'),
            start_time=dict(required=False, type='str'),
            end_time=dict(required=False, type='str'),
            log_content=dict(required=False, type='str'),
            hash_dict=dict(required=False, type='dict'),
        )
    )

    hash_name = module.params['hash_name']

    # Store start time of test run in hash
    start_time = module.params['start_time']
    store_in_hash(module, hash_name, 'start.time', start_time)

    # Store end time of test run in the hash
    end_time = module.params['end_time']
    store_in_hash(module, hash_name, 'end.time', end_time)

    # Store entire long content in hash
    log_content = module.params['log_content']
    store_in_hash(module, hash_name, 'result.raw', log_content)

    # Store key value pairs in the hash
    for key, value in module.params['hash_dict'].iteritems():
        store_in_hash(module, hash_name, key, value)

    out_msg = 'Stored the test result in hash: {}'.format(hash_name)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=out_msg
    )

if __name__ == '__main__':
    main()

