package k8sclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	goretry "github.com/avast/retry-go/v4"
	"github.com/go-resty/resty/v2"
	req "github.com/imroc/req/v3"
	jsoniter "github.com/json-iterator/go"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	thisNode *v1.Node
	once     sync.Once
)

type Metadata struct {
	Annotations map[string]json.RawMessage `json:"annotations"`
}

type update struct {
	Metadata Metadata `json:"metadata"`
}

// GetNodeLabels returns node labels.
// NODE_NAME environment variable is used to determine the node
func GetNodeLabels() (map[string]string, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return nil, nil
	}
	nodes, err := cSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.ObjectMeta.Labels, nil
}

// SetPodAnnotation adds or modifies annotation for pod
func SetPodAnnotation(pod *v1.Pod, key string, value string) error {
	cSet, err := createClientSet()
	if err != nil {
		return err
	}
	//merge := update{}
	//merge.Metadata.Annotations = make(map[string]json.RawMessage)
	//merge.Metadata.Annotations[key] = json.RawMessage(`"` + value + `"`)
	//
	//jsonData, err := json.Marshal(merge)
	//if err != nil {
	//	return err
	//}
	//_, err = cSet.CoreV1().Pods(pod.ObjectMeta.Namespace).Patch(context.TODO(), pod.ObjectMeta.Name, types.MergePatchType, jsonData, metav1.PatchOptions{})

	// 资源冲突时重试
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if _, ok := pod.Annotations[key]; !ok || pod.Annotations[key] != value {
			newPod, err := cSet.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			newPod.Annotations[key] = value
			_, err = cSet.CoreV1().Pods(newPod.Namespace).Update(context.TODO(), newPod, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if retryErr != nil {
		log.Println(retryErr.Error())
		return retryErr
	}
	return nil
}

// RefreshPod takes an existing Pod object as an input, and re-reads it from the K8s API
// Returns the refreshed Pod descriptor in case of success, or an error
func RefreshPod(pod v1.Pod) (*v1.Pod, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	return cSet.CoreV1().Pods(pod.ObjectMeta.Namespace).Get(context.TODO(), pod.ObjectMeta.Name, metav1.GetOptions{})
}

func GetRunningContainer(ctx context.Context, pod *v1.Pod, containerName string) (*v1.ContainerStatus, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}

	pod, err = RefreshPod(*pod)
	if err != nil {
		return nil, err
	}

	fmt.Println("Watch Kubernetes Pods in CrashLoopBackOff state")
	watcher, err := cSet.CoreV1().Pods(pod.Namespace).Watch(context.Background(),
		metav1.ListOptions{
			ResourceVersion: pod.ResourceVersion,
			FieldSelector:   fields.Set{"metadata.name": pod.Name}.AsSelector().String(),
			LabelSelector:   labels.Everything().String(),
		})
	if err != nil {
		fmt.Printf("error create pod watcher: %v\n", err)
		return nil, err
	}

	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				continue
			}
			for _, c := range pod.Status.ContainerStatuses {
				if !c.Ready {
					if c.State.Waiting != nil {
						fmt.Printf("PodName: %s, Namespace: %s, Phase: %s WaitingReason: %s \n", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace, pod.Status.Phase, c.State.Waiting.Reason)
					}
				}
				if c.Name == containerName && c.Ready {
					return &c, nil
				}

			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}
	return nil, err
}

func GetPod(namespace, podName string) (*v1.Pod, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	return cSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
}

func GetPVC(namespace, pvcName string) (*v1.PersistentVolumeClaim, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	return cSet.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
}

func GetPV(pvName string) (*v1.PersistentVolume, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	return cSet.CoreV1().PersistentVolumes().Get(context.TODO(), pvName, metav1.GetOptions{})
}

func PatchPVReclaimPolicy(pvName string, policy v1.PersistentVolumeReclaimPolicy) (*v1.PersistentVolume, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}

	data := fmt.Sprintf(`{"spec":{"persistentVolumeReclaimPolicy":"%s"}}`, policy)
	return cSet.CoreV1().PersistentVolumes().Patch(context.TODO(), pvName, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
}

func PatchPVClearClaimRef(pvName string) (*v1.PersistentVolume, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}

	data := fmt.Sprintf(`{"spec":{"claimRef": null}}`)
	return cSet.CoreV1().PersistentVolumes().Patch(context.TODO(), pvName, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
}

// GetMyPods returns all the Pods to the caller running on the same node as this process
// Node is identified by the NODE_NAME environment variable. The Pods are filtered based on their spec.nodeName attribute
func GetMyPods() (*v1.PodList, error) {
	cSet, err := createClientSet()
	if err != nil {
		return nil, err
	}
	nodeName := os.Getenv("NODE_NAME")
	// kubectl get pod  --field-selector spec.nodeName=${NODE_NAME} -A -o wide
	return cSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
}

func GetThisNode() *v1.Node {
	once.Do(func() {
		var err error
		nodeName := os.Getenv("NODE_NAME")
		if nodeName == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return
			}
			nodeName = hostname
		}

		cSet, err := createClientSet()
		if err != nil {
			return
		}

		thisNode, err = cSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return
		}
	})
	log.Println("this node is ", thisNode.Name)
	return thisNode
}

func ParseIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}

// GetPodsFromKubelet returns all the Pods to the caller running on the same node as this process
func GetPodsFromKubelet() (*v1.PodList, error) {
	config, err := createRestConfig()
	if err != nil {
		return nil, err
	}

	node := GetThisNode()
	if err != nil {
		return nil, err
	}

	// 获取 kubelet IP
	var ipAddress string
	for _, ipInfo := range node.Status.Addresses {
		if ipInfo.Type == v1.NodeInternalIP {
			ipAddress = ipInfo.Address
			break
		}
	}
	log.Println("nodeIP", ipAddress)

	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}

	var client *resty.Client
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	tlsConfig.InsecureSkipVerify = true
	client = resty.New().SetTLSClientConfig(tlsConfig)

	if transportConfig.HasTokenAuth() {
		client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.BearerToken))
	}
	podUrl := fmt.Sprintf("https://%s:%d/pods", ipAddress, node.Status.DaemonEndpoints.KubeletEndpoint.Port)

	_, iptype := ParseIP(ipAddress)
	if iptype == 6 {
		podUrl = fmt.Sprintf("https://[%s]:%d/pods", ipAddress, node.Status.DaemonEndpoints.KubeletEndpoint.Port)
	}
	log.Println("nodeUrl:", podUrl)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	resp, err := client.R().SetContext(ctx).SetDebug(false).Get(podUrl)
	if err != nil {
		return nil, err
	}
	var podList = &v1.PodList{}
	if err := jsoniter.Unmarshal(resp.Body(), podList); err != nil {
		return nil, err
	}
	return podList, nil
}

// GetPodsFromKubelet returns all the Pods to the caller running on the same node as this process
func GetMetricsFromKubelet(cadvisor bool) (map[string]*dto.MetricFamily, error) {
	config, err := createRestConfig()
	if err != nil {
		return nil, err
	}

	node := GetThisNode()
	// 获取 kubelet IP
	var ipAddress string
	for _, ipInfo := range node.Status.Addresses {
		if ipInfo.Type == v1.NodeInternalIP {
			ipAddress = ipInfo.Address
		}
	}

	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}

	var client *resty.Client
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	tlsConfig.InsecureSkipVerify = true
	client = resty.New().SetTLSClientConfig(tlsConfig)

	if transportConfig.HasTokenAuth() {
		client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.BearerToken))
	}

	metricsUri := fmt.Sprintf("https://%s:%d/metrics", ipAddress, node.Status.DaemonEndpoints.KubeletEndpoint.Port)
	if cadvisor {
		metricsUri, _ = url.JoinPath(metricsUri, "cadvisor")
	}

	resp, err := client.R().Get(metricsUri)
	if err != nil {
		return nil, err
	}
	//dec := expfmt.NewDecoder(resp.RawBody(), expfmt.Format(resp.Header().Get("Content-Type")))
	//mf := &dto.MetricFamily{}
	//dec.Decode(mf)
	//fmt.Println(resp)
	var parser expfmt.TextParser
	return parser.TextToMetricFamilies(bytes.NewReader(resp.Body()))
}

func CreatePvc(ctx context.Context, namespace string, req *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	//req := &v1.PersistentVolumeClaim{}
	cli, err := createClientSet()
	if err != nil {
		return nil, err
	}

	pvc, err := cli.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, req, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	if err := waitPvcBound(ctx, namespace, pvc.Name); err != nil {
		return nil, err
	}
	return pvc, nil
}

func DeletePvc(ctx context.Context, namespace, pvcName string) error {
	k8sCli, err := createClientSet()
	if err != nil {
		return err
	}
	if err := k8sCli.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return waitPvcDeleted(ctx, namespace, pvcName)
}

func DeletePod(ctx context.Context, namespace, podName string) error {
	k8sCli, err := createClientSet()
	if err != nil {
		return err
	}
	if err := k8sCli.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return waitPodDeleted(ctx, namespace, podName)
}

func CreatePod(ctx context.Context, namespace string, req *v1.Pod) (*v1.Pod, error) {
	var (
		resp = &v1.Pod{}
		//w    watch.Interface
	)

	cli, err := createClientSet()
	if err != nil {
		return nil, err
	}

	if resp, err = cli.CoreV1().Pods(namespace).Create(ctx, req, metav1.CreateOptions{}); err != nil {
		return nil, err
	}

	fmt.Printf("Pod created: %s", resp)
	if err := waitPodRunning(ctx, namespace, resp.Name); err != nil {
		return nil, err
	}

	return resp, nil
}

func SetPodImage(namespace, podName, containerToUpdate, image string) (*v1.Pod, error) {
	clientset, err := createClientSet()
	if err != nil {
		return nil, err
	}

	// 获取 Pod
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	// 找到你想要更新的容器，并更新其镜像
	for i, container := range pod.Spec.Containers {
		if container.Name == containerToUpdate {
			pod.Spec.Containers[i].Image = image
			break
		}
	}

	// 更新 Pod
	pod, err = clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return pod, err
}

func ClientSet() (*kubernetes.Clientset, error) {
	return createClientSet()
}

func RestConfig() (*rest.Config, error) {
	return createRestConfig()
}

// replace client-go's Transport with *req.Transport
func replaceTransport(config *rest.Config, t *req.Transport) {
	// Extract tls.Config from rest.Config
	tlsConfig, err := rest.TLSConfigFor(config)
	if err != nil {
		panic(err.Error())
	}
	// Set TLSClientConfig to req's Transport.
	t.TLSClientConfig = tlsConfig
	// Override with req's Transport.
	config.Transport = t
	// rest.Config.TLSClientConfig 必须为空，自定义的 Transport 不允许传证书，在 retryDialtls 中重建 tls.Config
	config.TLSClientConfig = rest.TLSClientConfig{}
}

func createClientSet() (*kubernetes.Clientset, error) {
	config, err := createRestConfig()
	if err != nil {
		return nil, err
	}

	reqClient := req.NewClient()
	reqClient.EnableDump(&req.DumpOptions{
		RequestHeader:  true,
		ResponseHeader: true,
	})
	reqClient.EnableDebugLog()

	reqClient.Transport.SetDialTLS(retryDialtls(config.TLSClientConfig))
	reqClient.Transport.SetDial(retryDial)
	replaceTransport(config, reqClient.Transport)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// RetriableError is a custom error that contains a positive duration for the next retry
type RetriableError struct {
	Err        error
	RetryAfter time.Duration
}

// Error returns error message and a Retry-After duration
func (e *RetriableError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Err.Error(), e.RetryAfter)
}

var _ error = (*RetriableError)(nil)

// retryDialtls tls 连接重试
func retryDialtls(cfg rest.TLSClientConfig) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		var err error
		var conn net.Conn
		dialer := &net.Dialer{
			Timeout: 10 * time.Second,
		}

		// 从 rest.TLSClientConfig 重建 tls 证书配置，用于建立 tls 连接
		certPool, err := x509.SystemCertPool()
		if cfg.CAData != nil {
			certPool.AppendCertsFromPEM(cfg.CAData)
		}
		if cfg.CAFile != "" {
			if caCertPEM, err := os.ReadFile(cfg.CAFile); err != nil {
				log.Println("cafile error")
			} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
				log.Println("invalid cert in CA PEM")
			}
		}

		var cert tls.Certificate
		if cfg.CertData != nil {
			cert, err = tls.X509KeyPair(cfg.CertData, cfg.KeyData)
			if err != nil {
				log.Println("cert error")
			}
		}

		if cfg.CertFile != "" {
			cert, err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
			if err != nil {
				log.Println("cert error")
			}
		}

		err = goretry.Do(
			func() error {
				conn, err = tls.DialWithDialer(dialer, network, addr, &tls.Config{
					RootCAs:      certPool,
					Certificates: []tls.Certificate{cert},
				})
				if err != nil {
					return &RetriableError{
						Err:        err,
						RetryAfter: time.Duration(3) * time.Second,
					}
				}
				return nil
			},
			goretry.DelayType(func(n uint, err error, config *goretry.Config) time.Duration {
				fmt.Println("Server fails with: "+err.Error(), n)
				if retriable, ok := err.(*RetriableError); ok {
					fmt.Printf("Client follows server recommendation to retry after %v\n", retriable.RetryAfter)
					return retriable.RetryAfter
				}
				// apply a default exponential back off strategy
				return goretry.BackOffDelay(n, err, config)
			}),
			goretry.Attempts(3), // 执行一次重试两次，共执行3次
		)
		return conn, err
	}
}

// retryDial tcp 连接重试
func retryDial(ctx context.Context, network, addr string) (net.Conn, error) {
	var err error
	var conn net.Conn
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	err = goretry.Do(
		func() error {
			conn, err = dialer.DialContext(ctx, network, addr)
			if err != nil {
				return &RetriableError{
					Err:        err,
					RetryAfter: time.Duration(3) * time.Second,
				}
			}
			return nil
		},
		goretry.DelayType(func(n uint, err error, config *goretry.Config) time.Duration {
			fmt.Println("Server fails with: "+err.Error(), n)
			if retriable, ok := err.(*RetriableError); ok {
				fmt.Printf("Client follows server recommendation to retry after %v\n", retriable.RetryAfter)
				return retriable.RetryAfter
			}
			// apply a default exponential back off strategy
			return goretry.BackOffDelay(n, err, config)
		}),
		goretry.Attempts(3), // 执行一次重试两次，共执行3次
	)
	return conn, err
}

func createRestConfig() (*rest.Config, error) {
	//.kube/config文件存在，就使用文件
	var kubeConfigFilePath string
	if home := homedir.HomeDir(); home != "" {
		kubeConfigFilePath = filepath.Join(home, ".kube", "config")
	}

	if os.Getenv("KUBECONFIG") != "" {
		kubeConfigFilePath = os.Getenv("KUBECONFIG")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			//当程序以pod方式运行时，就直接走这里的逻辑
			config, err = rest.InClusterConfig()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return config, nil
}
