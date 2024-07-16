package netpdu

const EthMacLen = 6

const (
	ETH_P_8021Q   = 0x8100
	ETH_P_8021AD  = 0x88a8
	ETH_P_ARP     = 0x806
	ETH_P_IP      = 0x800
	ETH_P_IPV6    = 0x86dd
	ETH_P_MPLS_UC = 0x8847
	ETH_P_MPLS_MC = 0x8848
)

const (
	IPPROTO_ICMP   = 0x1
	IPPROTO_ICMPV6 = 0x3a
	IPPROTO_TCP    = 0x6
	IPPROTO_UDP    = 0x11
)

const (
	IPv4len = 4
	IPv6len = 16
)

const (
	ICMPTypeEchoReply = iota
	_
	_
	ICMPTypeDestinationUnreachable
	ICMPTypeSourceQuench
	ICMPTypeRedirectMessage
	_
	_
	ICMPTypeEchoRequest
	ICMPTypeRouterAdvertisement
	ICMPTypeRouterSolicitation
	ICMPTypeTimeExceeded
	ICMPTypeInvalidParamter
	ICMPTypeTimestampRequest
	ICMPTypeTimestampReply
	ICMPTypeInformationRequest
	ICMPTypeInformationReply
	ICMPTypeAddressMaskRequest
	ICMPTypeAddressMaskReeply
)

const (
	ICMPTypeTraceRoute          = 30
	ICMPTypeExtendedEchoRequest = 42
	ICMPTypeExtendedEchoReply   = 42
)

const (
	ICMPUnreachableCodeNetworkUnreachable = iota
	ICMPUnreachableCodeHostUnreachable
	ICMPUnreachableCodeProtocolUnreachable
	ICMPUnreachableCodePortUnreachable
	ICMPUnreachableCodeFragmentationRequired
	ICMPUnreachableCodeSourceRouteFailed
	ICMPUnreachableCodeNetworkUnknown
	ICMPUnreachableCodeHostUnknown
	ICMPUnreachableCodeSourceHostIsolated
	ICMPUnreachableCodeNetworkAdministrativelyProhibited
	ICMPUnreachableCodeHostAdministrativelyProhibited
	ICMPUnreachableCodeTPSNetworkUnreachable
	ICMPUnreachableCodeTPSHostUnreachable
	ICMPUnreachableCodeCommunicationAdministrativelyProhibited
	ICMPUnreachableCodeHostPrecedenceViolation
	ICMPUnreachableCodePrecedenceCutoff
)

const (
	ICMPRedirectCodeNetwork = iota
	ICMPRedirectCodeHost
	ICMPRedirectCodeTOSNetwork
	ICMPRedirectCodeTOSHost
)

const (
	ICMPTimeExceededCodeTTL = iota
	ICMPTimeExceededCodeFragmentReassembly
)

const (
	ICMPInvalidParameterCodePointer = iota
	ICMPInvalidParameterCodeMissingOption
	ICMPInvalidParameterCodeLength
)

const (
	ICMPExtendedEchoReplyCodeNoError = iota
	ICMPExtendedEchoReplyCodeMalformedQuery
	ICMPExtendedEchoReplyCodeNoSuchInterface
	ICMPExtendedEchoReplyCodeNoSuchTableEntry
	ICMPExtendedEchoReplyCodeAmbiguousInterface
)

const (
	ICMP6TypeDestinationUnreachable                = 1
	ICMP6TypePacketTooBig                          = 2
	ICMP6TypeTimeExceeded                          = 3
	ICMP6TypeInvalidParameter                      = 4
	ICMP6TypeEchoRequest                           = 128
	ICMP6TypeEchoReply                             = 129
	ICMP6TypeMulticastListenerQuery                = 130
	ICMP6TypeMulticastListenerReport               = 131
	ICMP6TypeMulticastListenerDone                 = 132
	ICMP6TypeRouterSolicitation                    = 133
	ICMP6TypeRouterAdvertisement                   = 134
	ICMP6TypeNeighborSolicitation                  = 135
	ICMP6TypeNeighborAdvertisement                 = 136
	ICMP6TypeRedirectMessage                       = 137
	ICMP6TypeRouterRenumbering                     = 138
	ICMP6TypeNodeInformationQuery                  = 139
	ICMP6TypeNodeInformationResponse               = 140
	ICMP6TypeInverseNeighborDiscoverySolicitation  = 141
	ICMP6TypeInverseNeighborDiscoveryAdvertisement = 142
	ICMP6TypeMulticastListenerDiscovery            = 143
	ICMP6TypeHomeAgentAddressDiscoveryRequest      = 144
	ICMP6TypeHomeAgentAddressDiscoveryReply        = 145
	ICMP6TypeMobilePrefixSolicitation              = 146
	ICMP6TypeMobilePrefixAdvertisement             = 147
	ICMP6TypeCertificationPathSolicitation         = 148
	ICMP6TypeCertificationPathAdvertisement        = 149
	ICMP6TypeMulticastRouterAdvertisement          = 151
	ICMP6TypeMulticastRouterSolicitation           = 152
	ICMP6TypeMulticastRouterTermination            = 153
	ICMP6TypeRPLControlMessage                     = 155
)

const (
	ICMP6UnreachableCodeNoRouteToDestination = iota
	ICMP6UnreachableCodeCommunicationAdministrativelyProhibited
	ICMP6UnreachableCodeBeyondScopeOfSourceAddress
	ICMP6UnreachableCodeAddressUnreachable
	ICMP6UnreachableCodePortUnreachable
	ICMP6UnreachableCodeSourceAddressFailedIngressOrEgressPolicy
	ICMP6UnreachableCodeRejectRouteToDestination
	ICMP6UnreachableCodeErrorInSourceRoutingHeader
)

const (
	ICMP6TimeExceededCodeTransmitHopLimit = iota
	ICMP6TimeExceededCodeFragmentReassembly
)

const (
	ICMP6InvalidParameterCodeErroneousHeader = iota
	ICMP6InvalidParameterCodeUnrecognizedNextHeader
	ICMP6InvalidParameterCodeUnrecognizedOption
)

const (
	ICMP6RouterRenumberingCodeCommand = iota
	ICMP6RouterRenumberingCodeResult
)
