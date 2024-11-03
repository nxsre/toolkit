package cpufs

import (
	"context"
	"hash/fnv"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

// Set file owners to the current user,
// otherwise in OSX, we will fail to start.
var uid, gid uint32

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	uid32, _ := strconv.ParseUint(u.Uid, 10, 32)
	gid32, _ := strconv.ParseUint(u.Gid, 10, 32)
	uid = uint32(uid32)
	gid = uint32(gid32)
}

// A tree node in filesystem, it acts as both a directory and file
type Node struct {
	fs.Inode
	path string // File path to get to the current file

	rwMu    sync.RWMutex // Protect file content
	content []byte       // Internal buffer to hold the current file content
}

func (n *Node) OnAdd(ctx context.Context) {
	ch := n.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			Data: []byte("file.txt"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: 2})
	n.AddChild("file.txt", ch, false)
}

// NewRoot returns a file node - acting as a root, with inode sets to 1 and leaf sets to false
func NewRoot() *Node {
	return &Node{}
}

// List keys under a certain prefix from etcd, and output the next hierarchy level
func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	parent := n.resolve("")
	logrus.WithField("path", parent).Debug("Node Readdir")

	entrySet := make(map[string]fuse.DirEntry)
	// 增加一个文件
	log.Println("=========", n.path)
	entrySet["file.txt"] = fuse.DirEntry{Name: "file.txt", Mode: n.getMode(false), Ino: 222}
	entrySet["dira"] = fuse.DirEntry{Name: "diraa", Mode: n.getMode(true), Ino: 111}

	entries := make([]fuse.DirEntry, 0, len(entrySet))
	for _, e := range entrySet {
		entries = append(entries, e)
	}
	return fs.NewListDirStream(entries), fs.OK
}

// Returns next hierarchy level and tells if we have more hierarchies
// path "/foo", parent "/" => "foo"
func (n *Node) nextHierarchyLevel(path, parent string) (string, bool) {
	baseName := strings.TrimPrefix(path, parent)
	hierarchies := strings.SplitN(baseName, string(filepath.Separator), 2)
	return filepath.Clean(hierarchies[0]), len(hierarchies) >= 2
}

// resolve acts as `filepath.Join`, but we want the '/' separator always
func (n *Node) resolve(fileName string) string {
	return n.path + string(filepath.Separator) + fileName
}

// Lookup finds a file under the current node(directory)
func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fullPath := n.resolve(name)
	logrus.WithField("path", fullPath).Debug("Node Lookup")

	child := Node{
		path: fullPath,
	}

	return n.NewInode(ctx, &child, fs.StableAttr{Mode: child.getMode(true), Ino: n.inodeHash(child.path)}), fs.OK
}

func (n *Node) getMode(isLeaf bool) uint32 {
	if isLeaf {
		return 0644 | uint32(syscall.S_IFREG)
	} else {
		return 0755 | uint32(syscall.S_IFDIR)
	}
}

// Getattr outputs file attributes
// TODO: how to invalidate them?
func (n *Node) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	//out.Mode = n.getMode(n.isLeaf)
	log.Println("Getattr", n.IsDir(), n.path, fh)
	//logrus.WithField("path", "------").Debug("Node Getattr")
	out.Size = uint64(len(n.content))
	out.Ino = n.inodeHash(n.path)
	//now := time.Now()
	layout := "2006-01-02 15:04:05"
	value := "2023-12-14 17:09:47"
	now, _ := time.Parse(layout, value)
	out.SetTimes(&now, &now, &now)

	out.Atime = uint64(now.Unix())
	out.Ctime = uint64(now.Unix())
	out.Mtime = uint64(now.Unix())
	out.Atimensec = uint32(now.UnixNano())
	out.Ctimensec = uint32(now.UnixNano())
	out.Mtimensec = uint32(now.UnixNano())

	out.Uid = uid
	out.Gid = gid
	log.Println(out.Ctimensec, out.Ctime)
	return fs.OK
}

// Hash file path into inode number, so we can ensure the same file always gets the same inode number
func (n *Node) inodeHash(path string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(path))
	return h.Sum64()
}

var (
	_ fs.NodeGetattrer = &Node{}
	_ fs.NodeReaddirer = &Node{}
	_ fs.NodeLookuper  = &Node{}
	_ fs.NodeOnAdder   = &Node{}
)
