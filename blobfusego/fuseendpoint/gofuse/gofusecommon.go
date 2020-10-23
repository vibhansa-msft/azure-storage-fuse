// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gofuse

import (
	"context"
	"sync/atomic"
	"syscall"
	"time"

	FSIntf "github.com/blobfusego/fswrapper/fsinterface"
	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

//  Refer for Stat_t : https://golang.org/pkg/syscall/#Stat_t

// gofuseNode : Holds data for each node exposed to fuse driver
type gofuseNode struct {
	fs.Inode
	nodePath string          // Full path of the object
	nodeID   uint64          // NodeId assigned to this object
	stat     *syscall.Stat_t // Stats cached for this node
}

// gofuseNodeMap : global map to hold name to node mapping
var gofuseNodeMap = make(map[string]*gofuseNode)

// nextID : Function to assign unique id to each node
func (fd *gofuseFD) nextID() uint64 {
	if fd == nil {
		return 1
	}

	return atomic.AddUint64(&fd.nodeID, 1)
}

// NewGofuseRoot : Creates global object to represent the FS root
func NewGofuseRoot() (fs.InodeEmbedder, error) {
	rootId := instance.nextID()
	n := &gofuseNode{
		nodePath: "",
		nodeID:   rootId,
		stat: &syscall.Stat_t{
			Mode:  uint32(Config.BlobfuseConfig.DefaultPerm),
			Ino:   rootId,
			Size:  4096,
			Nlink: 2,
		},
	}
	n.stat.Mtim.Sec = time.Now().Unix()
	n.stat.Atim = n.stat.Mtim
	n.stat.Ctim = n.stat.Mtim
	gofuseNodeMap[""] = n

	return n, nil
}

// NewGofuseNode : Create new node for given file/dir in the system
func NewGofuseNode(path string, attr FSIntf.BlobAttr) *gofuseNode {
	nodeID := instance.nextID()

	n := &gofuseNode{
		nodePath: path,
		nodeID:   nodeID,
		stat: &syscall.Stat_t{
			Mode: uint32(Config.BlobfuseConfig.DefaultPerm),
			Ino:  nodeID,
		},
	}
	n.refreshAttr(attr)

	gofuseNodeMap[path] = n
	return n
}

// refreshAttr : Reset the values of node attributes from blob attr
func (n *gofuseNode) refreshAttr(attr FSIntf.BlobAttr) {
	n.stat.Size = int64(attr.Size)
	n.stat.Mtim.Sec = attr.Modtime.Unix()
	n.stat.Mtim.Nsec = attr.Modtime.UnixNano()
	n.stat.Atim = n.stat.Mtim
	n.stat.Ctim = n.stat.Mtim
}

// root : Converts INode Embedder to gofuseNode for the root of FS
func (n *gofuseNode) root() *gofuseNode {
	return instance.rootFD.(*gofuseNode)
}

// path : Return path of the given nodes
func (n *gofuseNode) path() string {
	return n.nodePath
}

var _ = (fs.NodeStatfser)((*gofuseNode)(nil))
var _ = (fs.NodeStatfser)((*gofuseNode)(nil))
var _ = (fs.NodeLookuper)((*gofuseNode)(nil))
var _ = (fs.NodeGetattrer)((*gofuseNode)(nil))

// Statfs : Return statistics of the tmp path as we do not have stats of mounted path as of now
func (n *gofuseNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	Logger.LogDebug("FD : Statfs called for " + n.nodePath)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(*Config.BlobfuseConfig.TmpPath, &stat); err != nil {
		Logger.LogErr("FD : Failed to do stat on root")
		return fs.ToErrno(err)
	}

	out.FromStatfsT(&stat)
	return fs.OK
}

// Getattr : Return attributes of the given object
// 1. ignore object if its present in ignore list
// 2. If node mpa has the cached values return from the cache
// 3. Get attributes from client and populate the fuse structures
func (n *gofuseNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	Logger.LogDebug("FD : Attr called for %s", n.path())

	if _, ignore := ignoreList[n.path()]; ignore {
		Logger.LogDebug("FD : Ignoring %s", n.path)
		return syscall.ENOENT
	}

	if nod, found := gofuseNodeMap[n.path()]; found {
		out.FromStat(nod.stat)
		return fs.OK
	}

	attr, err := instance.client.GetAttr(n.path())
	if err != nil {
		Logger.LogErr("FD : Failed to get attribute %s (%s)", n.path(), err)
		return fs.ToErrno(err)
	}

	nod := NewGofuseNode(attr.Name, attr)
	out.FromStat(nod.stat)

	return 0
}
