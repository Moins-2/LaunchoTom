package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type commandFinishedMsg struct{ err error }

func launchCommand(command session) {

	sessionName := session2string(command)
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName)
	err := cmd.Run()
	if err != nil {
		logf("Error creating session:", err)
	}

	cmd = exec.Command("tmux", "send-keys", "-t", sessionName)
	cmd.Args = append(cmd.Args, command.command+";echo 'scans: Command finished'")
	cmd.Args = append(cmd.Args, "C-m")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error sending keys:", err)
		os.Exit(1)
	}
}

func isLastLine(line string) bool {
	if strings.Contains(line, "scans: Command finished") {
		if !strings.Contains(line, "echo") {
			return true
		}
	}
	return false
}
