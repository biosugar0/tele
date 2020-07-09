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
		Version: params.Version,
		Use:     `tele --run "<shell command>"`,
		Short:   "simple Telepresence wrapper tool for development microservices",
		Long: `A simple Telepresence wrapper tool for microservice development.

 Find more information at: https://github.com/biosugar0/tele
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
		"telepresence --namespace %s --method inject-tcp --new-deployment %s",
		namespace,
		deployment,
	)
	if len(port) > 0 {
		telepresence += fmt.Sprintf(" --expose %s", port)
	}
	telepresence += fmt.Sprintf(
		" --run bash -c \"%s\"",
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
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&params.CMD, "run", "echo hello world", "shell command")
	rootCmd.PersistentFlags().StringVar(&params.ServerPort, "port", "", "expose http server port")
	rootCmd.PersistentFlags().StringVar(&params.User, "user", homedir, "user name for prefix of deployment name. default is home directory name")
	rootCmd.PersistentFlags().StringVar(&params.NameSpace, "namespace", "default", "name space of kubernetes")
	rootCmd.Execute()
}
