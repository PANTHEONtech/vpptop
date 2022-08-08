// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.3.5-44-g2c87563
//  VPP:              21.01-rc2~2-g0b374922d~b11
// source: /usr/share/vpp/api/core/mfib_types.api.json

// Package mfib_types contains generated bindings for API file mfib_types.api.
//
// Contents:
//   2 enums
//   1 struct
//
package mfib_types

import (
	"strconv"

	api "git.fd.io/govpp.git/api"
	fib_types "go.pantheon.tech/vpptop/stats/local/binapi/fib_types"
	_ "go.pantheon.tech/vpptop/stats/local/binapi/ip_types"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

// MfibEntryFlags defines enum 'mfib_entry_flags'.
type MfibEntryFlags uint32

const (
	MFIB_API_ENTRY_FLAG_NONE           MfibEntryFlags = 0
	MFIB_API_ENTRY_FLAG_SIGNAL         MfibEntryFlags = 1
	MFIB_API_ENTRY_FLAG_DROP           MfibEntryFlags = 2
	MFIB_API_ENTRY_FLAG_CONNECTED      MfibEntryFlags = 4
	MFIB_API_ENTRY_FLAG_ACCEPT_ALL_ITF MfibEntryFlags = 8
)

var (
	MfibEntryFlags_name = map[uint32]string{
		0: "MFIB_API_ENTRY_FLAG_NONE",
		1: "MFIB_API_ENTRY_FLAG_SIGNAL",
		2: "MFIB_API_ENTRY_FLAG_DROP",
		4: "MFIB_API_ENTRY_FLAG_CONNECTED",
		8: "MFIB_API_ENTRY_FLAG_ACCEPT_ALL_ITF",
	}
	MfibEntryFlags_value = map[string]uint32{
		"MFIB_API_ENTRY_FLAG_NONE":           0,
		"MFIB_API_ENTRY_FLAG_SIGNAL":         1,
		"MFIB_API_ENTRY_FLAG_DROP":           2,
		"MFIB_API_ENTRY_FLAG_CONNECTED":      4,
		"MFIB_API_ENTRY_FLAG_ACCEPT_ALL_ITF": 8,
	}
)

func (x MfibEntryFlags) String() string {
	s, ok := MfibEntryFlags_name[uint32(x)]
	if ok {
		return s
	}
	str := func(n uint32) string {
		s, ok := MfibEntryFlags_name[uint32(n)]
		if ok {
			return s
		}
		return "MfibEntryFlags(" + strconv.Itoa(int(n)) + ")"
	}
	for i := uint32(0); i <= 32; i++ {
		val := uint32(x)
		if val&(1<<i) != 0 {
			if s != "" {
				s += "|"
			}
			s += str(1 << i)
		}
	}
	if s == "" {
		return str(uint32(x))
	}
	return s
}

// MfibItfFlags defines enum 'mfib_itf_flags'.
type MfibItfFlags uint32

const (
	MFIB_API_ITF_FLAG_NONE           MfibItfFlags = 0
	MFIB_API_ITF_FLAG_NEGATE_SIGNAL  MfibItfFlags = 1
	MFIB_API_ITF_FLAG_ACCEPT         MfibItfFlags = 2
	MFIB_API_ITF_FLAG_FORWARD        MfibItfFlags = 4
	MFIB_API_ITF_FLAG_SIGNAL_PRESENT MfibItfFlags = 8
	MFIB_API_ITF_FLAG_DONT_PRESERVE  MfibItfFlags = 16
)

var (
	MfibItfFlags_name = map[uint32]string{
		0:  "MFIB_API_ITF_FLAG_NONE",
		1:  "MFIB_API_ITF_FLAG_NEGATE_SIGNAL",
		2:  "MFIB_API_ITF_FLAG_ACCEPT",
		4:  "MFIB_API_ITF_FLAG_FORWARD",
		8:  "MFIB_API_ITF_FLAG_SIGNAL_PRESENT",
		16: "MFIB_API_ITF_FLAG_DONT_PRESERVE",
	}
	MfibItfFlags_value = map[string]uint32{
		"MFIB_API_ITF_FLAG_NONE":           0,
		"MFIB_API_ITF_FLAG_NEGATE_SIGNAL":  1,
		"MFIB_API_ITF_FLAG_ACCEPT":         2,
		"MFIB_API_ITF_FLAG_FORWARD":        4,
		"MFIB_API_ITF_FLAG_SIGNAL_PRESENT": 8,
		"MFIB_API_ITF_FLAG_DONT_PRESERVE":  16,
	}
)

func (x MfibItfFlags) String() string {
	s, ok := MfibItfFlags_name[uint32(x)]
	if ok {
		return s
	}
	str := func(n uint32) string {
		s, ok := MfibItfFlags_name[uint32(n)]
		if ok {
			return s
		}
		return "MfibItfFlags(" + strconv.Itoa(int(n)) + ")"
	}
	for i := uint32(0); i <= 32; i++ {
		val := uint32(x)
		if val&(1<<i) != 0 {
			if s != "" {
				s += "|"
			}
			s += str(1 << i)
		}
	}
	if s == "" {
		return str(uint32(x))
	}
	return s
}

// MfibPath defines type 'mfib_path'.
type MfibPath struct {
	ItfFlags MfibItfFlags      `binapi:"mfib_itf_flags,name=itf_flags" json:"itf_flags,omitempty"`
	Path     fib_types.FibPath `binapi:"fib_path,name=path" json:"path,omitempty"`
}
