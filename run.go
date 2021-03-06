package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/conair/btrfs"
	"github.com/giantswarm/conair/nspawn"
)

var (
	flagBind     stringSlice
	flagSnapshot stringSlice
	cmdRun       = &Command{
		Name:    "run",
		Summary: "Run a container",
		Usage:   "[-bind=S] [-snapshot=S] <image> [<container>]",
		Run:     runRun,
		Description: `Run a new container

Example:
conair run base test

You can either bind mount a directory into the container or take a snapshot of a volume that will be deleted with the container.

conair run -bind=/var/data:/data base test
conair run -snapshot=mysnapshot:/data base test
`,
	}
)

func init() {
	cmdRun.Flags.Var(&flagBind, "bind", "Bind mount a directory into the container")
	cmdRun.Flags.Var(&flagSnapshot, "snapshot", "Add a snapshot into the container")
}

func runRun(args []string) (exit int) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Image name missing.")
		return 1
	}

	imagePath := args[0]

	var container string
	if len(args) < 2 {
		// add some hashing here
		container = imagePath
	} else {
		container = args[1]
	}
	containerPath := fmt.Sprintf(".#%s", container)

	fs, _ := btrfs.Init(home)
	if err := fs.Snapshot(imagePath, containerPath, false); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't create filesystem for container.", err)
		return 1
	}

	c := nspawn.Init(container, fmt.Sprintf("%s/%s", home, containerPath))
	if len(flagBind) > 0 {
		c.SetBinds(flagBind)
	}
	if len(flagSnapshot) > 0 {
		c.SetSnapshots(flagSnapshot)
	}

	for _, snap := range c.Snapshots {
		paths := strings.Split(snap, ":")

		if len(paths) < 2 {
			fmt.Fprintln(os.Stderr, "Couldn't create snapshot for container.")
			return 1
		}

		from := fmt.Sprintf(".cnr-snapshot-%s", paths[0])
		to := fmt.Sprintf("%s/%s", containerPath, paths[1])

		if fs.Exists(to) {
			if err := os.Remove(fmt.Sprintf("%s/%s", home, to)); err != nil {
				fmt.Fprintln(os.Stderr, "Couldn't remove existing directory for snapshot.")
				return 1
			}
		}

		if err := fs.Snapshot(from, to, false); err != nil {
			fmt.Fprintln(os.Stderr, "Couldn't create snapshot for container.", err)
			return 1
		}
	}

	if err := c.Enable(); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't enable container.", err)
		return 1
	}

	if err := c.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't start container.", err)
		return 1
	}

	return 0
}
