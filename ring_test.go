/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*
 * Content before git sha 34fdeebefcbf183ed7f916f931aa0586fdaa1b40
 * Copyright (c) 2016, The Gocql authors,
 * provided under the BSD-3-Clause License.
 * See the NOTICE file distributed with this work for additional information.
 */

package gocql

import (
	"net"
	"testing"
)

func TestRing_AddHostIfMissing_Missing(t *testing.T) {
	ring := &ring{}

	host := &HostInfo{hostId: MustRandomUUID().String(), connectAddress: net.IPv4(1, 1, 1, 1)}
	h1, ok := ring.addHostIfMissing(host)
	if ok {
		t.Fatal("host was reported as already existing")
	} else if !h1.Equal(host) {
		t.Fatalf("hosts not equal that are returned %v != %v", h1, host)
	} else if h1 != host {
		t.Fatalf("returned host same pointer: %p != %p", h1, host)
	}
}

func TestRing_AddHostIfMissing_Existing(t *testing.T) {
	ring := &ring{}

	host := &HostInfo{hostId: MustRandomUUID().String(), connectAddress: net.IPv4(1, 1, 1, 1)}
	ring.addHostIfMissing(host)

	h2 := &HostInfo{hostId: host.hostId, connectAddress: net.IPv4(2, 2, 2, 2)}

	h1, ok := ring.addHostIfMissing(h2)
	if !ok {
		t.Fatal("host was not reported as already existing")
	} else if !h1.Equal(host) {
		t.Fatalf("hosts not equal that are returned %v != %v", h1, host)
	} else if h1 != host {
		t.Fatalf("returned host same pointer: %p != %p", h1, host)
	}
}

// TestRing_GetHostByIP_StaleMapping verifies getHostByIP never reports a host as found while returning a nil
// *HostInfo. hostIPToUUID is keyed by nodeToNodeAddress() as computed at insertion time, but HostInfo.update
// fills in nil address fields, so the key a host would be deleted under can differ from the key it was
// inserted under. removeHost then drops hosts[hostID] and leaves the original mapping behind.
func TestRing_GetHostByIP_StaleMapping(t *testing.T) {
	ring := &ring{}

	// Discovered from system.peers, so only the peer address is known and it is the ip->uuid key.
	host := &HostInfo{
		hostId:         MustRandomUUID().String(),
		connectAddress: net.IPv4(1, 1, 1, 1),
		peer:           net.IPv4(1, 1, 1, 1),
	}
	ring.addHostIfMissing(host)

	// A later read supplies the broadcast address. update fills it because the field is nil, and
	// nodeToNodeAddress now prefers it over the peer, so the host's key has moved.
	host.update(&HostInfo{broadcastAddress: net.IPv4(2, 2, 2, 2)})

	ring.removeHost(host.HostID())

	h, ok := ring.getHostByIP("1.1.1.1")
	if ok && h == nil {
		t.Fatal("getHostByIP reported a host as found but returned nil")
	}
}
