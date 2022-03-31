package fswrap

import (
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

// Following variables are intended to be modified during tests.

type MkdirAllFn func(path string, perm os.FileMode) error

var MkdirAll MkdirAllFn = os.MkdirAll

type WriteFileFn func(filename string, data []byte, perm fs.FileMode) error

var WriteFile WriteFileFn = ioutil.WriteFile

type ReadFileFn func(filename string) ([]byte, error)

var ReadFile ReadFileFn = ioutil.ReadFile

type CommandFn func(name string, arg ...string) *exec.Cmd

var Command CommandFn = exec.Command

type ChrootFn func(path string) (func() error, error)

var Chroot ChrootFn = chroot

func chroot(path string) (func() error, error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		return syscall.Chroot(".")
	}, nil
}
