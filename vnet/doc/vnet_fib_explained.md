# vnet fib, neighbor, next hop, and adjacencies explained

For now this document applies to ipv4; many of these structures and concepts will be reused in ipv6.

There are multiple fib tables in goes.  Each one represents a namespace or vrf.  They are identified by its index.  The structure for ipv4 fib is:

```
type  Fib  struct {
index ip.FibIndex

reachable, unreachable MapFib

// Mtrie for fast lookups.
mtrie
}
```

```
// Maps for prefixes for /0 through /32; key in network byte order.
type MapFib [1 + 32]map[vnet.Uint32]mapFibResult
```
The key vnet.Uint32 is made up of the ip address with the mask applied.
```
func (p *Prefix) mapFibKey() vnet.Uint32     { return p.Address.AsUint32() & p.Mask() }
```
```
type mapFibResult struct {
        adj ip.Adj
        nh  mapFibResultNextHop
}
```
```
type mapFibResultNextHop map[idst]map[ipre]NextHopper
```

Fundamentally fib is a map of ip prefixes and the adjacencies for them.  The next hop for the prefix can be retrieved indirectly via the adjacency index adj in the mapFibResult  struct.   The nh variable in mapFibResult is not the next hop for the ip prefix.  Rather nh maps the other prefixes that uses this mapFibResult or its adjacency as their nexthop(s).

reachable is a map of all prefixes that have reachable next hops
unreachable is a map of all prefixes have have next hops that are not reachable

There are 3 ways fib entries are created.

### Interface Adjacency
This adjacency is created when an address is assigned to an interface.  vnet creates 2 fib entries for each interface: a "local" /32 address identify that traffic destined for it should be punted to Linux instead of routed at fe1; and a "glean" adjacency, which says that routes within that subnet should go out on this adjacency.  This is done in vnet/ip4/fib.go 
```
func (m *Main) addDelInterfaceAddressRoutes(ia ip.IfAddr, isDel bool)
```
The adjacency type for them are of type LookupNextLocal and LookupNextGlean, respectively.

### Neighbor Adjacency and Punt Adjacency
#### Neighbor
When a neighbor address becomes reachable via arp, the neighbor address and associated interface is added to vnet's ip neighbor structures managed in vnet/ethernet/neighbor.go.  The associated rewrite and the adjacency that identifies it is formed in neighbor.go.  From neighbor.go, a call will come into fib.go, via AddDelRoute, to add or delete a route prefix with that adjacency using
```
func (m *Main) addDelRoute(p *ip.Prefix, fi ip.FibIndex, baseAdj ip.Adj, isDel bool) (oldAdj ip.Adj, err error)
```
This creates 1 fib entry, and the adjacency is of type LookupNextRewrite.

#### Punt
When an virtual interface not associated with a front panel interface, e.g. type dummy, is created in Linux, a fib entry is created with adjacency of type LookupNextPunt.  This entry tells fe1 that packet with that prefix should be redirected toward Linux with no rewrite and no additional vlan encap.  Punt fib add/del comes through the same code path in fib.go as neighbor adjacency, via 
```
func (m *Main) addDelRoute(p *ip.Prefix, fi ip.FibIndex, baseAdj ip.Adj, isDel bool) (oldAdj ip.Adj, err error)
```
except that the adjacency installed is a of type LookupNextPunt instead of LookupNextRewrite.

### Next Hop
This is by far the most complicated type of fib entry add/del

With the previous 3 types, interface, neighbor, punt, there is 1 adjacency for 1 prefix.  The adjacency type is fixed, and has at most 1 rewrite.  Adding and deleting is straightforward; there are no interdependencies.

When a fib entry is created because there is a new next hop for it, the request comes into fib.go via
```
func (m *Main) AddDelRouteNextHop(p *Prefix, nh *NextHop, isDel bool, isReplace bool) (err error)
```
At a high level, three things happen after this.  First an new adjacency is created.  While it is the same adjacency structure, indexed under the same adj index scheme, as any other adjacencies, the same adj index indexes an additional multipathAdjacency heap that allows for multiple next hop rewrites.  Even if there is just a single next hop, a multipathAdjacency is created with a single entry in it.  Second of all,  the interdependencies of which adjacency references which other adjacencies its next hops are woven into the mapFibResult.nh map.  This is so that later if 1 adjacency is removed or if a fib entry is deleted, vnet can traverse and the mapFibResult.nh to figure out what other adjacencies need to be updated accordingly or if an adjacency can be deleted.  The code references an adjacency index that also indexes a multipathAdjacency often as mpAdj.  Lastly, depending on whether the adjacency is reachable or not, this fib entry is placed or moved into either the reachable MapFib or unreachable MapFib

### Example 1
An address is assigned to xeth1
``` 
platina@invader34:~$ ip addr add 10.0.0.1/24 brd + dev xeth1 
```
Now the glean and local adjacencies are created with adj index 3 and 4 respectively
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
 ```
In the actual hardware tcam table, 2 entries are created as a result.  Both are of the rxf_class_id "punt" meaning that they will be redirected toward Linux.
```
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH     
tcam      0    32          31       0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0           
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0  
``` 
In the hardware adjacency table, no new adjacency are created.  The existing entries are there by virtue of interfaces being admin up (fe1-cpu is the PCIE interface)
```
platina@invader34:~$ goes vnet show fe1 adj
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1 
```
### Example 2
Now let's add another interface, but this time in a namespace called R2, and give it an address
```
platina@invader34:~$ sudo ip netns add R2
platina@invader34:~$ sudo ip link set xeth2 netns R2
platina@invader34:~$ sudo ip netns exec R2 ip link set xeth2 up
platina@invader34:~$ sudo ip netns exec R2 ip addr add 10.0.0.2/24 brd + dev xeth2
```
glean and local are created, but now in a different fib table called R2
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.2/32       6: local xeth2
```
Tcam entries added.  vrf 1 is namesapce R2, vrf 0 is default namespace.  Note the mapping of namespace name to vrf index is not always the same (first namespace to appear grabs the next available vrf index).
```
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH       
tcam      0    32          30       1           0    1    10.0.0.2            255.255.255.255     true           punt      adj 0      
tcam      0    32          31       0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0           
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0      
tcam      0    24          63       1           0    1    10.0.0.0            255.255.255.0       true           punt      adj 0 
```
Adjacency table got 1 more entry because xeth2 is admin up
```
platina@invader34:~$ goes vnet show fe1 adj pipe 0
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1      
hardware  0       3     used      nil   unicast   meth-2       false    false    ...  2       l3_unicast          false      false      0       00:00:00:00:00:00 4          xeth2  
```
At this point, no neighbor has been establish because no arp was initiated.
```
platina@invader34:~$ sudo ip netns exec R2 ip neigh
platina@invader34:~$ 
```
Let's start a ping to arp and establish neighbor
```
platina@invader34:~$ sudo ip netns exec R2 ping 10.0.0.1 -c 2
PING 10.0.0.1 (10.0.0.1) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_seq=1 ttl=64 time=0.135 ms
64 bytes from 10.0.0.1: icmp_seq=2 ttl=64 time=0.118 ms
```
Now we have neighbors in both namespaces
```
platina@invader34:~$ sudo ip netns exec R2 ip neigh
10.0.0.1 dev xeth2 lladdr 50:18:4c:00:0a:44 REACHABLE
platina@invader34:~$ sudo ip neigh | grep 10.0.0.2
10.0.0.2 dev xeth1 lladdr 50:18:4c:00:0a:45 REACHABLE
```
Fib now have new entries established from the neighbor messages.  Adjacencies 7 and 8 are rewrites.
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
     default                   10.0.0.2/32       7: rewrite xeth1 IP4: 50:18:4c:00:0a:44 -> 50:18:4c:00:0a:45
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.1/32       8: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44
          R2                   10.0.0.2/32       6: local xeth2
```
Tcam table got 2 new entries.  Note their rxf_class_id are not "punt", and they have ecmp/adj index associated with them.
> The ecmp/adj index in the tcam table is an index to the hardware adjacency table, not the vnet fib adjancency.
```
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH    
tcam      0    32          29       1           0    1    10.0.0.1            255.255.255.255     true            nil      adj 5      
tcam      0    32          30       0           0    0    10.0.0.2            255.255.255.255     true            nil      adj 4      
tcam      0    32          30       1           0    1    10.0.0.2            255.255.255.255     true           punt      adj 0      
tcam      0    32          31       0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0         
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0      
tcam      0    24          63       1           0    1    10.0.0.0            255.255.255.0       true           punt      adj 0  
```
Adjacency table got 2 new entires.  Note they have dstRewrite and srcRewrite fields marked as true, and the ipAdj fields for them corresponds to the adjacency indices 7 and 8 from the fib table.
> The ecmp/adj index from the tcam table references the index in column l3_NH.  The associated vnet fib adjacency index is in the column ipAdj.
```
platina@invader34:~$ goes vnet show fe1 adj pipe 0
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1      
hardware  0       3     used      nil   unicast   meth-2       false    false    ...  2       l3_unicast          false      false      0       00:00:00:00:00:00 4          xeth2      
hardware  0       4     used      7     tunnel    xeth1        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:45 1          xeth1      
hardware  0       5     used      8     tunnel    xeth2        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:44 2          xeth2
``` 
### Example 3
Now let's add a route with 1 next hop.
```
platina@invader34:~$ sudo ip netns exec R2 ip route add 10.5.5.5 via 10.0.0.1 
platina@invader34:~$ sudo ip netns exec R2 ip route
10.0.0.0/24 dev xeth2  proto kernel  scope link  src 10.0.0.2 
```
We see this entries added to the tcam table.  But even though the rewrite is the same, it got assigned a different adjacency index 9.  Furthermore, this adjacency index 9 references adjacency index 8 as its next hop adjacency with a weight of 1.  Adjacency with index 9 is what the vnet code would consider a mpAdj.
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
     default                   10.0.0.2/32       7: rewrite xeth1 IP4: 50:18:4c:00:0a:44 -> 50:18:4c:00:0a:45
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.1/32       8: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44
          R2                   10.0.0.2/32       6: local xeth2
          R2                   10.5.5.5/32       9: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44 adj-range 9-9, weight 1 nh-adj 8
```
Tcam got a new entry for 10.5.5.5
```
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH      
tcam      0    32          29       0           0    1    10.5.5.5            255.255.255.255     true            nil      adj 9      
tcam      0    32          29       1           0    1    10.0.0.1            255.255.255.255     true            nil      adj 5      
tcam      0    32          30       0           0    0    10.0.0.2            255.255.255.255     true            nil      adj 4      
tcam      0    32          30       1           0    1    10.0.0.2            255.255.255.255     true           punt      adj 0      
tcam      0    32          31       0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0            
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0      
tcam      0    24          63       1           0    1    10.0.0.0            255.255.255.0       true           punt      adj 0 
```
Adjacency table got a new entry.  That L3_NH and ipAdj both have value 9 for this new entry is just coincidence. 
```
platina@invader34:~$ goes vnet show fe1 adj pipe 0
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1      
hardware  0       3     used      nil   unicast   meth-2       false    false    ...  2       l3_unicast          false      false      0       00:00:00:00:00:00 4          xeth2      
hardware  0       4     used      7     tunnel    xeth1        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:45 1          xeth1      
hardware  0       5     used      8     tunnel    xeth2        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:44 2          xeth2      
hardware  0       9     used      9     tunnel    xeth2        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:44 2          xeth2 
```

### Example 4
Let's restart goes and start over.  This time we create the name space and interface address as before, but we do not ping to establish arp yet.   We go ahead and add a route 10.6.6.6 via 10.0.0.1, and we see that the route is show in the fib table as "unreachable via 10.0.0.1".  This means the entries is added to the fib structure, but in the **unreachable** MapFib instead of the **reachable** MapFib.  Note that no adjacency index is assigned to it.
```
platina@invader34:~$ sudo goes restart
platina@invader34:~$ sudo ip netns add R2
platina@invader34:~$ sudo ip link set xeth2 netns R2
platina@invader34:~$ sudo ip netns exec R2 ip addr add 10.0.0.2/24 brd + dev xeth2
platina@invader34:~$ sudo ip netns exec R2 ip link set xeth2 up
platina@invader34:~$ sudo ip netns exec R2 ip neigh
platina@invader34:~$ sudo ip netns exec R2 ip route add 10.6.6.6 via 10.0.0.1                                                                                                
platina@invader34:~$ goes vnet show ip fib                                                                                                                                   
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.2/32       6: local xeth2
          R2                   10.6.6.6/32          unreachable via 10.0.0.1
```
For completion let dump the tcam and adjacency tables.  No adjacency or rewrite for 10.0.0.1 or 10.6.6.6 as we expected.
```
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH  
tcam      0    32          0        0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0      
tcam      0    32          30       1           0    1    10.0.0.2            255.255.255.255     true           punt      adj 0           
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0      
tcam      0    24          63       1           0    1    10.0.0.0            255.255.255.0       true           punt      adj 0      

platina@invader34:~$ goes vnet show fe1 adj pipe 0
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1      
hardware  0       3     used      nil   unicast   meth-2       false    false    ...  2       l3_unicast          false      false      0       00:00:00:00:00:00 1          xeth2 
```
Now we ping 0.0.0.1 to establish neighbor.
```
platina@invader34:~$ sudo ip netns exec R2 ping 10.0.0.1 -c 1
PING 10.0.0.1 (10.0.0.1) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_seq=1 ttl=64 time=0.086 ms
```
10.0.0.1 is now added to fib table with valid rewrite.  Because the interdependencies of between 10.0.0.1 and 10.6.6.6 is built into the fib structure, vnet will update and move 10.6.6.6 from unreachable to reachable MapFib and populate entries into the fe1 tcam and adjacency tables.
```
platina@invader34:~$ goes vnet show ip fib                                                                                                                                   
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
     default                   10.0.0.2/32       9: rewrite xeth1 IP4: 50:18:4c:00:0a:44 -> 50:18:4c:00:0a:45
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.1/32       7: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44
          R2                   10.0.0.2/32       6: local xeth2
          R2                   10.6.6.6/32       8: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44 adj-range 8-8, weight 1 nh-adj 7
platina@invader34:~$ goes vnet show fe1 tcam pipe 0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH  
tcam      0    32          0        0           0    0    10.0.0.1            255.255.255.255     true           punt      adj 0      
tcam      0    32          29       0           0    0    10.0.0.2            255.255.255.255     true            nil      adj 9      
tcam      0    32          29       1           0    1    10.6.6.6            255.255.255.255     true            nil      adj 5      
tcam      0    32          30       0           0    1    10.0.0.1            255.255.255.255     true            nil      adj 4      
tcam      0    32          30       1           0    1    10.0.0.2            255.255.255.255     true           punt      adj 0          
tcam      0    24          32       0           0    0    10.0.0.0            255.255.255.0       true           punt      adj 0      
tcam      0    24          63       1           0    1    10.0.0.0            255.255.255.0       true           punt      adj 0      

platina@invader34:~$ goes vnet show fe1 adj pipe 0
ING_L3_NEXT_HOP and EGR_L3_NEXT_HOP
  tcam transit routes to L3_NH of xeth
  acl-rxf rule punts to L3_NH of meth
sw/asic   rx_pipe L3_NH free/used ipAdj type      port         drop     copyCpu  ...  tx_pipe type                dstRewrite srcRewrite classID dstAddr           L3_INTF    si_name    
hardware  0       1     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 1          fe1-cpu    
hardware  0       2     used      nil   unicast   meth-1       false    false    ...  1       l3_unicast          false      false      0       00:00:00:00:00:00 2          xeth1      
hardware  0       3     used      nil   unicast   meth-2       false    false    ...  2       l3_unicast          false      false      0       00:00:00:00:00:00 1          xeth2      
hardware  0       4     used      7     tunnel    xeth2        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:44 1          xeth2      
hardware  0       5     used      8     tunnel    xeth2        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:44 1          xeth2      
hardware  0       9     used      9     tunnel    xeth1        false    false    ...  0       l3_unicast          true       true       0       50:18:4c:00:0a:45 2          xeth1 
```

### Example 5
Restart goes and add a dummy interface
```
platina@invader34:~$ sudo goes restart
platina@invader34:~$ sudo ip link add lo0 type dummy
```
A new entry is added to fib table
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.1.1.1/32       2: punt
```
and the tcam 
```
platina@invader34:~$ goes vnet show fe1 tcam pipe0
sw/tcam   pipe prefix_len  L3_DEFIP half_index  type vrf  key                 mask                valid  rxf_class_id ecmp/adj L3_NH  
tcam      0    32          0        0           0    0    10.1.1.1            255.255.255.255     true           punt      adj 0  
```

## Same Examples with code path trace

### Example 1
```
platina@invader34:~$ sudo goes restart
platina@invader34:~$ sudo ip link set xeth1 up
platina@invader34:~$ sudo ip addr add 10.0.0.1/24 brd + dev xeth1
```
After going through func (m *Main) addDelInterfaceAddressRoutes(ia ip.IfAddr, isDel bool), calls are made to:
```
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.1/24 adj 3 add: call fe1 hooks, addDelReachable
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.1/24 adj 3 IsMpAdj false
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.1/24 adj 3 IsMpAdj false finished: r {3 map[]
}
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.1/32 adj 4 add: call fe1 hooks, addDelReachable
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.1/32 adj 4 IsMpAdj false
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.1/32 adj 4 replaceWithMoreSpecific lr {3 map[
]} r {4 map[]}
Oct 17 17:47:04 invader34 goes.vnetd[11645]: 
Oct 17 17:47:04 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.1/32 adj 4 IsMpAdj false finished: r {4 map[]
}
```
addDel calls addDelReachable to add the fib entry to the reachable MapFib map, and also call the fibAddDelHooks that actually add/del the routes and adjacencies to the tcam

### Example 2
```
platina@invader34:~$ sudo ip link set xeth1 up
platina@invader34:~$ sudo ip addr add 10.0.0.1/24 brd + dev xeth1
platina@invader34:~$ sudo ip netns add R2
platina@invader34:~$ sudo ip link set xeth2 netns R2
platina@invader34:~$ sudo ip netns exec R2 ip link set xeth2 up
platina@invader34:~$ sudo ip netns exec R2 ip addr add 10.0.0.2/24 brd + dev xeth2
platina@invader34:~$ sudo ip netns exec R2 ping 10.0.0.1 -c 1
PING 10.0.0.1 (10.0.0.1) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_seq=1 ttl=64 time=0.407 ms
```
```
Oct 17 19:49:06 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/unix.(*net_namespace_main).add_del_namespace() add_namespace R2
Oct 17 19:49:06 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/unix.(*net_namespace_main).add_del_namespace() add_namespace R2
Oct 17 19:49:06 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/unix.(*net_namespace_main).add_del_namespace() add_namespace R2
Oct 17 19:49:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/unix.(*net_namespace).addDelMk1Interface() ns default isDel true ifname xeth2 ifindex 4930 address [80 24 76 0 10 69] devtype 0 iflinkindex 4 vlanid 4093
Oct 17 19:49:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/unix.(*net_namespace).addDelMk1Interface() ns R2 isDel false ifname xeth2 ifindex 4930 address [80 24 76 0 10 69] devtype 0 iflinkindex 4 vlanid 4093
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() R2 prefix 10.0.0.2/24 adj 5 add: call fe1 hooks, addDelReachable
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.2/24 adj 5 IsMpAdj false
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.2/24 adj 5 IsMpAdj false finished: r {5 map[]}
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() R2 prefix 10.0.0.2/32 adj 6 add: call fe1 hooks, addDelReachable
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.2/32 adj 6 IsMpAdj false
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.2/32 adj 6 replaceWithMoreSpecific lr {5 map[]} r {6 map[]}
Oct 17 19:50:42 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.2/32 adj 6 IsMpAdj false finished: r {6 map[]}
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0] adj 7 to xeth1
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.2 adj 7
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.2/32 adj 7 add: call fe1 hooks, addDelReachable
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 replaceWithMoreSpecific lr {3 map[]} r {7 map[]}
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false finished: r {7 map[]}
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0] adj 8 to xeth2
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.1 adj 8
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() R2 prefix 10.0.0.1/32 adj 8 add: call fe1 hooks, addDelReachable
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 IsMpAdj false
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 replaceWithMoreSpecific lr {5 map[]} r {8 map[]}
Oct 17 19:51:14 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 IsMpAdj false finished: r {8 map[]}
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0] adj 7 to xeth1
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.2 adj 7
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.2/32 adj 7 add: call fe1 hooks, addDelReachable
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 replaceWithMoreSpecific lr {3 map[]} r {7 map[]}
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false finished: r {7 map[]}
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0] adj 7 to xeth1
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.2 adj 7
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.2/32 adj 7 add: call fe1 hooks, addDelReachable
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 replaceWithMoreSpecific lr {3 map[]} r {7 map[]}
Oct 17 19:51:19 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false finished: r {7 map[]}
Oct 17 19:51:34 invader34 ntpd[2159]: Listen normally on 281 xeth1 fe80::5218:4cff:fe00:a44 UDP 123
Oct 17 19:51:34 invader34 ntpd[2159]: peers refreshed
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0] adj 8 to xeth2
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.1 adj 8
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() R2 prefix 10.0.0.1/32 adj 8 add: call fe1 hooks, addDelReachable
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 IsMpAdj false
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 replaceWithMoreSpecific lr {5 map[]} r {8 map[]}
Oct 17 19:51:38 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.0.0.1/32 adj 8 IsMpAdj false finished: r {8 map[]}
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ethernet.(*ipNeighborMain).AddDelIpNeighbor() call AddDelRoute to add [10 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0] adj 7 to xeth1
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).addDelRoute() addDelRoute add 10.0.0.2 adj 7
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() default prefix 10.0.0.2/32 adj 7 add: call fe1 hooks, addDelReachable
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 replaceWithMoreSpecific lr {3 map[]} r {7 map[]}
Oct 17 19:51:47 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: default 10.0.0.2/32 adj 7 IsMpAdj false finished: r {7 map[]}
```

### Example 3
```
platina@invader34:~$ sudo ip netns exec R2 ip route add 10.5.5.5 via 10.0.0.1
```
```
platina@invader34:~$ goes vnet show ip fib
 Table                   Destination                               Adjacency
     default                   10.0.0.0/24       3: glean xeth1
     default                   10.0.0.1/32       4: local xeth1
     default                   10.0.0.2/32       7: rewrite xeth1 IP4: 50:18:4c:00:0a:44 -> 50:18:4c:00:0a:45
     default                 172.18.0.1/32       2: punt
     default            192.168.101.154/32       2: punt
          R2                   10.0.0.0/24       5: glean xeth2
          R2                   10.0.0.1/32       8: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44
          R2                   10.0.0.2/32       6: local xeth2
          R2                   10.5.5.5/32       9: rewrite xeth2 IP4: 50:18:4c:00:0a:45 -> 50:18:4c:00:0a:44 adj-range 9-9, weight 1 nh-adj 8
```
```
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Main).AddDelRouteNextHop() call addDelRouteNextHop R2 prefix 10.5.5.5/32 oldAdj 0(AdjMiss) add nh [10 0 0 1] from AddDelRouteNextHop
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).AddDelNextHop() add adj 8 to 0(AdjMiss) oldNhs:, nnh 0, newNhs before resolve  8 weight 1; ...
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).createMpAdj() given nhs: 8 weight 1;
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).createMpAdj() resolved nhs: 8 weight 1;
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).createMpAdj() normalized nhs: 8 weight 1;
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).createMpAdj() create new block, adj 9
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).addDelHelper() multipathAdj 9 referenceCount++ 1
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelRouteNextHop() add: prefix 10.5.5.5/32, oldAdj 0(AdjMiss), newAdj 9
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDel() R2 prefix 10.5.5.5/32 adj 9 add: call fe1 hooks, addDelReachable
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.5.5.5/32 adj 9 IsMpAdj true
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).addDelReachable() add: R2 10.5.5.5/32 adj 9 IsMpAdj true finished: r {9 map[]}
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*mapFibResult).addDelNextHop() isDel false id {[10 0 0 1] 1} ip {{[10 5 5 5] 32} 1}: before &{8 map[]}
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*mapFibResult).addDelNextHop() after &{8 map[{[10 0 0 1] 1}:map[{{[10 5 5 5] 32} 1}:0xc00216c150]]}
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*Fib).setReachable() isDel false prefix 10.5.5.5/32 via 10.0.0.1/32 nha [10 0 0 1] adj 8, new mapFibResult {8 map[{[10 0 0 1] 1}:map[{{[10 5 5 5] 32} 1}:0xc00216c150]]}
```
This example illustrates the use of mapFibResult.nh, as illustrated in function call to addDelNexHop():
```
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip4.(*mapFibResult).addDelNextHop() after &{8 map[{[10 0 0 1] 1}:map[{{[10 5 5 5] 32} 1}:0xc00216c150]]}
```
Basically there is a fib entry for prefix 10.0.0.1/32 when the arp completed.
A mapFibKey is created for this prefix 10.0.0.1/32.  Because it is a /32, the entry it created is a mapFibResult in
```reachable[32][mapFibKey]```
The mapfibResult has adjacency index of 8
```reachable[32][mapFibKey].adj = 8```
and
```reachable[32][mapFibKey].nh = map[{[10 0 0 1] 1}:map[{{[10 5 5 5] 32} 1}:0xc00216c150]]```
The way to read this that prefix 10.0.0.1/32 has an adjacency index 8 that is a rewrite.  Futhermore this prefex has an address 10.0.0.1 in fib.index 1 (i.e. index associated with namespace R2), that is being referenced by prefix 10.5.5.5/32 in fib.index 1 as its next hop.

The fib entry for prefix 10.5.5.5/32 has a mapFibKey for 10.5.5.5/32
```reachable[32][mapFibKey].adj = 9```
but it's ```reachable[32][mapFibKey].nh``` is an empty map, because there is no other prefix that references adjacency index 9.

Note the debug log entry
```
Oct 17 19:57:37 invader34 goes.vnetd[11645]: github.com/platinasystems/go/vnet/ip.(*Main).addDelHelper() multipathAdj 9 referenceCount++ 1
```
Because adj 9, which is a mpAdj index, is referenced once (by prefix 10.5.5.5/32), it has a referenceCount incremented from 0 to 1.  If more entries references adj 9, the referenceCount will keep incrementing.  If one of these entry that references 9 is deleted, or no longer reference adj 9 (e.g. new next hop added), the referenceCount will decrement.  Only when the referenceCount == 0 will the adjacency for adj 9 actually get deleted and the heap entry freed from vnet and fe1.  At that point the adjacency index 9 is free and can be use for the next new adjacency.
