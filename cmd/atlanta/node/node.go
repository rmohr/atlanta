package node

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/ghodss/yaml"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
	"github.com/rmohr/atlanta/pkg/api"
	node2 "github.com/rmohr/atlanta/pkg/node"
	"github.com/rmohr/atlanta/pkg/sriov"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	nodeExample = `TODO`
)

var Scheme = runtime.NewScheme()

// Codecs provides access to encoding and decoding for the scheme
var Codecs = serializer.NewCodecFactory(Scheme)

// ParameterCodec handles versioning of objects that are converted to query parameters.
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	var localSchemeBuilder = runtime.SchemeBuilder{
		v1.AddToScheme,
	}
	localSchemeBuilder.AddToScheme(Scheme)
}

// NodeOptions provides information required to update
// the current context on a user's KUBECONFIG
type NodeOptions struct {
	configFlags  *genericclioptions.ConfigFlags
	sriovDevices []string

	restConfig *rest.Config
	args       []string
	node       string

	genericclioptions.IOStreams
	client *kubernetes.Clientset
}

// NewNodeOptions provides an instance of NodeOptions with default values
func NewNodeOptions(streams genericclioptions.IOStreams) *NodeOptions {
	return &NodeOptions{
		configFlags: genericclioptions.NewConfigFlags(true),

		IOStreams: streams,
	}
}

func NewCmdNode(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewNodeOptions(streams)

	cmd := &cobra.Command{
		Use:          "node",
		Short:        "render configuration objects for specific nodes",
		Example:      fmt.Sprintf(nodeExample, "atlanta"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	o.configFlags.AddFlags(cmd.Flags())
	cmd.Flags().StringArrayVar(&o.sriovDevices, "sriov", []string{}, "name of the sriov device resource name on the node")
	return cmd
}

// Complete sets all information required for updating the current context
func (o *NodeOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args
	var err error
	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	o.client, err = kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *NodeOptions) Validate() error {
	if len(o.args) == 0 {
		return fmt.Errorf("please provide a node name")
	}
	if len(o.args) > 1 {
		return fmt.Errorf("please provide only one node name")
	}

	o.node = o.args[0]

	return nil
}

// Run lists all available namespaces on a user's KUBECONFIG or updates the
// current context based on a provided namespace.
func (o *NodeOptions) Run() error {

	ctx := context.Background()

	node, err := o.client.CoreV1().Nodes().Get(ctx, o.node, metav1.GetOptions{})
	if err != nil {
		return err
	}

	caps, err := o.copyCapabilities(*o.configFlags.Namespace)
	if err != nil {
		return err
	}

	_ = api.NodeInfo{
		ObjectMeta: metav1.ObjectMeta{Name: o.node},
		Discovered: api.DiscoveredNodeInfo{
			Resources:         node.Status.Allocatable,
			NumaTopology:      caps.Host.NUMA.Cells,
			VMEnabled:         node.Labels["kubevirt.io/schedulable"] == "true",
			CPUManagerEnabled: node.Labels["cpumanager"] == "true",
		},
	}
	toMarshal := []interface{}{
		node2.NewMachineConfigPool(node.Name),
		node2.NewMachineConfig(node.Name),
		node2.NewKubeletConfig(node.Name, caps),
	}

	for i, resourceName := range o.sriovDevices {
		networkPolicy, err := sriov.NewSRIOVNetworkNodePolicy(fmt.Sprintf("%v-%v", node.Name, i), fmt.Sprintf("pf%v", i), resourceName, node)
		if err != nil {
			return err
		}
		network, err := sriov.NewSRIOVNetwork(fmt.Sprintf("%v-%v", node.Name, i), resourceName, node)
		if err != nil {
			return err
		}
		toMarshal = append(toMarshal, networkPolicy, network)
	}

	info, err := yaml.Marshal(toMarshal)
	fmt.Println(string(info))

	return nil
}

func (o *NodeOptions) copyCapabilities(namespace string) (*libvirtxml.Caps, error) {

	handler, err := o.getVirtHandler(o.node, namespace)
	if err != nil {
		return nil, err
	}

	req := o.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(handler.Name).
		Namespace(handler.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: "virt-handler",
		Command:   []string{"/usr/bin/cat", "/var/lib/kubevirt-node-labeller/capabilities.xml"},
		//Command: []string{"/usr/bin/echo", "hi"},
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		TTY:    false,
	}, ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(o.restConfig, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	out, in := io.Pipe()
	errChan := make(chan error, 1)
	go func() {
		defer in.Close()
		errChan <- exec.Stream(remotecommand.StreamOptions{
			Stdout: in,
			Stderr: in,
		})
	}()

	capabilities, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}
	caps := &libvirtxml.Caps{}
	err = xml.Unmarshal(capabilities, caps)
	if err != nil {
		return nil, err
	}

	err = <-errChan
	return caps, nil
}

func (o *NodeOptions) getVirtHandler(nodeName string, namespace string) (*v1.Pod, error) {

	handlerNodeSelector := fields.ParseSelectorOrDie("spec.nodeName=" + nodeName)
	labelSelector, err := labels.Parse("kubevirt.io in (virt-handler)")
	if err != nil {
		return nil, err
	}

	pods, err := o.client.CoreV1().Pods(namespace).List(context.Background(),
		metav1.ListOptions{
			FieldSelector: handlerNodeSelector.String(),
			LabelSelector: labelSelector.String()})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("Expected to find one Pod, found %d Pods", len(pods.Items))
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pod found")
	}
	return &pods.Items[0], nil
}
