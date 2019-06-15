package knload

import (
	"context"
	"crypto/tls"

	"time"

	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"

	"os"
	"os/user"
	"path/filepath"

	"flag"

	"github.com/golang/glog"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type Knload struct {
	Debug            bool
	Namespace        string
	LabelSelector    string
	SavePath         string
	Clientset        *kubernetes.Clientset
	GatewayAddress   string
	ServiceUrl string
	Stages           []*Stage
	ResultChan       chan *Result
	GatewayIPAddress string
	GatewayPort      int
}

type Stage struct {
	Concurrent int
	Duration   int
}

type Result struct {
	Concurrent   int
	ResponseTime float64
	PodNum       float64
}

var (
	kubeconfig, masterURL string
	log                   = logf.KBLog.WithName("client").WithName("config")
)

func init() {
	// TODO: Fix this to allow double vendoring this library but still register flags on behalf of users
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")

	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}

// GetConfig creates a *rest.Config for talking to a Kubernetes apiserver.
// If --kubeconfig is set, will use the kubeconfig file at that location.  Otherwise will assume running
// in cluster and use the cluster provided kubeconfig.
//
// Config precedence
//
// * --kubeconfig flag pointing at a file
//
// * KUBECONFIG environment variable pointing at a file
//
// * In-cluster config if running in cluster
//
// * $HOME/.kube/config if exists
func (hl *Knload) getKubeconfig() (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags(masterURL, os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func (hl *Knload) Run() {
	var wg sync.WaitGroup
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	resultChans := make([]chan *Result, len(hl.Stages))
	for index, s := range hl.Stages {
		rc := make(chan *Result, s.Duration)
		resultChans[index] = rc
	}

	kubeconfig, _ := hl.getKubeconfig()
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		glog.Fatalf("get kube client set NewForConfig error: %s ", err.Error())
	}
	hl.Clientset = clientset

	// 初始化第一个点
	pods, err := hl.Clientset.CoreV1().Pods(hl.Namespace).List(metav1.ListOptions{
		LabelSelector: hl.LabelSelector,
		Limit:         1000,
	})
	if err != nil {
		glog.Errorf("get pod list error: %s ", err.Error())
	}
	podNum := 0
	for _, pod := range pods.Items {
		//if pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning {
		if pod.DeletionTimestamp == nil {
			podNum += 1
		}
	}
	result := &Result{
		PodNum: float64(podNum),
	}
	hl.ResultChan <- result

	// merge result in order
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for _, rc := range resultChans {
			for r := range rc {
				hl.ResultChan <- r
			}
		}
	}()

	// collect result and draw html
	wg1.Add(1)
	go func() {
		defer wg1.Done()

		hl.Draw()
	}()

	wg.Add(len(hl.Stages))
	for index, s := range hl.Stages {
		// 在 goroutine 中执行，如果到了时间还没完成也要启动下一批。现在的代码逻辑是等待上一批结束再启动下一批，这样可能就会不真实。
		go func(i int, s *Stage) {
			defer wg.Done()
			hl.run(s, resultChans[i])
		}(index, s)

		time.Sleep(time.Second * time.Duration(s.Duration))
	}

	podNums := []int{}
	for {
		pods, err := hl.Clientset.CoreV1().Pods(hl.Namespace).List(metav1.ListOptions{
			LabelSelector: hl.LabelSelector,
			Limit:         1000,
		})
		if err != nil {
			glog.Errorf("get pod list error: %s ", err.Error())
			continue
		}
		podNum := 0
		for _, pod := range pods.Items {
			//if pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning {
			if pod.DeletionTimestamp == nil {
				podNum += 1
			}
		}

		podNums = append(podNums, podNum)
		if podNum == 0 {
			break
		}

		time.Sleep(time.Second)
	}
	wg.Wait()
	wg2.Wait()
	for _, pm := range podNums {
		result := &Result{
			PodNum: float64(pm),
		}
		hl.ResultChan <- result
	}

	close(hl.ResultChan)
	wg1.Wait()

}

type DrawHtml struct {
	ConcurrentStr   string
	ResponseTimeStr string
	PodNumStr       string
	JqueryJS        interface{}
	D3JS            interface{}
	C3JS            interface{}
	C3CSS           interface{}
}

func (hl *Knload) run(s *Stage, resultChan chan *Result) {
	var wg1 sync.WaitGroup
	wg1.Add(s.Duration + 1)
	rm := map[int]*Result{}
	var mutex sync.Mutex

	_run := func(index int) {
		var wg sync.WaitGroup
		wg.Add(s.Concurrent)
		_rc := make(chan float64, s.Concurrent)

		pods, err := hl.Clientset.CoreV1().Pods(hl.Namespace).List(metav1.ListOptions{
			LabelSelector: hl.LabelSelector,
			Limit:         1000,
		})
		if err != nil {
			glog.Errorf("get pod list error: %s ", err.Error())
		}
		podNum := 0
		for _, pod := range pods.Items {
			if pod.DeletionTimestamp == nil {
				podNum += 1
			}
		}

		for i := 0; i < s.Concurrent; i++ {
			go func(stage *Stage) {
				defer wg.Done()
				hl.doRequest(_rc)
			}(s)

			// 尽量打散
			if i%5 == 0 {
				time.Sleep(time.Millisecond * 50)
			}
		}

		wg.Wait()
		close(_rc)

		// 对当前秒求平均值
		var responseTimeTotal float64
		for rp := range _rc {
			responseTimeTotal += rp
		}

		mutex.Lock()
		defer mutex.Unlock()
		rm[index] = &Result{
			Concurrent:   s.Concurrent,
			PodNum:       float64(podNum),
			ResponseTime: responseTimeTotal / float64(s.Concurrent),
		}
	}

	go func() {
		defer wg1.Done()
		// Result 进行排序，并且尽早返回 result
		for ii := 0; ii < s.Duration; ii++ {
			for {
				mutex.Lock()
				if r, ok := rm[ii]; ok {
					resultChan <- r
					mutex.Unlock()
					break
				} else {
					mutex.Unlock()
					time.Sleep(time.Millisecond * 200)
				}
			}
		}
	}()

	for ii := 0; ii < s.Duration; ii++ {
		go func(i int) {
			defer wg1.Done()
			_run(i)
		}(ii)
		time.Sleep(time.Second)
	}

	wg1.Wait()
	close(resultChan)
}

func (hl *Knload) doRequest(resultChan chan float64) {
	// do request
	responseTime, err := hl.getResponseTime()
	if err != nil {
		// 强制重试一次
		responseTime, _ = hl.getResponseTime()
	}

	resultChan <- responseTime
}

func (hl *Knload) getResponseTime() (responseTime float64, err error) {
	u, err := url.Parse(fmt.Sprintf("http://%s", hl.ServiceUrl))
	if err != nil {
		glog.Error(err)
		return 0, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		glog.Errorf("unable to create request: %v", err)
		return 0, err
	}

	var t0, t1, t2, t3, t4, t5, t6, t7 time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// connecting to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				glog.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()

			glog.V(5).Infof("\nConnected to %s\n", addr)
		},

		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}

	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
	tr := &http.Transport{
		//Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	tr.DialContext = hl.dialContext("tcp4")

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that
			// manually if required.
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("failed to read response: %v", err)
		return 0, err
	}

	//bodyMsg := readResponseBody(req, resp)
	defer resp.Body.Close()
	hl.readResponseBody(resp)

	t7 = time.Now() // after read body
	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}

	hl.logResponseTime("tltal time", t7, t0)
	hl.logResponseTime("dns lookup", t1, t0)
	hl.logResponseTime("tcp connection", t2, t1)
	hl.logResponseTime("tls handshake", t6, t5)
	hl.logResponseTime("server processing", t4, t3)
	hl.logResponseTime("content transfer", t7, t4)
	hl.logResponseTime("namelookup", t1, t0)
	hl.logResponseTime("connect", t2, t0)
	hl.logResponseTime("pretransfer", t3, t0)
	hl.logResponseTime("starttransfer:", t4, t0)
	hl.logResponseTime("FirstResponse: ", t4, t1)
	return float64(t4.Sub(t1)) / float64(time.Second), nil
}

func (hl *Knload) logResponseTime(msg string, t2, t1 time.Time) {
	glog.V(5).Infof("%s %s\n", msg, t2.Sub(t1))
}

func (hl *Knload) dialContext(network string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: false,
		}).DialContext(ctx, network, hl.GatewayAddress)
	}
}

func (hl *Knload) readResponseBody(resp *http.Response) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Fatal(err)
	}
	bodyString := string(bodyBytes)
	glog.V(5).Infof("responseBody: %s", bodyString)
	return bodyString
}

func (s *Stage) getPodNum() (podNum int, err error) {
	return

}
