package node

import (
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
	"strings"
	"time"
)

func NewMachineConfigPool(name string) *MachineConfigPool {
	maxUnavailable := intstr.FromInt(1)
	return &MachineConfigPool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineConfigPool",
			APIVersion: "machineconfiguration.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
			Name: name,
		},
		Spec: MachineConfigPoolSpec{
			Paused:         false,
			MaxUnavailable: &maxUnavailable,
			MachineConfigSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"node-role.kubernetes.io/sap": "",
				},
			},
			NodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"node-role.kubernetes.io/sap": "",
				},
			},
		},
	}
}

func NewMachineConfig(name string) *MachineConfig {
	return &MachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineConfig",
			APIVersion: "machineconfiguration.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
			Name: name,
		},
		Spec: MachineConfigSpec{
			Config: IgnitionConfig{
				Version: "3.1.0",
			},
			KernelArguments: []string{"intel_iommu=on iommu=pt kvm.nx_huge_pages=off default_hugepagesz=1G hugepagesz=1G hugepages=2800"},
		},
	}
}

func NewKubeletConfig(name string, caps *libvirtxml.Caps) *KubeletConfig {

	lastCell := caps.Host.NUMA.Cells.Cells[len(caps.Host.NUMA.Cells.Cells)-1]
	coreIDs := []string{}
	for _, cpu := range lastCell.CPUS.CPUs {
		coreIDs = append(coreIDs, strconv.Itoa(cpu.ID))
	}

	return &KubeletConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeletConfig",
			APIVersion: "machineconfiguration.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
			Name: name,
		},
		Spec: KubeletConfigSpec{
			MachineConfigPoolSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"node-role.kubernetes.io/sap": "",
				},
			},
			KubeletConfig: &KubeletCPUConfiguration{
				CPUManagerPolicy:          "static",
				CPUManagerReconcilePeriod: metav1.Duration{Duration: 5 * time.Second},
				ReservedSystemCPUs:        strings.Join(coreIDs, ", "),
			},
		},
	}
}
