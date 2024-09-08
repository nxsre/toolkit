/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/cp/cp.go
//
//	"k8s.io/kubectl/pkg/cmd/cp"
package k8sclient

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/exec"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"net/url"
	"os"
	"strings"
	//cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// CopyOptions have the data required to perform the copy operation
type CopyOptions struct {
	Container  string
	Namespace  string
	NoPreserve bool
	MaxTries   int

	ClientConfig      *restclient.Config
	Clientset         kubernetes.Interface
	ExecParentCmdName string

	args []string

	genericiooptions.IOStreams
}

// NewCopyOptions creates the options for copy
func NewCopyOptions(ioStreams genericiooptions.IOStreams) *CopyOptions {
	return &CopyOptions{
		IOStreams: ioStreams,
	}
}

var (
	errFileSpecDoesntMatchFormat = errors.New("filespec must match the canonical format: [[namespace/]pod:]file/path")
)

func extractFileSpec(arg string) (fileSpec, error) {
	i := strings.Index(arg, ":")

	// filespec starting with a semicolon is invalid
	if i == 0 {
		return fileSpec{}, errFileSpecDoesntMatchFormat
	}
	if i == -1 {
		return fileSpec{
			File: newLocalPath(arg),
		}, nil
	}

	pod, file := arg[:i], arg[i+1:]
	pieces := strings.Split(pod, "/")
	switch len(pieces) {
	case 1:
		return fileSpec{
			PodName: pieces[0],
			File:    newRemotePath(file),
		}, nil
	case 2:
		return fileSpec{
			PodNamespace: pieces[0],
			PodName:      pieces[1],
			File:         newRemotePath(file),
		}, nil
	default:
		return fileSpec{}, errFileSpecDoesntMatchFormat
	}
}

// Validate makes sure provided values for CopyOptions are valid
func (o *CopyOptions) Validate() error {
	if len(o.args) != 2 {
		return fmt.Errorf("source and destination are required")
	}
	return nil
}

// Run performs the execution
func (o *CopyOptions) Run(src, dest string) error {
	srcSpec, err := extractFileSpec(src)
	if err != nil {
		return err
	}
	destSpec, err := extractFileSpec(dest)
	if err != nil {
		return err
	}

	if len(srcSpec.PodName) != 0 && len(destSpec.PodName) != 0 {
		return fmt.Errorf("one of src or dest must be a local file specification")
	}
	if len(srcSpec.File.String()) == 0 || len(destSpec.File.String()) == 0 {
		return errors.New("filepath can not be empty")
	}

	if len(srcSpec.PodName) != 0 {
		return o.copyFromPod(srcSpec, destSpec)
	}
	if len(destSpec.PodName) != 0 {
		destSpec.PodNamespace = o.Namespace
		fmt.Printf("%+v  ### %+v\n", srcSpec, destSpec)

		return o.copyToPod(srcSpec, destSpec, &exec.ExecOptions{})
	}
	return fmt.Errorf("one of src or dest must be a remote file specification")
}

// checkDestinationIsDir receives a destination fileSpec and
// determines if the provided destination path exists on the
// pod. If the destination path does not exist or is _not_ a
// directory, an error is returned with the exit code received.
func (o *CopyOptions) checkDestinationIsDir(dest fileSpec) error {
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				Out:    bytes.NewBuffer([]byte{}),
				ErrOut: bytes.NewBuffer([]byte{}),
			},

			Namespace: dest.PodNamespace,
			PodName:   dest.PodName,
		},

		Command:  []string{"test", "-d", dest.File.String()},
		Executor: &exec.DefaultRemoteExecutor{},
	}

	return o.execute(options)
}

func (o *CopyOptions) copyToPod(src, dest fileSpec, options *exec.ExecOptions) error {
	if _, err := os.Stat(src.File.String()); err != nil {
		return fmt.Errorf("%s doesn't exist in local filesystem", src.File)
	}
	reader, writer := io.Pipe()

	srcFile := src.File.(localPath)
	destFile := dest.File.(remotePath)

	if err := o.checkDestinationIsDir(dest); err == nil {
		// If no error, dest.File was found to be a directory.
		// Copy specified src into it
		destFile = destFile.Join(srcFile.Base())
	}

	go func(src localPath, dest remotePath, writer io.WriteCloser) {
		defer writer.Close()
		cmdutil.CheckErr(makeTar(src, dest, writer))
	}(srcFile, destFile, writer)
	var cmdArr []string

	// TODO: Improve error messages by first testing if 'tar' is present in the container?
	if o.NoPreserve {
		cmdArr = []string{"tar", "--no-same-permissions", "--no-same-owner", "-xmf", "-"}
	} else {
		cmdArr = []string{"tar", "-xmf", "-"}
	}
	destFileDir := destFile.Dir().String()
	if len(destFileDir) > 0 {
		cmdArr = append(cmdArr, "-C", destFileDir)
	}

	options.StreamOptions = exec.StreamOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     reader,
			Out:    o.Out,
			ErrOut: o.ErrOut,
		},
		Stdin:     true,
		Namespace: dest.PodNamespace,
		PodName:   dest.PodName,
	}

	options.Command = cmdArr
	//options.Executor = &exec.DefaultRemoteExecutor{}
	options.Executor = &CopyExecutor{}

	return o.execute(options)
}

// DefaultRemoteExecutor is the standard implementation of remote command execution
type CopyExecutor struct {
}

func (r *CopyExecutor) Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}

	return exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}

func (o *CopyOptions) copyFromPod(src, dest fileSpec) error {
	reader := newTarPipe(src, o)
	srcFile := src.File.(remotePath)
	destFile := dest.File.(localPath)
	// remove extraneous path shortcuts - these could occur if a path contained extra "../"
	// and attempted to navigate beyond "/" in a remote filesystem
	prefix := stripPathShortcuts(srcFile.StripSlashes().Clean().String())
	return o.untarAll(src.PodNamespace, src.PodName, prefix, srcFile, destFile, reader)
}

type TarPipe struct {
	src       fileSpec
	o         *CopyOptions
	reader    *io.PipeReader
	outStream *io.PipeWriter
	bytesRead uint64
	retries   int
}

func newTarPipe(src fileSpec, o *CopyOptions) *TarPipe {
	t := new(TarPipe)
	t.src = src
	t.o = o
	t.initReadFrom(0)
	return t
}

func (t *TarPipe) initReadFrom(n uint64) {
	t.reader, t.outStream = io.Pipe()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericclioptions.IOStreams{
				In:     nil,
				Out:    t.outStream,
				ErrOut: t.o.Out,
			},

			Namespace: t.src.PodNamespace,
			PodName:   t.src.PodName,
		},

		// TODO: Improve error messages by first testing if 'tar' is present in the container?
		Command:  []string{"tar", "cf", "-", t.src.File.String()},
		Executor: &exec.DefaultRemoteExecutor{},
	}
	if t.o.MaxTries != 0 {
		options.Command = []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", t.src.File, n)}
	}

	go func() {
		defer t.outStream.Close()
		cmdutil.CheckErr(t.o.execute(options))
	}()
}

func (t *TarPipe) Read(p []byte) (n int, err error) {
	n, err = t.reader.Read(p)
	if err != nil {
		if t.o.MaxTries < 0 || t.retries < t.o.MaxTries {
			t.retries++
			fmt.Printf("Resuming copy at %d bytes, retry %d/%d\n", t.bytesRead, t.retries, t.o.MaxTries)
			t.initReadFrom(t.bytesRead + 1)
			err = nil
		} else {
			fmt.Printf("Dropping out copy after %d retries\n", t.retries)
		}
	} else {
		t.bytesRead += uint64(n)
	}
	return
}

func makeTar(src localPath, dest remotePath, writer io.Writer) error {
	// TODO: use compression here?
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	srcPath := src.Clean()
	destPath := dest.Clean()
	return recursiveTar(srcPath.Dir(), srcPath.Base(), destPath.Dir(), destPath.Base(), tarWriter)
}

func recursiveTar(srcDir, srcFile localPath, destDir, destFile remotePath, tw *tar.Writer) error {
	matchedPaths, err := srcDir.Join(srcFile).Glob()
	if err != nil {
		return err
	}
	for _, fpath := range matchedPaths {
		stat, err := os.Lstat(fpath)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			files, err := ioutil.ReadDir(fpath)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				//case empty directory
				hdr, _ := tar.FileInfoHeader(stat, fpath)
				hdr.Name = destFile.String()
				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
			}
			for _, f := range files {
				if err := recursiveTar(srcDir, srcFile.Join(newLocalPath(f.Name())),
					destDir, destFile.Join(newRemotePath(f.Name())), tw); err != nil {
					return err
				}
			}
			return nil
		} else if stat.Mode()&os.ModeSymlink != 0 {
			//case soft link
			hdr, _ := tar.FileInfoHeader(stat, fpath)
			target, err := os.Readlink(fpath)
			if err != nil {
				return err
			}

			hdr.Linkname = target
			hdr.Name = destFile.String()
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
		} else {
			//case regular file or other file type like pipe
			hdr, err := tar.FileInfoHeader(stat, fpath)
			if err != nil {
				return err
			}
			hdr.Name = destFile.String()

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			f, err := os.Open(fpath)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			return f.Close()
		}
	}
	return nil
}

func (o *CopyOptions) untarAll(ns, pod string, prefix string, src remotePath, dest localPath, reader io.Reader) error {
	symlinkWarningPrinted := false
	// TODO: use compression here?
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		// All the files will start with the prefix, which is the directory where
		// they were located on the pod, we need to strip down that prefix, but
		// if the prefix is missing it means the tar was tempered with.
		// For the case where prefix is empty we need to ensure that the path
		// is not absolute, which also indicates the tar file was tempered with.
		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		// basic file information
		mode := header.FileInfo().Mode()
		// header.Name is a name of the REMOTE file, so we need to create
		// a remotePath so that it goes through appropriate processing related
		// with cleaning remote paths
		destFileName := dest.Join(newRemotePath(header.Name[len(prefix):]))

		if !isRelative(dest, destFileName) {
			fmt.Fprintf(o.IOStreams.ErrOut, "warning: file %q is outside target destination, skipping\n", destFileName)
			continue
		}

		if err := os.MkdirAll(destFileName.Dir().String(), 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName.String(), 0755); err != nil {
				return err
			}
			continue
		}

		if mode&os.ModeSymlink != 0 {
			if !symlinkWarningPrinted && len(o.ExecParentCmdName) > 0 {
				fmt.Fprintf(o.IOStreams.ErrOut,
					"warning: skipping symlink: %q -> %q (consider using \"%s exec -n %q %q -- tar cf - %q | tar xf -\")\n",
					destFileName, header.Linkname, o.ExecParentCmdName, ns, pod, src)
				symlinkWarningPrinted = true
				continue
			}
			fmt.Fprintf(o.IOStreams.ErrOut, "warning: skipping symlink: %q -> %q\n", destFileName, header.Linkname)
			continue
		}
		outFile, err := os.Create(destFileName.String())
		if err != nil {
			return err
		}
		defer outFile.Close()
		if _, err := io.Copy(outFile, tarReader); err != nil {
			return err
		}
		if err := outFile.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (o *CopyOptions) execute(options *exec.ExecOptions) error {
	if len(options.Namespace) == 0 {
		options.Namespace = o.Namespace
	}

	if len(o.Container) > 0 {
		options.ContainerName = o.Container
	}

	options.Config = o.ClientConfig
	options.PodClient = o.Clientset.CoreV1()

	if err := options.Validate(); err != nil {
		return err
	}

	if err := options.Run(); err != nil {
		return err
	}
	return nil
}
