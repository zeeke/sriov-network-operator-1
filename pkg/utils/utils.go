package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"syscall"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

//go:generate ../../bin/mockgen -destination mock/mock_utils.go -source utils.go
type CmdInterface interface {
	Chroot(string) (func() error, error)
	RunCommand(string, ...string) (string, string, error)
}

type utilsHelper struct {
}

func New() CmdInterface {
	return &utilsHelper{}
}

func (u *utilsHelper) Chroot(path string) (func() error, error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}
	vars.InChroot = true

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		vars.InChroot = false
		return syscall.Chroot(".")
	}, nil
}

// RunCommand runs a command
func (u *utilsHelper) RunCommand(command string, args ...string) (string, string, error) {
	log.Log.Info("RunCommand()", "command", command, "args", args)
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	log.Log.V(2).Info("RunCommand()", "output", stdout.String(), "error", err)
	return stdout.String(), stderr.String(), err
}

func GenerateRandomGUID() net.HardwareAddr {
	guid := make(net.HardwareAddr, 8)

	// First field is 0x01 - xfe to avoid all zero and all F invalid guids
	guid[0] = byte(1 + rand.Intn(0xfe))

	for i := 1; i < len(guid); i++ {
		guid[i] = byte(rand.Intn(0x100))
	}

	return guid
}

func HashConfigMap(cm *corev1.ConfigMap) string {
	var keys []string
	for k := range cm.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hash := fnv.New128()
	for _, k := range keys {
		hash.Write([]byte(k))
		hash.Write([]byte(cm.Data[k]))
	}
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

func IsCommandNotFound(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 127 {
			return true
		}
	}
	return false
}

func GetHostExtension() string {
	if vars.InChroot {
		return vars.FilesystemRoot
	}
	return filepath.Join(vars.FilesystemRoot, consts.Host)
}

func GetChrootExtension() string {
	if vars.InChroot {
		return vars.FilesystemRoot
	}
	return fmt.Sprintf("chroot %s%s", vars.FilesystemRoot, consts.Host)
}
