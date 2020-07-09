package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/biosugar0/tele/params"
	"github.com/biosugar0/tele/pkg/util"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tele",
		Short: "simple Telepresence wrapper tool for development microservices",
		Long: `tele
 A simple Telepresence wrapper tool for microservice development.
 This command uses the --new-deployment option of telepresense, as shown below example.

[example]
 telepresence --namespace {--namespace} --method inject-tcp --new-deployment {--user}-{repository}-{branch} --expose {--port} --run bash -c \"{--run}\"
`,
		RunE: Run,
	}
)

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

func Run(cmd *cobra.Command, args []string) error {
	repository, err := execute("git rev-parse --show-toplevel")
	if err != nil {
		return err
	}
	repo := filepath.Base(repository)
	fmt.Printf("[repository]:\n %s\n", repo)

	branch, err := execute(`git rev-parse --abbrev-ref @`)
	if err != nil {
		return err
	}
	fmt.Printf("[branch]:\n %s\n", branch)

	user := params.User
	fmt.Printf("[user]:\n %s\n", user)

	namespace := params.NameSpace
	fmt.Printf("[namespace]:\n %s\n", namespace)

	port := params.ServerPort
	fmt.Printf("[port]:\n %s\n", port)

	deployment := strings.Join([]string{
		user,
		repo,
		branch,
	}, "-")
	deployment = util.ToValidName(deployment)

	fmt.Printf("[deployment]:\n %s\n", deployment)

	run := params.CMD

	fmt.Printf("[request command]:\n %s\n", run)

	telepresence := fmt.Sprintf(
		"telepresence --namespace %s --method inject-tcp --new-deployment %s --expose %s --run bash -c \"%s\"",
		namespace,
		deployment,
		port,
		run,
	)

	fmt.Printf("[Telepreesence command]:\n %s\n", telepresence)

	fmt.Printf("[result]:\n ")
	result, err := execute(telepresence)
	if err != nil {
		return err
	}
	fmt.Println(result)

	return nil
}

func main() {
	homedir := filepath.Base(os.Getenv("HOME"))
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringVar(&params.ServerPort, "port", "5004:5004", "http server port")
	rootCmd.Flags().StringVar(&params.User, "user", homedir, "user name for prefix of deployment name. default is home directory name")
	rootCmd.Flags().StringVar(&params.CMD, "run", "go run main.go", "shell command")
	rootCmd.Flags().StringVar(&params.NameSpace, "namespace", "default", "name space of kubernetes")
	rootCmd.Execute()
}
