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
	"regexp"
	"strconv"

	// Community:
	"github.com/calavera/dkvolume"
)

//-----------------------------------------------------------------------------
// Structs definitions:
//-----------------------------------------------------------------------------

type rbdImage struct {
	size   int
	fsType string
}

type rbdDriver struct {
	volRoot   string
	defPool   string
	defSize   int
	defFsType string
	images    map[string]map[string]rbdImage
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

	log.Printf("[POST] /VolumeDriver.Create")

	// Parse the docker --volume option:
	if err := d.parsePoolNameSize(r.Name); err != nil {
		log.Printf("ERROR: parsing volume: %s", err)
	}

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

func (d *rbdDriver) parsePoolNameSize(src string) error {

	img := new(rbdImage)
	reg := regexp.MustCompile(`^(([-_.[:alnum:]]+)/)?([-_.[:alnum:]]+)(@([0-9]+))?$`)
	sub := reg.FindStringSubmatch(src)

	if len(sub) != 6 {
		return errors.New("Unable to parse docker --volume option: %s" + src)
	}

	// Default pool:
	pool := d.defPool
	if sub[2] != "" {
		pool = sub[2]
	}

	// Name:
	name := sub[3]

	// Default size:
	img.size = d.defSize
	if sub[5] != "" {
		img.size, _ = strconv.Atoi(sub[5])
	}

	// Default fsType:
	img.fsType = d.defFsType

	// Save:
	if d.images[pool] == nil {
		d.images[pool] = make(map[string]rbdImage)
	}

	d.images[pool][name] = *img

	// Return on success:
	return nil
}
