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
	"sync"
	"testing"
)

func TestEventDebounce(t *testing.T) {
	const eventCount = 150
	wg := &sync.WaitGroup{}
	wg.Add(1)

	eventsSeen := 0
	debouncer := newEventDebouncer("testDebouncer", func(events []frame) {
		defer wg.Done()
		eventsSeen += len(events)
	}, &defaultLogger{})
	defer debouncer.stop()

	for i := 0; i < eventCount; i++ {
		debouncer.debounce(&statusChangeEventFrame{
			change: "UP",
			host:   net.IPv4(127, 0, 0, 1),
			port:   9042,
		})
	}

	wg.Wait()
	if eventCount != eventsSeen {
		t.Fatalf("expected to see %d events but got %d", eventCount, eventsSeen)
	}
}

// TestSession_HandleNodeDown_StaleMapping verifies handleNodeDown survives an ip->uuid mapping that points at
// a host no longer in the ring. getHostByIP takes its found flag from hostIPToUUID but its value from hosts,
// so it can report found with a nil host, and handleNodeDown dereferences it. hostConnPool.fillingStopped
// reaches this path on every failed pool fill, so a stale mapping crashes the process.
func TestSession_HandleNodeDown_StaleMapping(t *testing.T) {
	s := &Session{}

	host := &HostInfo{
		hostId:         MustRandomUUID().String(),
		connectAddress: net.IPv4(1, 1, 1, 1),
		peer:           net.IPv4(1, 1, 1, 1),
	}
	s.ring.addHostIfMissing(host)
	host.update(&HostInfo{broadcastAddress: net.IPv4(2, 2, 2, 2)})
	s.ring.removeHost(host.HostID())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handleNodeDown panicked on a stale ip->uuid mapping: %v", r)
		}
	}()

	s.handleNodeDown(net.IPv4(1, 1, 1, 1), 9042)
}
