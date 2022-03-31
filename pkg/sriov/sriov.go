package sriov

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSRIOVNetworkNodePolicy(name string, pfName string, resourceName string, node *v1.Node) (*SriovNetworkNodePolicy, error) {
	var available resource.Quantity
	var exists bool
	if available, exists = node.Status.Allocatable[v1.ResourceName(resourceName)]; !exists {
		return nil, fmt.Errorf("node %s has no resource %s", node.Name, resourceName)
	}

	num, ok := available.AsInt64()
	if !ok {
		return nil, fmt.Errorf("failed to convert resource availability for resource %s", resourceName)
	}

	return &SriovNetworkNodePolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sriovnetwork.openshift.io/v1",
			Kind:       "SriovNetworkNodePolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
			Name:      name,
			Namespace: "sriov-network-operator",
		},
		Spec: SriovNetworkNodePolicySpec{
			DeviceType: "vfio-pci",
			Mtu:        9000,
			NicSelector: SriovNetworkNicSelector{
				PfNames: []string{
					pfName,
				},
			},
			NumVfs:       int(num),
			Priority:     90,
			ResourceName: resourceName,
			NodeSelector: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
		},
	}, nil
}

func NewSRIOVNetwork(name string, resourceName string, node *v1.Node) (*SriovNetwork, error) {
	return &SriovNetwork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sriovnetwork.openshift.io/v1",
			Kind:       "SriovNetwork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"node-role.kubernetes.io/sap": "",
			},
			Name:      name,
			Namespace: "sriov-network-operator",
		},
		Spec: SriovNetworkSpec{
			IPAM:             "{}",
			NetworkNamespace: "default",
			ResourceName:     resourceName,
			SpoofChk:         "off",
		},
	}, nil
}
