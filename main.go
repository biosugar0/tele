package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/biosugar0/tele/params"
	"github.com/biosugar0/tele/pkg/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	// PodTerminating means that the pod is being terminated by the system.
	PodTerminating v1.PodPhase = "Terminating"
)

var (
	rootCmd = &cobra.Command{
		Example:               "tele --port 8080:8080 go run main.go",
		DisableFlagsInUseLine: true,
		Version:               params.Version,
		Use:                   `tele [flags] <shell command>`,
		Short:                 "simple Telepresence wrapper tool for development microservices",
		Long: `A simple Telepresence wrapper tool for microservice development.

 Find more information at: https://github.com/biosugar0/tele
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a command string. example: tele go run main.go")
			}
			return nil
		},
		RunE: Run,
	}
)

type k8sClient struct {
	Client       *kubernetes.Clientset
	ResourceName *string
	NameSpace    string
}

type Resource struct {
	NameSpace  string
	Deployment *string
}

var (
	resourceStore = Resource{}
)

var stopSignalReceived = make(chan bool, 1)
var completeCommand = make(chan bool, 1)

func init() {
	rootCmd.SetOutput(os.Stdout)
}

func execute(cmdstr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Env = os.Environ()
	var sout bytes.Buffer
	var serr bytes.Buffer
	cmd.Stdout = &sout
	cmd.Stderr = &serr
	err := cmd.Run()
	if err != nil {
		return "", errors.New(serr.String())
	}
	result := strings.TrimRight(sout.String(), "\n")
	return result, nil
}

func executeStream(cmdstr string) error {
	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	streamReader := func(scanner *bufio.Scanner, output chan string, done chan bool) {
		defer close(output)
		defer close(done)
		for scanner.Scan() {
			output <- scanner.Text()
		}
		done <- true
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stdoutOutputChan := make(chan string)
	stdoutDoneChan := make(chan bool)
	stderrScanner := bufio.NewScanner(stderr)
	stderrOutputChan := make(chan string)
	stderrDoneChan := make(chan bool)
	go streamReader(stdoutScanner, stdoutOutputChan, stdoutDoneChan)
	go streamReader(stderrScanner, stderrOutputChan, stderrDoneChan)

	go func() {
		for range stopSignalReceived {
			if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
				fmt.Println("\n[tele]: shut down command")
				if e := cmd.Process.Signal(os.Interrupt); e != nil {
					fmt.Printf("err: %s", e)
					err = e
				}
			}
			completeCommand <- false
		}
	}()

	go func() {
		for {
			if cmd.ProcessState != nil {
				if cmd.ProcessState.Exited() {
					completeCommand <- true
					return
				}
			}
		}
	}()

	go func() {
		runnning := true
		for runnning {
			select {
			case <-stdoutDoneChan:
				runnning = false
			case line := <-stdoutOutputChan:
				fmt.Println(line)
			case line := <-stderrOutputChan:
				fmt.Println(line)
			}
		}
	}()

	if e := cmd.Wait(); e != nil {
		fmt.Printf("err: %s", e)
		err = e
	}

	if err != nil {
		return err
	}

	return nil
}

func Run(cmd *cobra.Command, args []string) error {
	repository, err := execute("git rev-parse --show-toplevel")
	if err != nil {
		return err
	}
	repo := filepath.Base(repository)
	cmd.Printf("[repository]:\n %s\n", repo)

	branch, err := execute(`git rev-parse --abbrev-ref @`)
	if err != nil {
		return err
	}
	cmd.Printf("[branch]:\n %s\n", branch)

	user := params.User
	cmd.Printf("[user]:\n %s\n", user)

	namespace := params.NameSpace
	cmd.Printf("[namespace]:\n %s\n", namespace)

	port := params.ServerPort
	cmd.Printf("[port]:\n %s\n", port)

	deployment := strings.Join([]string{
		user,
		repo,
		branch,
	}, "-")
	deployment = util.ToValidName(deployment)

	cmd.Printf("[deployment]:\n %s\n", deployment)

	cmdArgs := []string{}
	for _, v := range args {
		v := util.SpecialStr(v)
		cmdArgs = append(cmdArgs, v)
	}

	run := strings.Join(cmdArgs, " ")

	cmd.Printf("[request command]:\n %s\n", run)

	telepresence := fmt.Sprintf(
		"telepresence --namespace %s --method inject-tcp --new-deployment %s",
		namespace,
		deployment,
	)
	resourceStore.Deployment = &deployment
	resourceStore.NameSpace = namespace
	if len(port) > 0 {
		telepresence += fmt.Sprintf(" --expose %s", port)
	}
	telepresence += fmt.Sprintf(
		" --run bash -c \"%s\"",
		run,
	)

	if params.Sudo {
		telepresence = fmt.Sprintf("sudo %s", telepresence)
	}

	cmd.Printf("[Telepreesence command]:\n %s\n", telepresence)

	cmd.Printf("[result]:\n")
	err = executeStream(telepresence)
	if err != nil {
		return err
	}

	complete := <-completeCommand

	cleanup(resourceStore)

	switch complete {
	case true:
		fmt.Println("\n[tele]: complete")
	default:
		fmt.Println("\n[tele]: command killed")
	}

	return nil
}

func newClient(name *string) k8sClient {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
	}
	return k8sClient{
		Client:       clientset,
		ResourceName: name,
		NameSpace:    params.NameSpace,
	}
}

func (r *k8sClient) cleanService() error {
	if r.ResourceName == nil {
		return nil
	}
	service, err := r.Client.CoreV1().Services(r.NameSpace).Get(context.TODO(), *r.ResourceName, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		fmt.Println(err.Error())
		return nil
	}
	if service != nil {
		err = r.Client.CoreV1().Services(r.NameSpace).Delete(context.TODO(), *r.ResourceName, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("service has been deleted")
	}

	return nil
}

func (r *k8sClient) podPhase() *v1.PodPhase {
	pod, err := r.Client.CoreV1().Pods(r.NameSpace).Get(context.TODO(), *r.ResourceName, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		fmt.Println(err.Error())
		return nil
	}
	if pod != nil {
		st := pod.Status
		if pod.ObjectMeta.DeletionTimestamp != nil {
			v := PodTerminating
			return &v
		}
		return &st.Phase
	}
	fmt.Printf("%v:%v", pod.Name, pod.Status.Phase)

	return nil
}

func (r *k8sClient) cleanPod() error {
	time.Sleep(1 * time.Second)
	pods, err := r.Client.CoreV1().Pods(r.NameSpace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if pods != nil {
		podlist := *pods
		terminating := 0
		for _, v := range podlist.Items {
			if v.Name == *r.ResourceName {
				runningTime := 0

				for {
					finish := false

					time.Sleep(1 * time.Second)

					runningTime++
					switch phase := r.podPhase(); {
					case phase == nil:
						fmt.Println("pod has been terminated")
						finish = true
					case *phase == PodTerminating:
						terminating++
						if terminating == 1 {
							fmt.Println("waiting treminate pod...")
						}
					case *phase == v1.PodPending:
						err = r.Client.CoreV1().Pods(r.NameSpace).Delete(context.TODO(), *r.ResourceName, metav1.DeleteOptions{})
						if err != nil {
							fmt.Println(err.Error())
						}
						finish = true
					case *phase == v1.PodRunning:
						if runningTime > 3 {
							err = r.Client.CoreV1().Pods(r.NameSpace).Delete(context.TODO(), *r.ResourceName, metav1.DeleteOptions{})
							if err != nil {
								fmt.Println(err.Error())
							}
						}
					default:
						finish = true
					}

					if finish {
						break
					}
				}

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func cleanup(resource Resource) {
	fmt.Println("tele resource clening")
	client := newClient(resource.Deployment)

	if resource.Deployment != nil {
		err := client.cleanPod()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = client.cleanService()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func main() {
	homedir := filepath.Base(os.Getenv("HOME"))
	rootCmd.Flags().SetInterspersed(false)
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&params.ServerPort, "port", "", "expose http server port")
	rootCmd.PersistentFlags().StringVar(&params.User, "user", homedir, "user name for prefix of deployment name. default is home directory name")
	rootCmd.PersistentFlags().StringVar(&params.NameSpace, "namespace", "default", "name space of kubernetes")
	rootCmd.PersistentFlags().BoolVar(&params.Sudo, "sudo", false, "execute commands as a super user")

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		stopSignalReceived <- true
	}()

	if err := rootCmd.Execute(); err != nil {
		rootCmd.SetOutput(os.Stderr)
		rootCmd.Println(err)
		os.Exit(1)
	}

}
