package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/conair/btrfs"
	"github.com/giantswarm/conair/networkd"
	"github.com/giantswarm/conair/nspawn"
)

var cmdBootstrap = &Command{
	Name:        "bootstrap",
	Description: "Bootstrap conair base image",
	Summary:     "Creates an arch rootfs with pacstrap. If there is no pacstrap on your system use 'conair pull base' instead",
	Run:         runBootstrap,
}

func runBootstrap(args []string) (exit int) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Image name missing.")
		return 1
	}

	image := args[0]
	imagePath := fmt.Sprintf("machines/%s", image)

	fs, err := btrfs.Init(home)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't populate filesystem for conair.", err)
		return 1
	}

	err = fs.Subvolume(imagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Couldn't create subvolume for image %s.", image), err)
		return 1
	}

	err = nspawn.CreateImage(image, getImagesPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't create image.", err)
		return 1
	}

	err = networkd.CreateClientNetwork(fmt.Sprintf("%s/%s", getImagesPath(), image), destination)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't add networking to new image.", err)
		return 1
	}
	return 0
}
