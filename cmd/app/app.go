package app

import (
	"strings"

	"github.com/knative-sample/knload/cmd/app/options"
	"github.com/spf13/cobra"

	"strconv"

	"os"

	"github.com/knative-sample/knload/pkg/knload"
	"github.com/golang/glog"
)

// start edas api
func NewCommandStartServer(stopCh <-chan struct{}) *cobra.Command {
	ops := &options.Options{}
	mainCmd := &cobra.Command{
		Short: "Golang Sandbox ",
		Long:  "Alibaba Cloud Container Service Sandbox for Golang ",
		RunE: func(c *cobra.Command, args []string) error {
			glog.V(2).Infof("NewCommandStartServer main:%s", strings.Join(args, " "))
			run(stopCh, ops)
			return nil
		},
	}

	ops.SetOps(mainCmd)
	return mainCmd
}

func run(stopCh <-chan struct{}, ops *options.Options) {
	if ops.SavePath == "" {
		glog.Fatal("--save-path is empty")
	}

	if ops.Namespace == "" {
		glog.Fatal("--namespace is empty")
	}

	if ops.LabelSelector == "" {
		glog.Fatal("--label is empty")
	}

	if ops.ServiceUrl== "" {
		glog.Fatal("--service-url is empty")
	}

	if ops.GatewayAddress== "" {
		glog.Fatal("--gateway-address is empty")
	}

	ss := strings.Split(ops.Stages, ",")
	var stages []*knload.Stage
	for _, stageStr := range ss {
		stage := &knload.Stage{}
		_ss := strings.Split(stageStr, ":")
		if len(_ss) != 2 {
			continue
		}

		if cc, err := strconv.Atoi(_ss[0]); err != nil {
			continue
		} else {
			if cc == 0 {
				continue
			}
			stage.Concurrent = cc
		}

		if du, err := strconv.Atoi(_ss[1]); err != nil {
			continue
		} else {
			if du == 0 {
				continue
			}
			stage.Duration = du
		}
		stages = append(stages, stage)
	}

	kl := &knload.Knload{
		Namespace:      ops.Namespace,
		LabelSelector:  ops.LabelSelector,
		GatewayAddress: ops.GatewayAddress,
		SavePath:       ops.SavePath,
		ServiceUrl:  ops.ServiceUrl,
		Stages:         stages,
		ResultChan:     make(chan *knload.Result, 1000),
	}
	kl.Run()
	os.Exit(0)

	<-stopCh
}
