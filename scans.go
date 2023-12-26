package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/muesli/reflow/indent"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

const DEBUG = false

const listHeight = 14

var (
	menu = initItemsListFile()
)

type (
	tickMsg  struct{}
	frameMsg struct{}
)

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s (%s)", index+1, i.Title, i.Content)
	if i.Items != nil {
		str += " - SubMenu available"
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, fn(str))
}

func ParseFfufOutput(s string) (float64, bool) {
	// get the line with :: Progress: [150/578] :: Job [1/1] :: 100 req/sec :: Duration: [0:00:01] :: Errors: 150 ::

	//fmt.Println("Parsing:", s)
	// go through each line, starting from the end
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.Contains(line, "Progress") {
			// get the number between the brackets
			re := regexp.MustCompile(`\[(.*?)\]`)
			matches := re.FindAllStringSubmatch(line, -1)
			if len(matches) > 0 {
				// get the first match
				match := matches[0]
				if len(match) > 1 {
					// get the first group
					group := match[1]
					// split on the slash
					parts := strings.Split(group, "/")
					if len(parts) > 0 {
						// get the first part
						done := parts[0]

						// convert to float
						doneFloat, err := strconv.ParseFloat(done, 64)
						if err != nil {
							return -1, false
						}
						total := parts[1]
						totalFloat, err := strconv.ParseFloat(total, 64)
						if err != nil {
							return -1, false
						}
						return doneFloat / totalFloat, false
					}
				}
			}
		} else if isLastLine(line) {
			return 1, true
		}
	}

	return -2, false
}

func listTool() []string {
	var listTool []string
	for _, item := range menu.Items {
		listTool = append(listTool, item.Title)
	}

	return listTool
}

type model struct {
	list     list.Model
	quitting bool
	finished bool

	// progress bar
	Ticks    int
	Frames   int
	Progress float64
	Loaded   bool

	// command
	ActiveSession session
	listSessions  []session
}

var (
	master_pty, slave_pty *os.File
)

func (m model) Init() tea.Cmd {

	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "ctrl+c" || k == "q" || k == "esc" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	if !m.finished {
		return updateCommandChoice(msg, m)
	}
	return updateProgressBar(msg, m)
}

// Progress bar update
func updateProgressBar(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case frameMsg:
		if !m.Loaded {
			//	logf("progress: %f", m.Progress)
			progress, finished := getOutput(m.ActiveSession)
			if progress >= 0 {
				// Update the progress directly in the model
				m.Progress = progress
			}

			if finished {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 2
				return m, tick()
			}

			return m, frame()
		}

	case tickMsg:

		if m.Loaded {
			if m.Ticks == 0 {
				endSession(m.ActiveSession)

				m.quitting = true
				return m, tea.Quit
			}
			m.Ticks--
			return m, tick()
		}
	}

	return m, nil

}

// Choice of the command
func updateCommandChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {

		case "enter":

			selectedItem, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			newList := list.New(convertToItems(selectedItem.Items), itemDelegate{}, m.list.Width(), listHeight)
			if selectedItem.Items == nil {
				m.finished = true
				titleForSession := m.list.Title

				newList.Title = m.list.Title + " - " + selectedItem.Title
				m.list = newList

				if isUUID(selectedItem.Content) {
					m.ActiveSession = session{
						tool:        strings.Split(selectedItem.Title, " - ")[0],
						description: strings.Split(selectedItem.Title, " - ")[1],
						command:     "Can't get the command for now",
						uuid:        selectedItem.Content,
					}
				} else {
					m.ActiveSession = session{
						tool:        titleForSession,
						description: selectedItem.Title,
						command:     selectedItem.Content,
						uuid:        uuid.New().String(),
					}
					launchCommand(m.ActiveSession)
				}

				return m, frame()
			}
			newList.Title = selectedItem.Title
			m.list = newList

			return m, nil
		}
	case tickMsg:
		/* if m.Ticks == 0 {
			m.quitting = true
			return m, tea.Quit
		}
		m.Ticks--
		return m, tick() */
	}

	m.list, _ = m.list.Update(msg)

	return m, nil
}

func (m model) View() string {
	var s string
	if m.quitting {
		return quitTextStyle.Render(m.list.Title + "\n\nSee you later!\n")
	}
	if m.finished {
		s = progressBarView(m)
	} else {

		s = commandListView(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

func commandListView(m model) string {
	if len(m.list.Items()) == 0 {
		return quitTextStyle.Render(fmt.Sprintf("%s selected.", m.list.Title))
	}
	return "\n" + m.list.View()
}

func progressBarView(m model) string {
	label := "Scanning..."
	if m.Loaded {
		label = fmt.Sprintf("Scanned. Exiting in %s seconds...", colorFg(strconv.Itoa(m.Ticks), "79"))
	}
	msg := m.list.Title
	if m.Loaded {
		msg = "Done!"
	}

	return msg + "\n\n" + label + "\n" + progressbar(m.Progress) + "%"
}

func main() {

	m := model{list: initList(), Ticks: 10, Frames: 10, Progress: 0, Loaded: false}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
