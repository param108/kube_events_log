package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/events/v1beta1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func dumpObj(obj interface{}) {
	event := obj.(*v1beta1.Event)
	//fmt.Println(event.Reason)
	fmt.Printf("{\"reason\":\"%s\", \"resource\":\"%s\", \"note\":\"%s\", \"time\":\"%s\"}\n", event.Reason, event.ObjectMeta.SelfLink, event.Note, time.Now().String())
}

func main() {
	kubemaster := flag.String("master", "k8smaster:8080", "absolute path to the kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags(*kubemaster, "")
	if err != nil {
		glog.Errorln(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorln(err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientset, time.Second*30)
	svcInformer := kubeInformerFactory.Events().V1beta1().Events().Informer()

	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dumpObj,
		DeleteFunc: dumpObj,
		UpdateFunc: func(oldObj, newObj interface{}) {
			dumpObj(newObj)
		},
	})

	stop := make(chan struct{})
	defer close(stop)
	kubeInformerFactory.Start(stop)
	for {
		time.Sleep(time.Second)
	}
}
