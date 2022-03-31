package node

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MachineConfigPool describes a pool of MachineConfigs.
type MachineConfigPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec MachineConfigPoolSpec `json:"spec"`
}

// MachineConfigPoolSpec is the spec for MachineConfigPool resource.
type MachineConfigPoolSpec struct {
	// machineConfigSelector specifies a label selector for MachineConfigs.
	// Refer https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/ on how label and selectors work.
	MachineConfigSelector *metav1.LabelSelector `json:"machineConfigSelector,omitempty"`

	// nodeSelector specifies a label selector for Machines
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`

	// paused specifies whether or not changes to this machine config pool should be stopped.
	// This includes generating new desiredMachineConfig and update of machines.
	Paused bool `json:"paused"`

	// maxUnavailable defines either an integer number or percentage
	// of nodes in the pool that can go Unavailable during an update.
	// This includes nodes Unavailable for any reason, including user
	// initiated cordons, failing nodes, etc. The default value is 1.
	//
	// A value larger than 1 will mean multiple nodes going unavailable during
	// the update, which may affect your workload stress on the remaining nodes.
	// You cannot set this value to 0 to stop updates (it will default back to 1);
	// to stop updates, use the 'paused' property instead. Drain will respect
	// Pod Disruption Budgets (PDBs) such as etcd quorum guards, even if
	// maxUnavailable is greater than one.
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

// KubeletConfig describes a customized Kubelet configuration.
type KubeletConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec KubeletConfigSpec `json:"spec"`
}

// KubeletConfigSpec defines the desired state of KubeletConfig
type KubeletConfigSpec struct {
	AutoSizingReserved        *bool                    `json:"autoSizingReserved,omitempty"`
	LogLevel                  *int32                   `json:"logLevel,omitempty"`
	MachineConfigPoolSelector *metav1.LabelSelector    `json:"machineConfigPoolSelector,omitempty"`
	KubeletConfig             *KubeletCPUConfiguration `json:"kubeletConfig,omitempty"`
}

// MachineConfig defines the configuration for a machine
type MachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MachineConfigSpec `json:"spec"`
}

// MachineConfigSpec is the spec for MachineConfig
type MachineConfigSpec struct {
	// OSImageURL specifies the remote location that will be used to
	// fetch the OS.
	OSImageURL string `json:"osImageURL"`
	// Config is a Ignition Config object.
	Config IgnitionConfig `json:"config"`

	// +nullable
	KernelArguments []string `json:"kernelArguments"`
	Extensions      []string `json:"extensions"`

	FIPS       bool   `json:"fips"`
	KernelType string `json:"kernelType"`
}

type IgnitionConfig struct {
	Ignition string `json:"ignition,omitempty"`
	Version  string `json:"version,omitempty"`
}

type KubeletCPUConfiguration struct {
	CPUManagerPolicy          string      `json:"cpuManagerPolicy"`
	CPUManagerReconcilePeriod metav1.Duration `json:"cpuManagerReconcilePeriod"`
	ReservedSystemCPUs        string      `json:"reservedSystemCPUs"`
}
