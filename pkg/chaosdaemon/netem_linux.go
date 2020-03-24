// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package chaosdaemon

import (
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/vishvananda/netlink"
)

func applyNetem(netem *pb.Netem, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("NetemApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	return applyQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return netlink.NewNetem(netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		}, ToNetlinkNetemAttrs(netem))
	})
}

func deleteNetem(netem *pb.Netem, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("NetemCancelError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	return deleteQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return &netlink.Netem{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    netlink.MakeHandle(1, 0),
				Parent:    netlink.HANDLE_ROOT,
			},
		}
	})
}
