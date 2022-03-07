package main

import (
	"context"
	"io/ioutil"
	"kube-pod-terminator/internal/k8s"
	"kube-pod-terminator/internal/logging"
	"kube-pod-terminator/internal/options"
	"os"
	"strings"
	"time"

	"github.com/dimiro1/banner"
	"go.uber.org/zap"
)

var (
	logger            *zap.Logger
	kubeConfigPathArr []string
	kpto              *options.KubePodTerminatorOptions
)

func init() {
	kpto = options.GetKubePodTerminatorOptions()
	logger = logging.GetLogger()
	logger = logger.With(zap.Bool("inCluster", kpto.InCluster))

	bannerBytes, _ := ioutil.ReadFile("banner.txt")
	banner.Init(os.Stdout, true, false, strings.NewReader(string(bannerBytes)))
}

func main() {
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	kubeConfigPathArr = strings.Split(kpto.KubeConfigPaths, ",")
	for _, path := range kubeConfigPathArr {
		go func(p string) {
			logger = logger.With(zap.String("kubeConfigPath", p))
			logger.Info("starting generating clientset for kubeconfig")
			restConfig, err := k8s.GetConfig(p, kpto.InCluster)
			if err != nil {
				logger.Fatal("fatal error occurred while getting k8s config", zap.String("error", err.Error()))
			}

			clientSet, err := k8s.GetClientSet(restConfig)
			if err != nil {
				logger.Fatal("fatal error occurred while getting clientset", zap.String("error", err.Error()))
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(kpto.ContextTimeoutSeconds)*time.Second)
			defer cancel()
			k8s.Run(ctx, kpto.Namespace, clientSet, restConfig.Host)
			os.Exit(0)
		}(path)
	}

	select {}
}
