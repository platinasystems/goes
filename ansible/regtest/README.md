# ANSIBLE REGRESSION SUITE

To run any playbook, please follow below guidelines:

Make sure you are in `go/ansible/regtest` directory.

Once you are in this path, all playbooks can be found inside `playbooks/` directory. 

Before executing the playbook, make sure the required package (quagga/bird/frr/gobgp) is installed in the testbed, on which you are trying to execute the playbook. All package installation and uninstallation playbooks can be found in `playbooks/installation` directory.

Suppose, you want to execute `bird_bgp_peering_ebgp_route_advertise.yml` playbook on testbed2, then first run the bird installation playbook:

```
    ansible-playbook -i hosts_testbed2 playbooks/installation/bird_install.yml -K
```

And then run the bird_bgp_peering_ebgp_route_advertise.yml playbook by executing this command:

```
    ansible-playbook -i hosts_testbed2 playbooks/bgp/bird_bgp_peering_ebgp_route_advertise.yml -K
```
