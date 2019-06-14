package options

import (
	"github.com/spf13/cobra"
)

type Options struct {
	Namespace      string
	LabelSelector  string
	GatewayAddress string
	ServiceUrl string
	SavePath       string
	Stages         string
}

func (s *Options) SetOps(ac *cobra.Command) {
	ac.Flags().StringVar(&s.Namespace, "namespace", s.Namespace, "kubernetes namespace")
	ac.Flags().StringVar(&s.LabelSelector, "label", s.LabelSelector, "Pod label")
	ac.Flags().StringVar(&s.SavePath, "save-path", s.GatewayAddress, "save path")
	ac.Flags().StringVar(&s.GatewayAddress, "gateway-address", s.SavePath, "gateway address ")
	ac.Flags().StringVar(&s.ServiceUrl, "service-url", s.ServiceUrl, "service url")
	ac.Flags().StringVar(&s.Stages, "stages", "5:120,20:120,40:120,80:120", "test stages, default is 5:120,20:120,40:120,80:120")
}
