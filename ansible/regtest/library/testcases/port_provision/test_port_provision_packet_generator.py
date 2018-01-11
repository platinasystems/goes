#!/usr/bin/python
""" Test/Verify Port Provisioning on Packet Generator """

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
module: test_port_provision_packet_generator
author: Platina Systems
short_description: Module to execute and verify port provisioning on packet
generator.
description:
    Module to execute and verify port provisioning on packet generator.
options:
    switch_name:
      description:
        - Name of the switch on which tests will be performed.
      required: False
      type: str
    ce_list:
      description:
        - List of ce ports described as string.
      required: False
      type: str
      default: ''
    speed:
      description:
        - Speed of the eth interface port.
      required: False
      type: str
    create_cint_file:
      description:
        - Name of create cint file.
      required: False
      type: str
    delete_cint_file:
      description:
        - Name of delete cint file.
      required: False
      type: str
    autoeng:
      description:
        - Flag to indicate if autoeng is on or off.
      required: False
      type: bool
      default: False
    reset_config:
      description:
        - Flag to indicate if config needs to be reset.
      required: False
      type: bool
      default: False
    hash_name:
      description:
        - Name of the hash in which to store the result in redis.
      required: False
      type: str
"""

EXAMPLES = """
- name: Execute and verify port provisioning
  test_port_provision_packet_generator:
    switch_name: "{{ inventory_hostname }}"
    ce_list: "1,3,5,7,9,11,13,15"
    speed: "100g"
    hash_name: "{{ hostvars['server_emulator']['hash_name'] }}"
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


def verify_port_provisioning(module):
    """
    Method to execute and verify port provisioning.
    :param module: The Ansible module to fetch input parameters.
    """
    global RESULT_STATUS, HASH_DICT
    failure_summary = ''
    switch_name = module.params['switch_name']
    speed = module.params['speed']
    fec = module.params['fec']
    create_cint_file = module.params['create_cint_file']
    delete_cint_file = module.params['delete_cint_file']
    autoeng = module.params['autoeng']
    ce_list = module.params['ce_list'].split(',')
    initial_cli = "python /home/platina/bin/bcm.py"
    cint_path = '/home/platina/bin/'

    xe_list = ['3', '4', '5', '6', '10', '11', '12', '13', '17', '18', '19',
               '20', '24', '25', '26', '27', '31', '32', '33', '34', '38',
               '39', '40', '41', '45', '46', '47', '48', '52', '53', '54', '55']

    if not module.params['reset_config']:
        if speed == '100g':
            if not autoeng:
                # Port provision ce ports to given speed
                for ce in ce_list:
                    cmd = "{} 'port ce{} speed={} if=cr4'".format(initial_cli,
                                                                  ce, speed)
                    execute_commands(module, cmd)

                # Execute cint configuration script
                cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                              create_cint_file)
                execute_commands(module, cmd)
            else:
                # Port provision ce ports to given speed
                for ce in ce_list:
                    cmd = "{} 'port ce{} adv={} an=1 if=cr4'".format(
                        initial_cli, ce, speed)
                    execute_commands(module, cmd)

            # Verify if ce ports are up
            for ce in ce_list:
                cmd = "{} 'ps ce{}'".format(initial_cli, ce)
                out = execute_commands(module, cmd)

                if 'up' not in out.lower():
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'ce{} port is not up\n'.format(ce)

        elif speed == '25g':
            # Install copper cable and configure lanes on ce ports
            for ce in range(1, 9):
                cmd = "{} 'port ce{} en=f'".format(initial_cli, ce)
                execute_commands(module, cmd)
                cmd = "{} 'port ce{} lanes 1'".format(initial_cli, ce)
                execute_commands(module, cmd)

            # Set the speed to given speed
            for xe in xe_list:
                cmd = "{} 'port xe{} an=t'".format(initial_cli, xe)
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} an=f'".format(initial_cli, xe)
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} an=f'".format(initial_cli, xe)
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} speed=25000 an=0 if=cr'".format(
                    initial_cli, xe, speed)
                execute_commands(module, cmd)

            # Execute cint configuration script
            cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                          create_cint_file)
            execute_commands(module, cmd)

            for xe in xe_list:
                cmd = "{} 'port xe{} ena=t'".format(initial_cli, xe)
                execute_commands(module, cmd)

            # Verify if xe ports are up
            for xe in xe_list:
                cmd = "{} 'ps xe{}'".format(initial_cli, xe)
                out = execute_commands(module, cmd)

                if 'up' not in out.lower():
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'xe{} port is not up\n'.format(xe)

        elif speed == '40g':
            xe_list = ['3', '10', '17', '24', '31', '38', '45', '52']
            if not autoeng:
                # Port provision ce ports to given speed
                for ce in ce_list:
                    cmd = "{} 'port ce{} en=0'".format(initial_cli, ce)
                    execute_commands(module, cmd)

                for ce in range(1, 9):
                    cmd = "{} 'port ce{} speed=40000 if=cr'".format(
                        initial_cli, ce)
                    execute_commands(module, cmd)

                if fec:
                    for xe in xe_list:
                        cmd = "{} 'phy xe{} AN_X4_LD_BASE_ABIL1r FEC_REQ=3'".format(
                            initial_cli, xe)
                        execute_commands(module, cmd)

                # Execute cint configuration script
                cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                              create_cint_file)
                execute_commands(module, cmd)

                for xe in xe_list:
                    cmd = "{} 'port xe{} ena=1'".format(initial_cli, xe)
                    execute_commands(module, cmd)
            else:
                # Port provision ce ports to given speed
                for ce in range(1, 9):
                    cmd = "{} 'port ce{} en=0'".format(initial_cli, ce)
                    execute_commands(module, cmd)

                    cmd = "{} 'port ce{} adv={} an=1 if=cr4'".format(
                        initial_cli, ce, speed)
                    execute_commands(module, cmd)

                    req = '3' if fec == 'cl74' else '1'

                    cmd = "{} 'phy ce{} AN_X4_LD_BASE_ABIL1r FEC_REQ={}'".format(
                        initial_cli, ce, req)
                    execute_commands(module, cmd)

                    cmd = "{} 'port ce{} ena=1'".format(initial_cli, ce)
                    execute_commands(module, cmd)

            # Verify if xe ports are up
            for xe in xe_list:
                cmd = "{} 'ps xe{}'".format(initial_cli, xe)
                out = execute_commands(module, cmd)

                if 'up' not in out.lower():
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'xe{} port is not up\n'.format(xe)
        elif speed == '10g':
            # Install copper cable and configure lanes on ce ports
            for ce in range(1, 9):
                cmd = "{} 'port ce{} en=f'".format(initial_cli, ce)
                execute_commands(module, cmd)
                cmd = "{} 'port ce{} lanes 1'".format(initial_cli, ce)
                execute_commands(module, cmd)

            if autoeng:
                # Port provision xe ports to given speed
                for xe in xe_list:
                    cmd = "{} 'port xe{} adv={} an=1 if=kr'".format(
                        initial_cli, xe, speed)
                    execute_commands(module, cmd)

                    cmd = "{} 'phy xe{} AN_X4_LD_BASE_ABIL1r FEC_REQ=3'".format(
                        initial_cli, xe)
                    execute_commands(module, cmd)
            else:
                for xe in xe_list:
                    cmd = "{} 'port xe{} speed=10000 an=0 if=kr'".format(
                        initial_cli, xe)
                    execute_commands(module, cmd)

                # Execute cint configuration script
                cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                              create_cint_file)
                execute_commands(module, cmd)

                for xe in xe_list:
                    cmd = "{} 'port xe{} ena=t'".format(initial_cli, xe)
                    execute_commands(module, cmd)

            # Verify if xe ports are up
            for xe in xe_list:
                cmd = "{} 'ps xe{}'".format(initial_cli, xe)
                out = execute_commands(module, cmd)

                if 'up' not in out.lower():
                    RESULT_STATUS = False
                    failure_summary += 'On switch {} '.format(switch_name)
                    failure_summary += 'xe{} port is not up\n'.format(xe)

        # Generate traffic
        cmd = 'python /home/platina/bin/autoTest_pktDrop_by_pktSizes.py '
        cmd += '30 31 1 10'
        execute_commands(module, cmd)
    else:
        # Reset all config
        if speed == '40g':
            for xe in ['3', '9', '15', '21', '27', '33', '39', '45']:
                cmd = "{} 'port xe{} en=f'".format(initial_cli, xe)
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} lanes 4'".format(initial_cli, xe)
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} speed=100000 an=0 if=cr4'".format(
                    initial_cli, xe)
                execute_commands(module, cmd)
        elif speed == '25g' or speed == '10g':
            port_dict = {}
            for ce in ce_list:
                index = ce_list.index(ce) * 4
                port_dict[ce] = [xe_list[i] for i in range(index, index + 4)]

            for ce, xe_ports in port_dict.items():
                for xe in xe_ports:
                    cmd = "{} 'port xe{} en=f'".format(initial_cli, xe)
                    execute_commands(module, cmd)

                cmd = "{} 'port xe{} lanes 4'".format(initial_cli, xe_ports[0])
                execute_commands(module, cmd)

                cmd = "{} 'port xe{} speed=100000 an=0 if=cr4'".format(
                    initial_cli, xe_ports[0])
                execute_commands(module, cmd)
        else:
            for ce in ce_list:
                cmd = "{} 'port ce{} en=f'".format(initial_cli, ce)
                execute_commands(module, cmd)

                cmd = "{} 'port ce{} lanes 4'".format(initial_cli, ce)
                execute_commands(module, cmd)

                cmd = "{} 'port ce{} speed=100000 an=0 if=cr4'".format(
                    initial_cli, ce)
                execute_commands(module, cmd)

        if speed == '100g' or speed == '40g':
            if not autoeng:
                # Delete cint configuration
                cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                              delete_cint_file)
                execute_commands(module, cmd)
        elif speed == '25g' or speed == '10g':
            # Delete cint configuration
            cmd = "{} 'cint {}{}'".format(initial_cli, cint_path,
                                          delete_cint_file)
            execute_commands(module, cmd)

        for ce in ce_list:
            cmd = "{} 'port ce{} ena=t'".format(initial_cli, ce)
            execute_commands(module, cmd)

    HASH_DICT['result.detail'] = failure_summary


def main():
    """ This section is for arguments parsing """
    module = AnsibleModule(
        argument_spec=dict(
            switch_name=dict(required=False, type='str'),
            ce_list=dict(required=False, type='str', default=''),
            speed=dict(required=False, type='str'),
            fec=dict(required=False, type='str', default=''),
            create_cint_file=dict(required=False, type='str'),
            delete_cint_file=dict(required=False, type='str'),
            autoeng=dict(required=False, type='bool', default=False),
            reset_config=dict(required=False, type='bool', default=False),
            hash_name=dict(required=False, type='str')
        )
    )

    global HASH_DICT, RESULT_STATUS

    # Verify port_provisioning
    verify_port_provisioning(module)

    # Calculate the entire test result
    HASH_DICT['result.status'] = 'Passed' if RESULT_STATUS else 'Failed'

    # Exit the module and return the required JSON.
    module.exit_json(
        hash_dict=HASH_DICT
    )

if __name__ == '__main__':
    main()

