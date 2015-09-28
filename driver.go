//-----------------------------------------------------------------------------
// Package membership:
//-----------------------------------------------------------------------------

package main

//-----------------------------------------------------------------------------
// Imports:
//-----------------------------------------------------------------------------

import (

	// Standard library:
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	// Community:
	"github.com/calavera/dkvolume"
)

//-----------------------------------------------------------------------------
// Package constant declarations factored into a block:
//-----------------------------------------------------------------------------

const (
	lockID = "dockerLock"
)

//-----------------------------------------------------------------------------
// Package variable declarations factored into a block:
//-----------------------------------------------------------------------------

var (
	nameRegex = regexp.MustCompile(`^(([-_.[:alnum:]]+)/)?([-_.[:alnum:]]+)(@([0-9]+))?$`)
	lockRegex = regexp.MustCompile(`^(client.[0-9]+) ` + lockID)
	commands  = [...]string{"rbd", "mkfs"}
)

//-----------------------------------------------------------------------------
// Structs definitions:
//-----------------------------------------------------------------------------

type rbdDriver struct {
	volRoot   string
	defPool   string
	defFsType string
	defSize   int
	cmd       map[string]string
}

//-----------------------------------------------------------------------------
// initDriver
//-----------------------------------------------------------------------------

func initDriver(volRoot, defPool, defFsType string, defSize int) rbdDriver {

	// Variables
	var err error
	cmd := make(map[string]string)

	// Search for binaries
	for _, i := range commands {
		cmd[i], err = exec.LookPath(i)
		if err != nil {
			log.Fatal("Make sure binary %s is in your PATH", i)
		}
	}

	// Initialize the struct
	driver := rbdDriver{
		volRoot:   volRoot,
		defPool:   defPool,
		defFsType: defFsType,
		defSize:   defSize,
		cmd:       cmd,
	}

	return driver
}

//-----------------------------------------------------------------------------
// POST /VolumeDriver.Create
//
// Request:
//  { "Name": "volume_name" }
//  Instruct the plugin that the user wants to create a volume, given a user
//  specified volume name. The plugin does not need to actually manifest the
//  volume on the filesystem yet (until Mount is called).
//
// Response:
//  { "Err": null }
//  Respond with a string error if an error occurred.
//-----------------------------------------------------------------------------

func (d *rbdDriver) Create(r dkvolume.Request) dkvolume.Response {

	// Parse the docker --volume option
	pool, name, size, err := d.parsePoolNameSize(r.Name)
	if err != nil {
		log.Printf("ERROR: parsing volume: %s", err)
		return dkvolume.Response{Err: err.Error()}
	}

	mountpoint := filepath.Join(d.volRoot, pool, name)

	// Create RBD image if not exist
	if exists, err := d.imageExists(pool, name); !exists && err == nil {
		if err = d.createImage(pool, name, d.defFsType, size); err != nil {
			return dkvolume.Response{Err: err.Error()}
		}
	} else if err != nil {
		log.Printf("ERROR: checking for RBD Image: %s", err)
		return dkvolume.Response{Err: err.Error()}
	}

	log.Printf("Mountpoint: %s", mountpoint)

	return dkvolume.Response{}
}

//-----------------------------------------------------------------------------
// POST /VolumeDriver.Remove
//
// Request:
//  { "Name": "volume_name" }
//  Delete the specified volume from disk. This request is issued when a user
//  invokes docker rm -v to remove volumes associated with a container.
//
// Response:
//  { "Err": null }
//  Respond with a string error if an error occurred.
//-----------------------------------------------------------------------------

func (d *rbdDriver) Remove(r dkvolume.Request) dkvolume.Response {
	log.Printf("Remove: %s", r.Name)
	return dkvolume.Response{}
}

//-----------------------------------------------------------------------------
// POST /VolumeDriver.Path
//
// Request:
//  { "Name": "volume_name" }
//  Docker needs reminding of the path to the volume on the host.
//
// Response:
//  { "Mountpoint": "/path/to/directory/on/host", "Err": null }
//  Respond with the path on the host filesystem where the volume has been
//  made available, and/or a string error if an error occurred.
//-----------------------------------------------------------------------------

func (d *rbdDriver) Path(r dkvolume.Request) dkvolume.Response {
	log.Printf("Path: %s", r.Name)
	return dkvolume.Response{Mountpoint: "/path/to/directory/on/host"}
}

//-----------------------------------------------------------------------------
// POST /VolumeDriver.Mount
//
// Request:
//  { "Name": "volume_name" }
//  Docker requires the plugin to provide a volume, given a user specified
//  volume name. This is called once per container start.
//
// Response:
//  { "Mountpoint": "/path/to/directory/on/host", "Err": null }
//  Respond with the path on the host filesystem where the volume has been
//  made available, and/or a string error if an error occurred.
//-----------------------------------------------------------------------------

func (d *rbdDriver) Mount(r dkvolume.Request) dkvolume.Response {
	log.Printf("Mount: %s", r.Name)
	return dkvolume.Response{Mountpoint: "/path/to/directory/on/host"}
}

//-----------------------------------------------------------------------------
// POST /VolumeDriver.Unmount
//
// Request:
//  { "Name": "volume_name" }
//  Indication that Docker no longer is using the named volume. This is called
//  once per container stop. Plugin may deduce that it is safe to deprovision
//  it at this point.
//
// Response:
//  { "Err": null }
//  Respond with a string error if an error occurred.
//-----------------------------------------------------------------------------

func (d *rbdDriver) Unmount(r dkvolume.Request) dkvolume.Response {
	log.Printf("Umount: %s", r.Name)
	return dkvolume.Response{}
}

//-----------------------------------------------------------------------------
// parsePoolNameSize
//-----------------------------------------------------------------------------

func (d *rbdDriver) parsePoolNameSize(src string) (string, string, int, error) {

	sub := nameRegex.FindStringSubmatch(src)

	if len(sub) != 6 {
		return "", "", 0, errors.New("Unable to parse docker --volume option: %s" + src)
	}

	// Set defaults
	pool := d.defPool
	name := sub[3]
	size := d.defSize

	// Pool overwrite
	if sub[2] != "" {
		pool = sub[2]
	}

	// Size overwrite
	if sub[5] != "" {
		var err error
		size, err = strconv.Atoi(sub[5])
		if err != nil {
			size = d.defSize
		}
	}

	return pool, name, size, nil
}

//-----------------------------------------------------------------------------
// imageExists
//-----------------------------------------------------------------------------

func (d *rbdDriver) imageExists(pool, name string) (bool, error) {

	// List RBD images
	out, err := exec.Command(d.cmd["rbd"], "ls", pool).Output()
	if err != nil {
		return false, errors.New("Unable to list images")
	}

	// Parse the output
	list := strings.Split(string(out), "\n")
	for _, item := range list {
		if item == name {
			return true, nil
		}
	}

	return false, nil
}

//-----------------------------------------------------------------------------
// createImage
//-----------------------------------------------------------------------------

func (d *rbdDriver) createImage(pool, name, fstype string, size int) error {

	// Create the image device
	err := exec.Command(
		d.cmd["rbd"], "create",
		"--pool", pool,
		"--size", strconv.Itoa(size),
		name,
	).Run()

	if err != nil {
		return errors.New("Unable to create the image device")
	}

	// Add image lock
	locker, err := d.lockImage(pool, name, lockID)
	if err != nil {
		return err
	}

	// Remove image lock
	err = d.unlockImage(pool, name, lockID, locker)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------
// lockImage
//-----------------------------------------------------------------------------

func (d *rbdDriver) lockImage(pool, name, lockID string) (string, error) {

	// Lock the image
	err := exec.Command(
		d.cmd["rbd"], "lock add",
		"--pool", pool,
		name, lockID,
	).Run()

	if err != nil {
		return "", errors.New("Unable to lock the image")
	}

	// List the locks
	out, err := exec.Command(
		d.cmd["rbd"], "lock list",
		"--pool", pool, name,
	).Output()

	if err != nil {
		return "", errors.New("Unable to list the image locks")
	}

	// Parse the locker ID
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			sub := lockRegex.FindStringSubmatch(line)
			if len(sub) == 2 {
				return sub[1], nil
			}
		}
	}

	return "", errors.New("Unable to parse locker ID")
}

//-----------------------------------------------------------------------------
// unlockImage
//-----------------------------------------------------------------------------

func (d *rbdDriver) unlockImage(pool, name, lockID, locker string) error {

	// Unlock the image
	err := exec.Command(
		d.cmd["rbd"], "lock remove",
		name, lockID, locker,
	).Run()

	if err != nil {
		return errors.New("Unable to unlock the image")
	}

	return nil
}
