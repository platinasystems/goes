#!/usr/bin/python
import sys
import argparse
from subprocess import call
from subprocess import check_output
import re

vid_base  = 3000
port_base = 1
numports = 32

def verify_links():
  link_down = 0
  for portname in portlist:
    status = check_output("ip link show " + portname, shell=True)
    link = re.findall(r'state [^ ]* ', status, re.DOTALL)
    if "UP" not in str(link):
      print portname + ":" + str(link)
      link_down += 1

  if link_down:
    print "All " + str(numports) + " links must be up to run test, " + str(link_down) + " links down."
    exit()
  else:
    print "All " + str(numports) + " links up"

def verify_wiring():
  call("goes vnet clear int", shell=True)
  for idx in range(0,numports,2):
    call("./scapy_tx.py -i " + portlist[idx], shell=True)
    status = check_output("goes vnet show ha " + portlist[idx+1], shell=True)
    rx = re.findall(r'.*rx packets.*', status)
    print rx

def config_pkt_loop():
  for portname in portlist:
    call("goes vnet set fe1 port-t l2 true " + portname, shell=True)

  # each port pair is looped externally
  # each offset pair is in a vlan
  vid = vid_base
  for idx in range(0,numports,2):
    prev_idx = idx - 1
    if prev_idx < 0:
      prev_idx = numports - 1

    vidports = portlist[prev_idx] + " " + portlist[idx]
    call("goes vnet set fe1 vlan-create " + str(vid), shell=True)
    call("goes vnet set fe1 vlan-port-add " + str(vid) + " u " + vidports, shell=True)
    call("goes vnet set fe1 port-t ovid " + str(vid) + " " + vidports, shell=True)
    vid += 1

def tx():
  call("./scapy_tx.py -i " + portlist[0], shell=True)
  call("./scapy_tx.py -i " + portlist[1], shell=True)

parser = argparse.ArgumentParser()
parser.add_argument('-first_port', '--first_port')
parser.add_argument('-num_ports', '--num_ports')
parser.add_argument('-link', '--link', action='store_true', help='verify link status')
parser.add_argument('-wiring', '--wiring', action='store_true', help='verify wiring')
parser.add_argument('-config', '--config', action='store_true', help='configure loop test')
parser.add_argument('-tx', '--tx', action='store_true', help='start traffic')
parser.add_argument('-stats', '--stats', action='store_true', help='show stats')

args = parser.parse_args()

if args.first_port:
  port_base = int(args.first_port)

if args.num_ports:
  numports = int(args.num_ports)
  if numports & 1:
    print "Must select even number of ports"
    exit()

portlist= []
for i in range(numports):
  portlist.append("xeth" + str(port_base+i))

if args.link:
  verify_links()

if args.wiring:
  verify_wiring()

if args.config:
  config_pkt_loop()

if args.tx:
  tx()

if args.stats:
  status = check_output("goes vnet show ha|grep -e ^x -e x.pack;goes vnet clear int", shell=True)
  print status
