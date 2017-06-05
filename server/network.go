/*

Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server

import (
	"encoding/binary"
	"math/rand"
	"net"
)

func generateIPv4(srcNet *net.IPNet) net.IP {
	randIP := rand.Uint32()
	ones, total := srcNet.Mask.Size()
	bitmask := ((uint32(1) << uint(total-ones)) - 1)
	netAddr := (uint32(srcNet.IP[0]) << 24) | (uint32(srcNet.IP[1]) << 16) | (uint32(srcNet.IP[2]) << 8) | uint32(srcNet.IP[3])
	randIP = (randIP & bitmask) | netAddr

	if randIP&255 == 0 {
		randIP++
	} else if randIP&255 == 255 {
		randIP--
	}

	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, randIP)

	return ip
}
