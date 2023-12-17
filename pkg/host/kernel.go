package host

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

// GetCurrentKernelArgs This retrieves the kernel cmd line arguments
func (h *HostManager) GetCurrentKernelArgs() (string, error) {
	path := consts.ProcKernelCmdLine
	if !vars.UsingSystemdMode {
		path = filepath.Join("/host", path)
	}

	path = filepath.Join(vars.FilesystemRoot, path)
	cmdLine, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("GetCurrentKernelArgs(): Error reading %s: %v", path, err)
	}
	return string(cmdLine), nil
}

// IsKernelArgsSet This checks if the kernel cmd line is set properly. Please note that the same key could be repeated
// several times in the kernel cmd line. We can only ensure that the kernel cmd line has the key/val kernel arg that we set.
func (h *HostManager) IsKernelArgsSet(cmdLine string, karg string) bool {
	elements := strings.Fields(cmdLine)
	for _, element := range elements {
		if element == karg {
			return true
		}
	}
	return false
}
