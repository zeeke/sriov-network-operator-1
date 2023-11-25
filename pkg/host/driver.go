package host

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"

	dputils "github.com/k8snetworkplumbingwg/sriov-network-device-plugin/pkg/utils"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/global/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/global/vars"
)

// Unbind unbind driver for one device
func (h *HostManager) Unbind(pciAddr string) error {
	log.Log.V(2).Info("Unbind(): unbind device driver for device", "device", pciAddr)
	yes, driver := h.HasDriver(pciAddr)
	if !yes {
		return nil
	}

	filePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDrivers, driver, "unbind")
	err := os.WriteFile(filePath, []byte(pciAddr), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "Unbind(): fail to unbind driver for device", "device", pciAddr)
		return err
	}
	return nil
}

// BindDpdkDriver bind dpdk driver for one device
// Bind the device given by "pciAddr" to the driver "driver"
func (h *HostManager) BindDpdkDriver(pciAddr, driver string) error {
	log.Log.V(2).Info("BindDpdkDriver(): bind device to driver",
		"device", pciAddr, "driver", driver)

	if yes, d := h.HasDriver(pciAddr); yes {
		if driver == d {
			log.Log.V(2).Info("BindDpdkDriver(): device already bound to driver",
				"device", pciAddr, "driver", driver)
			return nil
		}

		if err := h.Unbind(pciAddr); err != nil {
			return err
		}
	}

	driverOverridePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, "driver_override")
	err := os.WriteFile(driverOverridePath, []byte(driver), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "BindDpdkDriver(): fail to write driver_override for device",
			"device", pciAddr, "driver", driver)
		return err
	}
	bindPath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDrivers, driver, "bind")
	err = os.WriteFile(bindPath, []byte(pciAddr), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "BindDpdkDriver(): fail to bind driver for device",
			"driver", driver, "device", pciAddr)
		_, err := os.Readlink(filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, "iommu_group"))
		if err != nil {
			log.Log.Error(err, "Could not read IOMMU group for device", "device", pciAddr)
			return fmt.Errorf(
				"cannot bind driver %s to device %s, make sure IOMMU is enabled in BIOS. %w", driver, pciAddr, err)
		}
		return err
	}
	err = os.WriteFile(driverOverridePath, []byte(""), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "BindDpdkDriver(): failed to clear driver_override for device", "device", pciAddr)
		return err
	}

	return nil
}

// BindDefaultDriver bind driver for one device
// Bind the device given by "pciAddr" to the default driver
func (h *HostManager) BindDefaultDriver(pciAddr string) error {
	log.Log.V(2).Info("BindDefaultDriver(): bind device to default driver", "device", pciAddr)

	if yes, d := h.HasDriver(pciAddr); yes {
		if !sriovnetworkv1.StringInArray(d, vars.DpdkDrivers) {
			log.Log.V(2).Info("BindDefaultDriver(): device already bound to default driver",
				"device", pciAddr, "driver", d)
			return nil
		}
		if err := h.Unbind(pciAddr); err != nil {
			return err
		}
	}

	driverOverridePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, "driver_override")
	err := os.WriteFile(driverOverridePath, []byte("\x00"), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "BindDefaultDriver(): failed to write driver_override for device", "device", pciAddr)
		return err
	}

	pciDriversProbe := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDriversProbe)
	err = os.WriteFile(pciDriversProbe, []byte(pciAddr), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "BindDefaultDriver(): failed to bind driver for device", "device", pciAddr)
		return err
	}

	return nil
}

// Workaround function to handle a case where the vf default driver is stuck and not able to create the vf kernel interface.
// This function unbind the VF from the default driver and try to bind it again
// bugzilla: https://bugzilla.redhat.com/show_bug.cgi?id=2045087
func (h *HostManager) RebindVfToDefaultDriver(vfAddr string) error {
	log.Log.Info("RebindVfToDefaultDriver()", "vf", vfAddr)
	if err := h.Unbind(vfAddr); err != nil {
		return err
	}
	if err := h.BindDefaultDriver(vfAddr); err != nil {
		log.Log.Error(err, "RebindVfToDefaultDriver(): fail to bind default driver", "device", vfAddr)
		return err
	}

	log.Log.Info("RebindVfToDefaultDriver(): workaround implemented", "vf", vfAddr)
	return nil
}

func (h *HostManager) UnbindDriverIfNeeded(vfAddr string, isRdma bool) error {
	if isRdma {
		log.Log.Info("UnbindDriverIfNeeded(): unbinding driver", "device", vfAddr)
		if err := h.Unbind(vfAddr); err != nil {
			return err
		}
		log.Log.Info("UnbindDriverIfNeeded(): unbounded driver", "device", vfAddr)
	}
	return nil
}

func (h *HostManager) HasDriver(pciAddr string) (bool, string) {
	driver, err := dputils.GetDriverName(pciAddr)
	if err != nil {
		log.Log.V(2).Info("HasDriver(): device driver is empty for device", "device", pciAddr)
		return false, ""
	}
	log.Log.V(2).Info("HasDriver(): device driver for device", "device", pciAddr, "driver", driver)
	return true, driver
}
