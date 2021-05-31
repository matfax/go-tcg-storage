// Copyright (c) 2021 by library authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements TCG Storage Core Table operations on Locking SP tables

package table

import (
	"github.com/bluecmd/go-tcg-storage/pkg/core"
	"github.com/bluecmd/go-tcg-storage/pkg/core/stream"
)

var (
	Locking_LockingTable            = TableUID{0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x00, 0x00}
	LockingInfoObj           RowUID = [8]byte{0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00, 0x01}
	EnterpriseLockingInfoObj RowUID = [8]byte{0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00, 0x00}
	MBRControlObj            RowUID = [8]byte{0x00, 0x00, 0x08, 0x03, 0x00, 0x00, 0x00, 0x01}
)

type EncryptSupport uint
type KeysAvailableConds uint

type ResetType uint

const (
	ResetPowerOff ResetType = 0
	ResetHardware ResetType = 1
	ResetHotPlug  ResetType = 2
)

type LockingInfoRow struct {
	UID                  RowUID
	Name                 *string
	Version              *uint32
	EncryptSupport       *EncryptSupport
	MaxRanges            *uint32
	MaxReEncryptions     *uint32
	KeysAvailableCfg     *KeysAvailableConds
	AlignmentRequired    *bool
	LogicalBlockSize     *uint32
	AlignmentGranularity *uint64
	LowestAlignedLBA     *uint64
}

func LockingInfo(s *core.Session) (*LockingInfoRow, error) {
	rowUID := RowUID{}
	if s.ProtocolLevel == core.ProtocolLevelEnterprise {
		copy(rowUID[:], EnterpriseLockingInfoObj[:])
	} else {
		copy(rowUID[:], LockingInfoObj[:])
	}

	val, err := GetFullRow(s, rowUID)
	if err != nil {
		return nil, err
	}
	row := LockingInfoRow{}
	for col, val := range val {
		switch col {
		case "0", "UID":
			v, ok := val.([]byte)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			copy(row.UID[:], v[:8])
		case "1", "Name":
			v, ok := val.([]byte)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := string(v)
			row.Name = &vv
		case "2", "Version":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint32(v)
			row.Version = &vv
		case "3", "EncryptSupport":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := EncryptSupport(v)
			row.EncryptSupport = &vv
		case "4", "MaxRanges":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint32(v)
			row.MaxRanges = &vv
		case "5", "MaxReEncryptions":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint32(v)
			row.MaxReEncryptions = &vv
		case "6", "KeysAvailableCfg":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := KeysAvailableConds(v)
			row.KeysAvailableCfg = &vv
		case "7":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			var vv bool
			if v > 0 {
				vv = true
			}
			row.AlignmentRequired = &vv
		case "8":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint32(v)
			row.LogicalBlockSize = &vv
		case "9":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint64(v)
			row.AlignmentGranularity = &vv
		case "10":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint64(v)
			row.LowestAlignedLBA = &vv
		}
	}
	return &row, nil
}

func Locking_Enumerate(s *core.Session) ([]RowUID, error) {
	return Enumerate(s, Locking_LockingTable)
}

type LockingRow struct {
	UID              RowUID
	Name             *string
	RangeStart       *uint64
	RangeLength      *uint64
	ReadLockEnabled  *bool
	WriteLockEnabled *bool
	ReadLocked       *bool
	WriteLocked      *bool
	LockOnReset      []ResetType
	ActiveKey        *RowUID
	// NOTE: There are more fields in the standards that have been omited
}

func Locking_Get(s *core.Session, row RowUID) (*LockingRow, error) {
	val, err := GetFullRow(s, row)
	if err != nil {
		return nil, err
	}
	lr := LockingRow{}
	for col, val := range val {
		switch col {
		case "0", "UID":
			v, ok := val.([]byte)
			if !ok || len(v) != 8 {
				return nil, core.ErrMalformedMethodResponse
			}
			copy(lr.UID[:], v[:8])
		case "1", "Name":
			v, ok := val.([]byte)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := string(v)
			lr.Name = &vv
		case "3", "RangeStart":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint64(v)
			lr.RangeStart = &vv
		case "4", "RangeLength":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := uint64(v)
			lr.RangeLength = &vv
		case "5", "ReadLockEnabled":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			var vv bool
			if v > 0 {
				vv = true
			}
			lr.ReadLockEnabled = &vv
		case "6", "WriteLockEnabled":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			var vv bool
			if v > 0 {
				vv = true
			}
			lr.WriteLockEnabled = &vv
		case "7", "ReadLocked":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			var vv bool
			if v > 0 {
				vv = true
			}
			lr.ReadLocked = &vv
		case "8", "WriteLocked":
			v, ok := val.(uint)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			var vv bool
			if v > 0 {
				vv = true
			}
			lr.WriteLocked = &vv
		case "9", "LockOnReset":
			vl, ok := val.(stream.List)
			if !ok {
				return nil, core.ErrMalformedMethodResponse
			}
			for _, val := range vl {
				v, ok := val.(uint)
				if !ok {
					return nil, core.ErrMalformedMethodResponse
				}
				lr.LockOnReset = append(lr.LockOnReset, ResetType(v))
			}
		case "10", "ActiveKey":
			v, ok := val.([]byte)
			if !ok || len(v) != 8 {
				return nil, core.ErrMalformedMethodResponse
			}
			vv := RowUID{}
			copy(vv[:], v)
			lr.ActiveKey = &vv
		}
	}
	return &lr, nil
}

func Locking_Set(s *core.Session, row *LockingRow) error {
	mc := NewSetCall(s, row.UID)

	if row.Name != nil {
		mc.StartOptionalParameter(1, "Name")
		mc.Bytes([]byte(*row.Name))
		mc.EndOptionalParameter()
	}

	// TODO: Add these columns
	//mc.StartOptionalParameter(3, "RangeStart")
	//mc.StartOptionalParameter(4, "RangeLength")
	//mc.StartOptionalParameter(5, "ReadLockEnabled")
	//mc.StartOptionalParameter(6, "WriteLockEnabled")

	if row.ReadLockEnabled != nil {
		mc.StartOptionalParameter(5, "ReadLockEnabled")
		mc.Bool(*row.ReadLockEnabled)
		mc.EndOptionalParameter()
	}
	if row.WriteLockEnabled != nil {
		mc.StartOptionalParameter(6, "WriteLockEnabled")
		mc.Bool(*row.WriteLockEnabled)
		mc.EndOptionalParameter()
	}
	if row.ReadLocked != nil {
		mc.StartOptionalParameter(7, "ReadLocked")
		mc.Bool(*row.ReadLocked)
		mc.EndOptionalParameter()
	}

	if row.WriteLocked != nil {
		mc.StartOptionalParameter(8, "WriteLocked")
		mc.Bool(*row.WriteLocked)
		mc.EndOptionalParameter()
	}

	// TODO: Add these columns
	//mc.StartOptionalParameter(8, "WriteLocked")
	//mc.StartOptionalParameter(9, "LockOnReset")
	//mc.StartOptionalParameter(10, "ActiveKey")

	FinishSetCall(s, mc)
	_, err := s.ExecuteMethod(mc)
	return err
}

type MBRControl struct {
	Enable         *bool
	Done           *bool
	MBRDoneOnReset *[]ResetType
}

func MBRControl_Set(s *core.Session, row *MBRControl) error {
	mc := NewSetCall(s, MBRControlObj)

	if row.Enable != nil {
		mc.StartOptionalParameter(1, "Enable")
		mc.Bool(*row.Enable)
		mc.EndOptionalParameter()
	}
	if row.Done != nil {
		mc.StartOptionalParameter(2, "Done")
		mc.Bool(*row.Done)
		mc.EndOptionalParameter()
	}
	if row.MBRDoneOnReset != nil {
		mc.StartOptionalParameter(3, "MBRDoneOnReset")
		mc.StartList()
		for _, x := range *row.MBRDoneOnReset {
			mc.UInt(uint(x))
		}
		mc.EndList()
		mc.EndOptionalParameter()
	}
	FinishSetCall(s, mc)
	_, err := s.ExecuteMethod(mc)
	return err
}
