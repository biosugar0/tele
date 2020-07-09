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
		Long:  `simple Telepresence wrapper tool for development microservices`,
		RunE:  Run,
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
	fmt.Println("repository =", repo)

	branch, err := execute(`git rev-parse --abbrev-ref @`)
	if err != nil {
		return err
	}
	fmt.Println("branch = ", branch)

	user := params.User
	fmt.Println("user = ", user)

	port := params.ServerPort

	deployment := strings.Join([]string{
		user,
		repo,
		branch,
	}, "-")
	deployment = util.ToValidName(deployment)

	fmt.Println(deployment)

	run := params.CMD

	fmt.Println(run)

	telepresence := fmt.Sprintf(
		"telepresence --namespace default --method inject-tcp --new-deployment %s --expose %s:%s --run bash -c \"%s\"",
		deployment,
		port,
		port,
		run,
	)

	fmt.Println("==============")
	fmt.Println(telepresence)
	fmt.Println("==============")

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
	rootCmd.Flags().StringVar(&params.ServerPort, "port", "5000", "http server port")
	rootCmd.Flags().StringVar(&params.User, "user", homedir, "user name")
	rootCmd.Flags().StringVar(&params.CMD, "run", "go run main.go", "shell command")
	rootCmd.Execute()
}
