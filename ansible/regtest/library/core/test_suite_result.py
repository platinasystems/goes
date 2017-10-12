#!/usr/bin/python
""" Test Suite Result """

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

from ansible.module_utils.basic import AnsibleModule

DOCUMENTATION = """
---
module: test_suite_result
author: Platina Systems
short_description: Module to report test suite result.
description: Module to report test suite result.
options:
    result_data:
      description: String containing names of failed test
                   hashes parsed from txt file.
      required: True
      type: str
"""

EXAMPLES = """
- name: Report test suite result
  test_suite_result:
    result_data: "{{ lookup('file', '{{ txt_file }}') }}"
"""

RETURN = """
test_suite_result:
  description: Dictionary containing status and hash names of failed tests
  returned: always
  type: dict
"""


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            result_data=dict(required=True, type='str'),
        )
    )

    result_data = module.params['result_data']
    output = {}

    if result_data:
        output['Result'] = 'Failed'
        output['Hash names of failed tests'] = result_data.split('\n')
    else:
        output['Result'] = 'Passed'

    # Exit the module and return the required JSON.
    module.exit_json(
        test_suite_status=output,
    )

if __name__ == '__main__':
    main()

