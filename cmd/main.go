package main

import (
	"flag"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/knative-sample/knload/cmd/app"
	"github.com/knative-sample/knload/cmd/app/signals"
	"github.com/knative-sample/knload/pkg/utils/logs"
	"github.com/golang/glog"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	stopCh := signals.SetupSignalHandler()

	// Start runner
	cmd := app.NewCommandStartServer(stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})

	if err := cmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}
