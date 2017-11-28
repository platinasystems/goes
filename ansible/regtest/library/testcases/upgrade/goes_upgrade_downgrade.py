#!/usr/bin/python
""" GOES Upgrade/Downgrade """

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
module: goes_upgrade_downgrade
author: Platina Systems
short_description: Module to upgrade and downgrade goes.
description:
    Module to upgrade and downgrade goes version.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    installer_dir:
      description:
        - Directory path where GOES upgrade installer file is stored.
      required: False
      type: str
    upgrade_installer_name:
      description:
        - Name of the GOES upgrade installer file.
      required: False
      type: str
    downgrade_installer_name:
      description:
        - Name of the GOES downgrade installer file.
      required: False
      type: str
    upgrade_version:
      description:
        - String describing goes upgrade version details.
      required: False
      type: str
    downgrade_version:
      description:
        - String describing goes downgrade version details.
      required: False
      type: str
    hash_name:
      description:
        - Name of the hash in which to store the result in redis.
      required: False
      type: str
    log_dir_path:
      description:
        - Path to log directory where logs will be stored.
      required: False
      type: str
"""

EXAMPLES = """
- name: Upgrade and downgrade goes
  goes_upgrade_downgrade:
    switch_name: "{{ inventory_hostname }}"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
    log_dir_path: "{{ log_dir_path }}"
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


def verify_goes_status(module, switch_name):
    """
    Method to verify if goes status is ok or not
    :param module: The Ansible module to fetch input parameters.
    :param switch_name: Name of the switch.
    :return: String describing if goes status is ok or not
    """
    global RESULT_STATUS
    failure_summary = ''

    # Get the GOES status info
    goes_status = execute_commands(module, 'goes status')

    if 'not ok' in goes_status.lower():
        RESULT_STATUS = False
        failure_summary += 'On switch {} '.format(switch_name)
        failure_summary += 'goes status is not ok\n'

    return failure_summary


def upgrade_downgrade_goes(module, installer, state):
    """
    Method to upgrade and downgrade goes version.
    :param module: The Ansible module to fetch input parameters.
    :param installer: Installer file.
    :param state: String describing if we want to upgrade or downgrade goes.
    :return: String describing if goes upgrade/downgrade was ok or not.
    """
    global RESULT_STATUS
    failure_summary = ''
    switch_name = module.params['switch_name']
    installer_dir = module.params['installer_dir']
    upgrade_version = module.params['upgrade_version'].splitlines()
    downgrade_version = module.params['downgrade_version'].splitlines()

    # Verify goes status before upgrade/downgrade
    failure_summary += verify_goes_status(module, switch_name)

    # upgrade/downgrade goes
    cmd = '{}./{}'.format(installer_dir, installer)
    out = execute_commands(module, cmd)

    if out is not None:
        if 'timeout' in out or 'exit status 1' in out:
            RESULT_STATUS = False
            failure_summary += 'On switch {} '.format(switch_name)
            failure_summary += 'goes {} failed\n'.format(state)

    # Verify version after upgrade/downgrade
    if state == 'upgrade':
        versions_to_verify = upgrade_version
    else:
        versions_to_verify = downgrade_version

    current_version = get_current_goes_version(module)
    for version in versions_to_verify:
        version = version.strip()
        if version not in current_version:
            RESULT_STATUS = False
            failure_summary += 'On switch {}'.format(switch_name)
            failure_summary += 'goes versions are not matching with given '
            failure_summary += '{} versions after {}\n'.format(state, state)

    # Verify goes status after upgrade/downgrade
    failure_summary += verify_goes_status(module, switch_name)

    return failure_summary


def get_current_goes_version(module):
    """
    Method to get current goes version.
    :param module: The Ansible module to fetch input parameters.
    :return: Current goes version.
    """
    return execute_commands(module, 'goes hget platina packages')


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            installer_dir=dict(required=False, type='str'),
            upgrade_installer_name=dict(required=False, type='str'),
            downgrade_installer_name=dict(required=False, type='str'),
            upgrade_version=dict(required=False, type='str'),
            downgrade_version=dict(required=False, type='str'),
            hash_name=dict(required=False, type='str'),
            log_dir_path=dict(required=False, type='str'),
        )
    )

    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    upgrade_installer = module.params['upgrade_installer_name']
    downgrade_installer = module.params['downgrade_installer_name']
    upgrade_version = module.params['upgrade_version'].splitlines()
    downgrade_version = module.params['downgrade_version'].splitlines()
    upgrade_version.pop()
    downgrade_version.pop()
    run_downgrade = False

    # Get the current goes version
    current_version = get_current_goes_version(module)

    for version in upgrade_version:
        version = version.strip()
        if version in current_version:
            run_downgrade = True

    # Upgrade/downgrade 5 times
    if run_downgrade:
        failure_summary += upgrade_downgrade_goes(module, downgrade_installer,
                                                  'downgrade')
        failure_summary += upgrade_downgrade_goes(module, upgrade_installer,
                                                  'upgrade')
        failure_summary += upgrade_downgrade_goes(module, downgrade_installer,
                                                  'downgrade')
        failure_summary += upgrade_downgrade_goes(module, upgrade_installer,
                                                  'upgrade')
        failure_summary += upgrade_downgrade_goes(module, downgrade_installer,
                                                  'downgrade')
    else:
        failure_summary += upgrade_downgrade_goes(module, upgrade_installer,
                                                  'upgrade')
        failure_summary += upgrade_downgrade_goes(module, downgrade_installer,
                                                  'downgrade')
        failure_summary += upgrade_downgrade_goes(module, upgrade_installer,
                                                  'upgrade')
        failure_summary += upgrade_downgrade_goes(module, downgrade_installer,
                                                  'downgrade')
        failure_summary += upgrade_downgrade_goes(module, upgrade_installer,
                                                  'upgrade')

    HASH_DICT['result.detail'] = failure_summary

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Create a log file
    log_file_path = module.params['log_dir_path']
    log_file_path += '/{}.log'.format(module.params['hash_name'])
    log_file = open(log_file_path, 'w')
    for key, value in HASH_DICT.iteritems():
        log_file.write(key)
        log_file.write('\n')
        log_file.write(str(value))
        log_file.write('\n')
        log_file.write('\n')

    log_file.close()

    # Exit the module and return the required JSON.
    module.exit_json(
        hash_dict=HASH_DICT,
        log_file_path=log_file_path
    )

if __name__ == '__main__':
    main()

