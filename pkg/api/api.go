package api

import (
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeInfo struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Discovered DiscoveredNodeInfo `json:"discovered"`
}

type DiscoveredNodeInfo struct {
	Resources v1.ResourceList `json:"resources,omitempty"`

	NumaTopology *libvirtxml.CapsHostNUMACells `json:"numaTopology,omitempty"`

	VMEnabled bool `json:"vmEnabled"`

	CPUManagerEnabled bool `json:"cpuManagerEnabled"`
}