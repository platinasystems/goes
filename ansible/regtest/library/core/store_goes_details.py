#!/usr/bin/python
""" Store GOES Version and Tag details in Redis """

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
module: store_goes_details
author: Platina Systems
short_description: Module to get and store GOES version and tag details.
description:
    Module to get GOES version and tag details and store test
    result in redis db using hash on server emulator.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    hash_name:
      description:
        - Name of the hash in which to store the result.
      required: False
      type: str
    version_details:
      description:
        - String describing different sha versions of goes
      required: False
      type: str
    tag_details:
      description:
        - String describing different tags of goes
      required: False
      type: str
"""

EXAMPLES = """
- name: Store GOES version and tag details in redis db
  store_goes_details:
    hash_name: "{{ valid_out.hash_name }}"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
    version_details: "{{ version_out.stdout }}"
    tag_details: "{{ tag_out.stdout }}"
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
    cli = get_cli()
    cli += 'hset {0} "{1}" "{2}"'.format(hash_name, key, value)
    run_cli(module, cli)


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            version_details=dict(required=False, type='str'),
            tag_details=dict(required=False, type='str'),
        )
    )

    switch_name = module.params['switch_name']
    hash_name = module.params['hash_name']
    version_details = module.params['version_details']
    tag_details = module.params['tag_details']

    version_details = version_details.splitlines()
    sha1 = []
    for version in version_details:
        sha1.append(version.split(': ')[1])

    key = '{}.version.go.sha1'.format(switch_name)
    store_in_hash(module, hash_name, key, sha1[0])

    key = '{}.version.fe1.sha1'.format(switch_name)
    store_in_hash(module, hash_name, key, sha1[1])

    key = '{}.version.firmware-fe1a.sha1'.format(switch_name)
    store_in_hash(module, hash_name, key, sha1[2])

    tag_details = tag_details.splitlines()
    tags = []
    for tag in tag_details:
        tags.append(tag.split(': ')[1])

    key = '{}.version.go.tag'.format(switch_name)
    store_in_hash(module, hash_name, key, tags[0])

    key = '{}.version.fe1.tag'.format(switch_name)
    store_in_hash(module, hash_name, key, tags[1])

    key = '{}.version.firmware-fe1a.tag'.format(switch_name)
    store_in_hash(module, hash_name, key, tags[2])

    out_msg = 'Stored the test result in hash: {}'.format(hash_name)

    # Exit the module and return the required JSON.
    module.exit_json(
        msg=out_msg
    )

if __name__ == '__main__':
    main()

