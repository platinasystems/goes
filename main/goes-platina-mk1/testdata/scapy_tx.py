#!/usr/bin/python
#
# depends on scapy, find it here:
#   http://scapy.readthedocs.io/en/latest/installation.html
#
import sys
import argparse
from scapy.all import *

def rand_hex_byte():
  return '{:02x}'.format(random.randrange(0,255))

def rand_mac():
  return "ee:ee:ee:"+rand_hex_byte()+":"+rand_hex_byte()+":"+rand_hex_byte()

def rand_ip():
  return "192.168."+str(random.randrange(1,255))+"."+str(random.randrange(1,255))

def rand_ip_port():
  return random.randrange(50000,60000)

def rand_vid():
  return random.randrange(1,4096)

def rand_pri():
  return random.randrange(0,8)

def mac_incr(mac, offset):
  val = int(mac.replace(':',''), 16)
  macstr = '{:012x}'.format(val+offset)
  return ':'.join(macstr[i:i+2] for i in range(0, len(macstr), 2))

def getmac(interface):

  try:
    mac = open('/sys/class/net/'+interface+'/address').readline()
  except:
    mac = ""

  return mac[0:17]



parser = argparse.ArgumentParser(description='send packets')
parser.add_argument('-mac', '--mac', nargs="*", help='mac [intf | da | bcast | bpdu] [intf | sa]')
parser.add_argument('-smi', '--smi', action='store_true', help='SA incr')
parser.add_argument('-v', '--vidpri', nargs="*", help='vlan [prio]')
parser.add_argument('-ip', '--ip', nargs="*", help='ip [dstip] [srcip] [udp | tcp | arp]')
parser.add_argument('-t', '--txintf', help='tx interface')
parser.add_argument('-l', '--length', help='length')
parser.add_argument('-c', '--count', help='count')

args = parser.parse_args()

# defaults
DA = "00:00:ee:00:00:0d"
DA_bcast = "ff:ff:ff:ff:ff:ff"
DA_bpdu = "01:80:c2:00:00:00"
SA = "00:00:ee:00:00:05"
prio = 0
vid = 0
dstip = ""
srcip = "192.168.1.1"
proto = "udp"
payload = "012345678901234567890123456789012345678901234567890123456789"
payload60 = "012345678901234567890123456789012345678901234567890123456789"
length = 128
count = 1
txintf = ""
smi = 0

if args.mac:
  if args.mac[0] == "bcast":
    DA = DA_bcast
  elif args.mac[0] == "bpdu":
    DA = DA_bpdu
  else:
    DA = getmac(args.mac[0])
    if len(DA) == 0:
      DA = args.mac[0]
  if len(args.mac) > 1:
    SA = getmac(args.mac[1])
    if len(SA) == 0:
      SA = args.mac[1]

if args.vidpri:
  vid = int(args.vidpri[0])
  if len(args.vidpri) > 1:
    prio = int(args.vidpri[1])

if args.ip:
  dstip = args.ip[0]
  if len(args.ip) > 1:
    srcip = args.ip[1]
  if len(args.ip) > 2:
    proto = args.ip[2]

if args.length:
  length = args.length

if args.count:
  count = int(args.count)

if args.txintf:
  txintf = args.txintf

if args.smi:
  smi = 1

for dst_port in range(0,count):
  if dstip:
    if vid != 0:
      if proto == "udp":
        p=(Ether(src=SA, dst=DA)/Dot1Q(vlan=vid, prio=prio)/IP(src=srcip,dst=dstip)/UDP(dport=50000)/payload)
      elif proto == "tcp":
        p=(Ether(src=SA, dst=DA)/Dot1Q(vlan=vid, prio=prio)/IP(src=srcip,dst=dstip)/TCP(dport=50000)/payload)
      elif proto == "arp":
        p=(Ether(src=SA, dst=DA_bcast)/Dot1Q(vlan=vid, prio=prio)/ARP(psrc=srcip,pdst=dstip,hwsrc=SA,hwtype=1,ptype=2048,op=1)/payload)
    else:
      if proto == "udp":
        p=(Ether(src=SA, dst=DA)/IP(src=srcip,dst=dstip)/UDP(dport=50000)/payload)
      elif proto == "tcp":
        p=(Ether(src=SA, dst=DA)/IP(src=srcip,dst=dstip)/TCP(dport=50000)/payload)
      elif proto == "arp":
        p=(Ether(src=SA, dst=DA_bcast)/ARP(psrc=srcip,pdst=dstip,hwsrc=SA,hwtype=1,ptype=2048,op=1)/payload)
  else:
    if vid != 0:
      p=(Ether(src=SA, dst=DA)/Dot1Q(vlan=vid, prio=prio, type=60)/payload60)
    else:
      p=(Ether(src=SA, dst=DA, type=60)/payload60)

  if len(txintf) == 0:
    print p.show(dump=True)
  else:
    sendp(p, iface=txintf, verbose=0)

  if smi:
    SA = mac_incr(SA, 1)

if len(txintf) == 0:
  print "No packets sent"
else:
  print str(count) + " packets sent to " + txintf
