package whereami

import (
	"fmt"
	"io"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/clusterinfo"
)

type WhereamiOptions struct {
	clusterinfo.ClusterInfoOptions
}

// NewWhereamiOptions creates the options for showing whereami
func NewWhereamiOptions(ioStreams genericclioptions.IOStreams) *WhereamiOptions {
	return &WhereamiOptions{
		IOStreams: genericclioptions.NewIOStreams(),

	}
}

func NewWhereamiShow(w *WhereamiOptions, kubePath string) error {
	// Print environment
	w.
	// Print cluster info
	// Print list of namespaces
	return nil
}

func printService(out io.Writer, name, link string) {
	ct.ChangeColor(ct.Green, false, ct.None, false)
	fmt.Fprint(out, name)
	ct.ResetColor()
	fmt.Fprint(out, " is running at ")
	ct.ChangeColor(ct.Yellow, false, ct.None, false)
	fmt.Fprint(out, link)
	ct.ResetColor()
	fmt.Fprintln(out, "")
}
