package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"

	"os"
	"os/signal"
	"syscall"

	"sync"

	"k8s.io/api/events/v1beta1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func dumpObj(obj interface{}) {
	event := obj.(*v1beta1.Event)
	mutex.Lock()

	//fmt.Printf("{\"reason\":\"%s\", \"resource\":\"%s\", \"note\":\"%s\", \"time\":\"%s\"}\n", event.Reason, event.ObjectMeta.SelfLink, event.Note, time.Now().String())

	fmt.Fprintf(logfile, "{\"reason\":\"%s\", \"resource\":\"%s\", \"note\":\"%s\", \"time\":\"%s\"}\n", event.Reason, event.ObjectMeta.SelfLink, event.Note, time.Now().String())

	mutex.Unlock()
}

var logfile *os.File
var mutex sync.Mutex
var logfilename *string
var quit chan int = make(chan int)

func handleSignals() {
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	ilogfile, err := os.OpenFile(*logfilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	logfile = ilogfile
	go func() {
		defer func() {
			quit <- 1
		}()
		for {
			s := <-signal_chan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				mutex.Lock()
				logfile.Close()
				ilogfile, err := os.OpenFile(*logfilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					fmt.Println(err)
					return
				}
				logfile = ilogfile
				mutex.Unlock()
			default:
				mutex.Lock()
				logfile.Close()
				mutex.Unlock()
				return
			}
		}
	}()
}

func main() {
	kubemaster := flag.String("master", "k8smaster:8080", "server url")
	logfilename = flag.String("logfile", "/var/log/eventreader", "server url")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags(*kubemaster, "")
	if err != nil {
		glog.Errorln(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorln(err)
	}

	handleSignals()

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
	<-quit
}
