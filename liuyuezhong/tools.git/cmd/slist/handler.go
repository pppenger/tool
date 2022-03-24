package main

import (
	"fmt"
	"github.com/fatih/color"
	"os/exec"
)

func Run(ch chan CommandResult, cmd []string, host string) {
	go func() {
		c := exec.Command(cmd[0], cmd[1:]...)
		output, err := c.CombinedOutput()
		ch <- CommandResult{
			cmd:  c.String(),
			err:  err,
			data: output,
			host: host,
		}
	}()
}

// BatchRun 在远程机器上批量执行输出
func BatchRun(infos []ServiceInfo, cmd string) {
	ssh := []string{"ssh"}
	if sshOption != "" {
		ssh = append(ssh, "-o", sshOption)
	}

	_, _ = color.New(color.FgBlue).Printf("query service[%s] in %d hosts \n\n", serviceName, len(infos))
	ch := make(chan CommandResult, len(infos))

	num := 0
	for _, h := range infos {
		if IgnoreHost(h.HostName) {
			if verbose {
				fmt.Println("ignore hostname: ", h.HostName)
			}

			continue
		}

		subCmd := fmt.Sprintf("cd %s", h.Path)
		if len(cmd) > 0 {
			subCmd += " && " + cmd
		}

		var cmd []string
		cmd = append(cmd, ssh...)
		cmd = append(cmd, fmt.Sprintf("%s@%s", h.User, h.HostName), subCmd)

		Run(ch, cmd, h.HostName)
		num++
	}

	for count := 0; count < num; count++ {
		select {
		case result := <-ch:
			if len(result.data) == 0 && !verbose {
				continue
			}

			if !rawMode {
				_, _ = color.New(color.FgGreen).Println(">>>>>>> ", result.host)
			}

			if verbose {
				_, _ = color.New(color.FgBlue).Println("|exec cmd: ", result.cmd)
			}
			if len(result.data) > 0 {
				fmt.Print(string(result.data))
			}

			if result.err != nil && !rawMode {
				_, _ = color.New(color.FgRed).Println("|err: ", result.err)
			}
		}
	}
}
