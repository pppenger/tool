package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"
)

var rootCmd = &cobra.Command{
	Use:   "daenerys",
	Short: "Generate Usability Code, which base on INKE Daenerys Framework.",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
}

func checkAndRunSelfUpdate() bool {
	if _, exist := os.LookupEnv("MODE"); exist {
		return false
	}
	defer func() {
		recover()
	}()

	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}
	git := gitlab.NewClient(&httpClient, "")
	git.SetBaseURL("https://git.inke.cn")
	// branch := "master"
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	stat, err := os.Stat(ex)
	if err != nil {
		return false
	}
	resp, _, err := git.Commits.ListCommits(7299, &gitlab.ListCommitsOptions{ListOptions: gitlab.ListOptions{Page: 0, PerPage: 1}}, nil)
	if err != nil || len(resp) == 0 {
		return false
	}
	if resp[0].AuthoredDate.After(stat.ModTime()) {
		fmt.Printf("Your tools is too old, begin self update, latest update@%s, local update@%s.\n", resp[0].AuthoredDate.Local().Format("2006/01/02 15:04:05"), stat.ModTime().Format("2006/01/02 15:04:05"))
		selfUpdate()
		return true
	}
	fmt.Printf("Your tools version is latest, skip self update.\n")
	return false
}

func Execute() {
	updated := checkAndRunSelfUpdate()
	if !updated  {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Dir, _ = os.Getwd()
	cmd.Args = os.Args[0:]
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "MODE=UPDATE")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)

}
