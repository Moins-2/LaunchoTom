package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	prefixSession = "scanSession"
)

type session struct {
	tool        string
	description string
	command     string
	uuid        string
}

func session2string(s session) string {
	return prefixSession + "|" + s.tool + "|" + s.description + "|" + s.uuid
}

func logSession(s session) {

	logf("\n\ntool: %s", s.tool)
	logf("description: %s", s.description)
	logf("command: %s", s.command)
	logf("uuid: %s\n\n", s.uuid)
}

func getActiveSessions() []session {
	cmd := exec.Command("tmux", "list-sessions")
	output, err := cmd.Output()
	var sessionList []session
	if err != nil {
		//fmt.Println("Error getting output:", err)
		logf("output: %s", output)
	}

	// Get all session names in a []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, prefixSession) {
			sessionName := strings.Split(line, ":")[0]
			//append to the variable sessions
			sessionList = append(sessionList, string2session(sessionName))
			//logSession(string2session(sessionName))
		}
	}
	return sessionList

}

func getOutput(command session) (float64, bool) {
	sessionName := session2string(command)
	//logSession(command)
	//logf("Getting output for session: %s", sessionName)

	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", sessionName)
	output, err := cmd.Output()
	if err != nil {
		logf("Error getting output: %s", err)
		os.Exit(1)
	}
	progress, finished := ParseFfufOutput(string(output))
	//progressTmp, _ := ParseFloat(string(output))

	//fmt.Println("Progress: %f", progress*100)
	// same without decimals
	if progress > 0 {
		//logf("Progress:", int(progressTmp*100), "%")
	}

	return progress, finished
}

func endSession(command session) {
	sessionName := session2string(command)

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error ending session:", err)
		os.Exit(1)
	}
}

func string2session(s string) session {
	tool := strings.Split(s, "|")[1]
	description := strings.Split(s, "|")[2]
	uuid := strings.Split(s, "|")[3]
	// cut after the requiered number of character for uuid

	return session{
		tool:        tool,
		description: description,
		uuid:        uuid,
		command: func() string {
			for _, item := range menu.Items {
				if item.Title == tool {
					for _, subitem := range item.Items {
						// contains description
						if strings.Contains(subitem.Title, description) {
							//logf("Found command: %s", subitem.Content)
							return subitem.Content
						}
					}
				}
			}
			return ""
		}(),
	}

}
