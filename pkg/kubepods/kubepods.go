package kubepods

import (
	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"elastic-gpu-exporter/pkg/util"
	"time"

	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/cache"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const(
	RecommendedKubeConfigPathEnv = "KUBECONFIG"
)

type Handler struct {
	AddFunc func(pod *v1.Pod)
	DelFunc func(pod *v1.Pod)
}

type Watcher interface {
	Run(stop <-chan struct{})
}

type KubeWatcher struct {
	labelSet     map[string]struct{}
	node         string
	client       *kubernetes.Clientset
	informers    informers.SharedInformerFactory
	podInformers cache.SharedIndexInformer
	podLister    v12.PodLister
	handler      *Handler
}

func NewWatcher(handler *Handler, gpuLabels []string, node string) Watcher {
	//var kubeconfig *string
	//if home := homedir.HomeDir(); home != "" {
	//	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()
	//
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	klog.Fatalf("Could not get config")
	//}
    //------------------
	// create the clientset
	//clientset, err = kubernetes.NewForConfig(restConfig)

	// Grab a dynamic interface that we can create informers from
	//dc, err := dynamic.NewForConfig(cfg)
	//if err != nil {
	//	logrus.WithError(err).Fatal("could not generate dynamic client for config")
	//}


	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("create watcher failed: %s", err.Error())
	}

	//------------
	//kubeConfig := ""
	//if len(os.Getenv(RecommendedKubeConfigPathEnv)) > 0 {
	//	// use the current context in kubeconfig
	//	// This is very useful for running locally.
	//	kubeConfig = os.Getenv(RecommendedKubeConfigPathEnv)
	//}
	//
	//// Get kubernetes config.
	//restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	//if err != nil {
	//	klog.Fatalf("Error building kubeconfig: %s", err.Error())
	//}
	//
	//// create the clientset
	//client, err := kubernetes.NewForConfig(restConfig)
	//if err != nil {
	//	klog.Fatalf("Failed to init rest config due to %v", err)
	//}
	//--------------
	client, _ := kubernetes.NewForConfig(config)
	informersFactory := informers.NewSharedInformerFactoryWithOptions(client, time.Second, informers.WithTweakListOptions(nodeNameFilter(node)))
	labelSet := make(map[string]struct{})
	for _, label := range gpuLabels {
		labelSet[label] = struct{}{}
	}
	return &KubeWatcher{
		labelSet:     labelSet,
		node:         node,
		client:       client,
		informers:    informersFactory,
		podInformers: informersFactory.Core().V1().Pods().Informer(),
		handler:      handler,
	}
}

func (w *KubeWatcher) Run(stop <-chan struct{}) {
	w.podInformers.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Errorf("Cannot convert to *v1.Pod: %t %v", obj, obj)
				return
			}
			if !util.PodHasResource(pod, w.labelSet) {
				return
			}
			w.handler.AddFunc(pod)
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Errorf("Cannot convert to *v1.Pod: %t %v", obj, obj)
				return
			}
			if !util.PodHasResource(pod, w.labelSet) {
				return
			}
			w.handler.DelFunc(pod)
		},
	})
	w.informers.Start(stop)
	w.informers.WaitForCacheSync(stop)
}

func nodeNameFilter(nodeName string) func(options *metav1.ListOptions) {
	return func(options *metav1.ListOptions) {
		options.FieldSelector = fields.OneTermEqualSelector(util.NodeNameField, nodeName).String()
	}
}
