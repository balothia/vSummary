package poller

import (
	"context"
	"time"

	"github.com/gbolo/vsummary/common"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func (p *Poller) GetVirtualMachines() (vmList []common.VirtualMachine, err error) {

	// log time on debug
	defer common.ExecutionTime(time.Now(), "poll")

	// Create view of VirtualMachine objects
	m := view.NewManager(p.VmwareClient.Client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v, err := m.CreateContainerView(ctx, p.VmwareClient.Client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(
		ctx,
		[]string{"VirtualMachine"},
		[]string{"summary", "config", "guest", "runtime", "parent", "resourcePool", "parentVApp"},
		&vms,
	)
	if err != nil {
		return
	}

	// Print summary per vm (see also: govc/vm/info.go)
	for _, vm := range vms {

		// create vm struct
		vmStruct := common.VirtualMachine{
			Name:                 vm.Summary.Config.Name,
			Moref:                vm.Summary.Vm.Value,
			VmxPath:              vm.Config.Files.VmPathName,
			Vcpu:                 vm.Config.Hardware.NumCPU,
			MemoryMb:             vm.Config.Hardware.MemoryMB,
			ConfigGuestOs:        vm.Config.GuestId,
			ConfigVersion:        vm.Config.Version,
			SmbiosUuid:           vm.Config.Firmware,
			InstanceUuid:         vm.Config.Uuid,
			ConfigChangeVersion:  vm.Config.ChangeVersion,
			GuestToolsVersion:    vm.Guest.ToolsVersion,
			GuestToolsRunning:    vm.Guest.ToolsRunningStatus,
			GuestHostname:        vm.Guest.HostName,
			GuestIp:              vm.Guest.IpAddress,
			GuestOs:              vm.Guest.GuestId,
			StatCpuUsage:         vm.Summary.QuickStats.OverallCpuUsage,
			StatHostMemoryUsage:  vm.Summary.QuickStats.HostMemoryUsage,
			StatGuestMemoryUsage: vm.Summary.QuickStats.GuestMemoryUsage,
			StatUptimeSec:        vm.Summary.QuickStats.UptimeSeconds,
			PowerState:           string(vm.Runtime.PowerState),
			EsxiMoref:            vm.Runtime.Host.Value,
			Template:             vm.Config.Template,
			VcenterId:            v.Client().ServiceContent.About.InstanceUuid,
		}

		// folder may not exist
		if vm.Parent != nil {
			vmStruct.FolderMoref = vm.Parent.Value
			//vmStruct.FolderId = common.GetMD5Hash(fmt.Sprintf("%s%s", vmStruct.VcenterId, vmStruct.FolderMoref))
		}

		// vapps may not exist
		if vm.ParentVApp != nil {
			vmStruct.VappMoref = vm.ParentVApp.Value
			//vmStruct.VappId = common.GetMD5Hash(fmt.Sprintf("%s%s", vmStruct.VcenterId, vmStruct.VappMoref))
			vmStruct.FolderId = "vapp"
		} else {
			vmStruct.VappId = "none"
		}

		// resourcepool may not exist
		if vm.ResourcePool != nil {
			vmStruct.ResourcePoolMoref = vm.ResourcePool.Value
			//vmStruct.ResourcePoolId = common.GetMD5Hash(fmt.Sprintf("%s%s", vmStruct.VcenterId, vmStruct.ResourcePoolId))
		}

		// Fill in some required Ids
		//vmStruct.Id = common.GetMD5Hash(fmt.Sprintf("%s%s", vmStruct.VcenterId, vmStruct.Moref))
		//vmStruct.EsxiId = common.GetMD5Hash(fmt.Sprintf("%s%s", vmStruct.VcenterId, vmStruct.EsxiMoref))

		vmList = append(vmList, vmStruct)

	}

	log.Infof("poller fetched summary of %d vms", len(vmList))
	return

}