// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package clab

import (
	border0 "github.com/srl-labs/containerlab/nodes/border0"
	bridge "github.com/srl-labs/containerlab/nodes/bridge"
	c8000 "github.com/srl-labs/containerlab/nodes/c8000"
	ceos "github.com/srl-labs/containerlab/nodes/ceos"
	checkpoint_cloudguard "github.com/srl-labs/containerlab/nodes/checkpoint_cloudguard"
	crpd "github.com/srl-labs/containerlab/nodes/crpd"
	cvx "github.com/srl-labs/containerlab/nodes/cvx"
	ext_container "github.com/srl-labs/containerlab/nodes/ext_container"
	host "github.com/srl-labs/containerlab/nodes/host"
	ipinfusion_ocnos "github.com/srl-labs/containerlab/nodes/ipinfusion_ocnos"
	keysight_ixiacone "github.com/srl-labs/containerlab/nodes/keysight_ixiacone"
	linux "github.com/srl-labs/containerlab/nodes/linux"
	ovs "github.com/srl-labs/containerlab/nodes/ovs"
	rare "github.com/srl-labs/containerlab/nodes/rare"
	sonic "github.com/srl-labs/containerlab/nodes/sonic"
	srl "github.com/srl-labs/containerlab/nodes/srl"
	vr_aoscx "github.com/srl-labs/containerlab/nodes/vr_aoscx"
	vr_csr "github.com/srl-labs/containerlab/nodes/vr_csr"
	vr_ftosv "github.com/srl-labs/containerlab/nodes/vr_ftosv"
	vr_n9kv "github.com/srl-labs/containerlab/nodes/vr_n9kv"
	vr_nxos "github.com/srl-labs/containerlab/nodes/vr_nxos"
	vr_pan "github.com/srl-labs/containerlab/nodes/vr_pan"
	vr_ros "github.com/srl-labs/containerlab/nodes/vr_ros"
	vr_sros "github.com/srl-labs/containerlab/nodes/vr_sros"
	vr_veos "github.com/srl-labs/containerlab/nodes/vr_veos"
	vr_vmx "github.com/srl-labs/containerlab/nodes/vr_vmx"
	vr_vqfx "github.com/srl-labs/containerlab/nodes/vr_vqfx"
	vr_vsrx "github.com/srl-labs/containerlab/nodes/vr_vsrx"
	vr_vjunosswitch "github.com/srl-labs/containerlab/nodes/vr_vjunosswitch"
	vr_xrv "github.com/srl-labs/containerlab/nodes/vr_xrv"
	vr_xrv9k "github.com/srl-labs/containerlab/nodes/vr_xrv9k"
	xrd "github.com/srl-labs/containerlab/nodes/xrd"
)

// RegisterNodes registers all the nodes/kinds supported by containerlab.
func (c *CLab) RegisterNodes() {
	bridge.Register(c.Reg)
	ceos.Register(c.Reg)
	checkpoint_cloudguard.Register(c.Reg)
	crpd.Register(c.Reg)
	cvx.Register(c.Reg)
	ext_container.Register(c.Reg)
	host.Register(c.Reg)
	ipinfusion_ocnos.Register(c.Reg)
	keysight_ixiacone.Register(c.Reg)
	linux.Register(c.Reg)
	ovs.Register(c.Reg)
	sonic.Register(c.Reg)
	srl.Register(c.Reg)
	vr_aoscx.Register(c.Reg)
	vr_csr.Register(c.Reg)
	vr_ftosv.Register(c.Reg)
	vr_n9kv.Register(c.Reg)
	vr_nxos.Register(c.Reg)
	vr_pan.Register(c.Reg)
	vr_ros.Register(c.Reg)
	vr_sros.Register(c.Reg)
	vr_veos.Register(c.Reg)
	vr_vmx.Register(c.Reg)
	vr_vsrx.Register(c.Reg)
	vr_vqfx.Register(c.Reg)
	vr_vjunosswitch.Register(c.Reg)
	vr_xrv.Register(c.Reg)
	vr_xrv9k.Register(c.Reg)
	xrd.Register(c.Reg)
	rare.Register(c.Reg)
	c8000.Register(c.Reg)
	border0.Register(c.Reg)
}
