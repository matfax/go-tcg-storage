// Copyright (c) 2022 by library authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	plugins "github.com/matfax/go-tcg-storage/pkg/cli"
	"github.com/matfax/go-tcg-storage/pkg/core/hash"

	"github.com/alecthomas/kong"
	"github.com/matfax/go-tcg-storage/pkg/core"
	"github.com/matfax/go-tcg-storage/pkg/locking"
	// TODO: Move to locking API when it has MBR functions
)

var (
	programName = "sedlockctl"
	programDesc = "Go SEDlock control (temporary name)"
)

func main() {
	// Parse kong flags and sub-commands
	ctx := kong.Parse(&cli,
		kong.Name(programName),
		kong.Description(programDesc),
		kong.UsageOnError(),
		kong.Resolvers(plugins.ResolvePassword()),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	// Set up connection and initialize session to device.
	coreObj, err := core.NewCore(cli.Device)
	if err != nil {
		log.Fatalf("drive.Open: %v", err)
	}
	defer func(coreObj *core.Core) {
		_ = coreObj.Close()
	}(coreObj)

	snRaw, err := coreObj.DriveIntf.SerialNumber()
	if err != nil {
		log.Fatalf("drive.SerialNumber: %v", err)
	}
	sn := string(snRaw)

	spin := []byte{}
	if cli.Sidpin != "" {
		switch cli.Sidhash {
		case "sedutil-dta":
			spin = hash.HashSedutilDTA(cli.Sidpin, sn)
		case "sedutil-sha512":
			spin = hash.HashSedutil512(cli.Sidpin, sn)
		default:
			log.Fatalf("Unknown hash method %q", cli.Sidhash)
		}
	}

	initOps := []locking.InitializeOpt{}
	if len(spin) > 0 {
		initOps = append(initOps, locking.WithAuth(locking.DefaultAdminAuthority(spin)))
	}
	if cli.Sidpinmsid {
		initOps = append(initOps, locking.WithAuth(locking.DefaultAuthorityWithMSID))
	}

	cs, lmeta, err := locking.Initialize(coreObj, initOps...)
	if err != nil {
		log.Fatalf("locking.Initalize: %v", err)
	}
	defer func(cs *core.ControlSession) {
		_ = cs.Close()
	}(cs)

	var auth locking.LockingSPAuthenticator

	var pin []byte
	if cli.Password != "" {
		switch cli.Hash {
		case "sedutil-dta":
			pin = hash.HashSedutilDTA(cli.Password, sn)
		case "sedutil-sha512":
			pin = hash.HashSedutil512(cli.Password, sn)
		default:
			log.Fatalf("Unknown hash method %q", cli.Hash)
		}
	}
	if cli.User != "" {
		var ok bool
		auth, ok = locking.AuthorityFromName(cli.User, pin)
		if !ok {
			log.Fatalf("Authority %q is not known for this device", cli.User)
		}
	} else {
		if len(pin) == 0 {
			auth = locking.DefaultAuthorityWithMSID
		} else {
			auth = locking.DefaultAuthority(pin)
		}
	}

	l, err := locking.NewSession(cs, lmeta, auth)
	if err != nil {
		log.Fatalf("locking.NewSession: %v", err)
	}
	defer func(l *locking.LockingSP) {
		_ = l.Close()
	}(l)

	// Run the command
	err = ctx.Run(&context{session: l})
	ctx.FatalIfErrorf(err)
}
