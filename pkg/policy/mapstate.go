// Copyright 2016-2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/policy/trafficdirection"
)

var (
	// localHostKey represents an ingress L3 allow from the local host.
	localHostKey = Key{
		Identity:         identity.ReservedIdentityHost.Uint32(),
		TrafficDirection: trafficdirection.Ingress.Uint8(),
	}

	// worldKey represents an ingress L3 allow from the world.
	worldKey = Key{
		Identity:         identity.ReservedIdentityWorld.Uint32(),
		TrafficDirection: trafficdirection.Ingress.Uint8(),
	}
)

// MapState is a state of a policy map.
type MapState map[Key]MapStateEntry

// Key is the userspace representation of a policy key in BPF. It is
// intentionally duplicated from pkg/maps/policymap to avoid pulling in the
// BPF dependency to this package.
type Key struct {
	// Identity is the numeric identity to / from which traffic is allowed.
	Identity uint32
	// DestPort is the port at L4 to / from which traffic is allowed, in
	// host-byte order.
	DestPort uint16
	// NextHdr is the protocol which is allowed.
	Nexthdr uint8
	// TrafficDirection indicates in which direction Identity is allowed
	// communication (egress or ingress).
	TrafficDirection uint8
}

// MapStateEntry is the configuration associated with a Key in a
// MapState. This is a minimized version of policymap.PolicyEntry.
type MapStateEntry struct {
	// The proxy port, in host byte order.
	// If 0 (default), there is no proxy redirection for the corresponding
	// Key.
	ProxyPort uint16
}

// DetermineAllowFromWorld determines whether world should be allowed to
// communicate with the endpoint, based on legacy Cilium 1.0 behaviour. It
// inserts the Key corresponding to the world in the desiredPolicyKeys
// if the legacy mode is enabled.
//
// This must be run after DetermineAllowLocalhost().
//
// For more information, see https://cilium.link/host-vs-world
func (keys MapState) DetermineAllowFromWorld() {

	_, localHostAllowed := keys[localHostKey]
	if option.Config.HostAllowsWorld && localHostAllowed {
		keys[worldKey] = MapStateEntry{}
	}
}

// DetermineAllowLocalhost determines whether communication should be allowed to
// the localhost. It inserts the Key corresponding to the localhost in
// the desiredPolicyKeys if the endpoint is allowed to communicate with the
// localhost.
func (keys MapState) DetermineAllowLocalhost(l4Policy *L4Policy) {

	if option.Config.AlwaysAllowLocalhost() || (l4Policy != nil && l4Policy.HasRedirect()) {
		keys[localHostKey] = MapStateEntry{}
	}
}
