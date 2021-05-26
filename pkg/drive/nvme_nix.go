// Copyright (c) 2021 by library authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package drive

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/bluecmd/go-tcg-storage/pkg/drive/ioctl"
)

const (
	NVME_ADMIN_IDENTIFY = 0x06
	NVME_SECURITY_SEND  = 0x81
	NVME_SECURITY_RECV  = 0x82
)

var (
	NVME_IOCTL_ADMIN_CMD = ioctl.Iowr('N', 0x41, unsafe.Sizeof(nvmePassthruCommand{}))
)

// Defined in <linux/nvme_ioctl.h>
type nvmePassthruCommand struct {
	opcode       uint8
	flags        uint8
	rsvd1        uint16
	nsid         uint32
	cdw2         uint32
	cdw3         uint32
	metadata     uint64
	addr         uint64
	metadata_len uint32
	data_len     uint32
	cdw10        uint32
	cdw11        uint32
	cdw12        uint32
	cdw13        uint32
	cdw14        uint32
	cdw15        uint32
	timeout_ms   uint32
	result       uint32
}

type nvmeAdminCommand nvmePassthruCommand

type nvmeDrive struct {
	fd uintptr
}

func (d *nvmeDrive) IFRecv(proto SecurityProtocol, sps uint16, data *[]byte) error {
	cmd := nvmeAdminCommand{
		opcode:   NVME_SECURITY_RECV,
		nsid:     0,
		addr:     uint64(uintptr(unsafe.Pointer(&(*data)[0]))),
		data_len: uint32(len(*data)),
		cdw10:    uint32(proto&0xff)<<24 | uint32(sps)<<8,
		cdw11:    uint32(len(*data)),
	}

	return ioctl.Ioctl(d.fd, NVME_IOCTL_ADMIN_CMD, uintptr(unsafe.Pointer(&cmd)))
}

func (d *nvmeDrive) IFSend(proto SecurityProtocol, sps uint16, data []byte) error {
	cmd := nvmeAdminCommand{
		opcode:   NVME_SECURITY_SEND,
		nsid:     0,
		addr:     uint64(uintptr(unsafe.Pointer(&data[0]))),
		data_len: uint32(len(data)),
		cdw10:    uint32(proto&0xff)<<24 | uint32(sps)<<8,
		cdw11:    uint32(len(data)),
	}

	return ioctl.Ioctl(d.fd, NVME_IOCTL_ADMIN_CMD, uintptr(unsafe.Pointer(&cmd)))
}

func (d *nvmeDrive) Identify() (string, error) {
	i, err := identifyNvme(d.fd)
	if err != nil {
		return "", err
	}
	return i.String(), err
}

func (d *nvmeDrive) Close() error {
	return os.NewFile(d.fd, "").Close()
}

func NVMEDrive(fd FdIntf) *nvmeDrive {
	return &nvmeDrive{fd: fd.Fd()}
}

type nvmeIdentity struct {
	_            uint16 /* Vid */
	_            uint16 /* Ssvid */
	SerialNumber [20]byte
	ModelNumber  [40]byte
	Firmware     [8]byte
}

func (i *nvmeIdentity) String() string {
	return fmt.Sprintf("Protocol=NVMe, Model=%s, Serial=%s, Revision=%s",
		strings.TrimSpace(string(i.ModelNumber[:])),
		strings.TrimSpace(string(i.SerialNumber[:])),
		strings.TrimSpace(string(i.Firmware[:])))
}

func identifyNvme(fd uintptr) (*nvmeIdentity, error) {
	raw := make([]byte, 4096)

	cmd := nvmePassthruCommand{
		opcode:   NVME_ADMIN_IDENTIFY,
		nsid:     0, // Namespace 0, since we are identifying the controller
		addr:     uint64(uintptr(unsafe.Pointer(&raw[0]))),
		data_len: uint32(len(raw)),
		cdw10:    1, // Identify controller
	}

	// TODO: Replace with https://go-review.googlesource.com/c/sys/+/318210/ if accepted
	err := ioctl.Ioctl(fd, NVME_IOCTL_ADMIN_CMD, uintptr(unsafe.Pointer(&cmd)))
	if err != nil {
		return nil, err
	}

	info := nvmeIdentity{}
	buf := bytes.NewBuffer(raw)
	// NVMe *seems* to use little endian, no experience though - but since we are
	// reading byte arrays it matters not.
	if err := binary.Read(buf, binary.LittleEndian, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func isNVME(f FdIntf) bool {
	i, err := identifyNvme(f.Fd())
	return err == nil && i != nil
}
