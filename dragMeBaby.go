package main

// A simple program that counts down from 5 and then exits.
// https://github.com/charmbracelet/bubbletea/blob/master/examples/realtime/main.go
import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	stopwatch                          time.Duration
	sub                                chan struct{}
	keys                               keyMap
	help                               help.Model
	active                             bool
	quitting                           bool
	stg                                int
	stgT                               times
	timer                              time.Time //	station    stations
	raceMsg                            string
	falseStart                         bool
	stageStyle, yellowStyle            lipgloss.Style
	greenStyle, greyStyle, jumpedStyle lipgloss.Style
	Help                               help.Model
}

type times struct {
	preStg  int
	fullStg int
	Yellow  float32
	Green   float32
}
type keyMap struct {
	Twenty key.Binding
	Quit   key.Binding
	Action key.Binding
	Help   key.Binding
	Reset  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help}, {k.Reset}, // second column
		{k.Quit, k.Action}, // first column
	}
}

var keys = keyMap{
	Action: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("(g)", "Action"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("(q)", "Quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("(?)", "toggle help"),
	),
	Reset: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("(r)", "Reset"),
	),
}

func newModel() model {
	return model{
		sub:         make(chan struct{}),
		keys:        keys,
		help:        help.New(),
		stg:         0,
		stgT:        times{preStg: 2, fullStg: 2, Yellow: 1.2, Green: .400},
		stageStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		yellowStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")),
		greenStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		greyStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#798486")),
		jumpedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#EF333F")),
	}
}
func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenForActivity(m.sub), // generate activity
		waitForActivity(m.sub),   // wait for activity
	)
}

type responseMsg struct{}

func listenForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		for {
			//time.Sleep(time.Millisecond * 100) // nolint:gosec
			sub <- struct{}{}
		}
	}
}
func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	//This needs to be done right so that there is a constant running even that sends a message when changed.
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width

	// make better logic you dunce
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Action):
			if m.stg == 4 {
				x := time.Now()
				m.stopwatch = x.Sub(m.timer)
				m.stg++
				m.timer = time.Now()
				m.active = true
				m.raceMsg = " --FINISHED--"
			} else {
				m.stg = 99
				m.active = false
				m.falseStart = true
				m.raceMsg = "  --FALSE START--"
				m.timer = time.Now()

			}
		case key.Matches(msg, m.keys.Reset):
			m.stg = 99
			m.active = false
			m.raceMsg = "   --RESET--"
			m.timer = time.Now()

		}
		return m, waitForActivity(m.sub)
	case responseMsg:
		switch {
		case !m.active && m.raceMsg != "":
			x := m.timer.Add(time.Second * 5)
			current := time.Now()
			if current.After(x) {
				m.stg = 0
				m.timer = current
				m.raceMsg = ""

			}

		case !m.active && m.stg == 0:
			now := time.Now()
			m.timer = now
			m.active = true
			m.keys.Action.SetEnabled(true)
			m.falseStart = false
		case m.active:
			switch {

			case m.stg == 0:
				x := m.setStgtime(m.stgT.preStg)
				// x := m.timer.Add(time.Millisecond * time.Duration(1000*m.stgT.preStg))
				current := time.Now()
				if current.After(x) {
					m.stg++
					m.timer = current
				}

			case m.stg == 1:
				x := m.timer.Add(time.Millisecond * time.Duration(1000*m.stgT.fullStg))
				current := time.Now()
				if current.After(x) {
					m.stg++
					m.timer = current
				}

			case m.stg == 2:
				x := m.timer.Add(time.Millisecond * time.Duration(1000*m.stgT.Yellow))
				current := time.Now()
				if current.After(x) {
					m.stg++
					m.timer = current
				}
			case m.stg == 3:
				x := m.timer.Add(time.Millisecond * time.Duration(1000*m.stgT.Green))
				current := time.Now()
				if current.After(x) {
					m.stg++
					m.timer = current

				}
			case m.stg == 4:
				m.timer = time.Now()
				m.active = false

			case m.stg == 5:
				x := m.timer.Add(time.Second * 5)
				current := time.Now()
				if current.After(x) {
					m.stg = 0
					m.active = false
					m.stopwatch = 0

				}
			}
		}
	}
	return m, waitForActivity(m.sub)
}

// Main Function
func main() {

	if os.Getenv("HELP_DEBUG") != "" {
		f, err := tea.LogToFile("debug.log", "help")
		if err != nil {
			fmt.Println("Couldn't open a file for logging:", err)
			os.Exit(1)
		}
		defer f.Close() // nolint:errcheck
	}

	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}

// this is super ugly
func (m model) View() string {
	var raceTime string
	var tree string
	var line1, stage1, stage2, line2, yellows, line3, greens, bottoms string

	helpView := m.help.View(m.keys)

	if m.stopwatch > 0 {
		raceTime = "Elapsed Time: " + m.stopwatch.String()
	}

	line1 = "______________"
	if !m.falseStart {
		if m.stg >= 1 {
			stage1 = "|(" + m.stageStyle.Render("oo") + ")=||=(" + m.stageStyle.Render("oo") + ")|"
		} else {
			stage1 = "|(" + m.greyStyle.Render("oo") + ")=||=(" + m.greyStyle.Render("oo") + ")|"
		}

		if m.stg >= 2 {
			stage2 = "|(" + m.stageStyle.Render("oo") + ")=||=(" + m.stageStyle.Render("oo") + ")|"
		} else {
			stage2 = "|(" + m.greyStyle.Render("oo") + ")=||=(" + m.greyStyle.Render("oo") + ")|"

		}

		if m.stg >= 3 {
			yellows = " |(" + m.yellowStyle.Render("0") + ")=||=(" + m.yellowStyle.Render("0") + ")|"
		} else {
			yellows = " |(" + m.greyStyle.Render("0") + ")=||=(" + m.greyStyle.Render("0") + ")|"

		}

		if m.stg >= 4 {
			greens = " |(" + m.greenStyle.Render("0") + ")=||=(" + m.greenStyle.Render("0") + ")|"
		} else {
			greens = " |(" + m.greyStyle.Render("0") + ")=||=(" + m.greyStyle.Render("0") + ")|"
		}
	} else {
		stage1 = "|(" + m.jumpedStyle.Render("oo") + ")=||=(" + m.jumpedStyle.Render("oo") + ")|"
		stage2 = "|(" + m.jumpedStyle.Render("oo") + ")=||=(" + m.jumpedStyle.Render("oo") + ")|"
		yellows = " |(" + m.jumpedStyle.Render("0") + ")=||=(" + m.jumpedStyle.Render("0") + ")|"
		greens = " |(" + m.jumpedStyle.Render("0") + ")=||=(" + m.jumpedStyle.Render("0") + ")|"

	}

	line2 = "  =========="
	line3 = " |====||====|"
	bottoms = `  ==========
     ||||
     ||||
     ||||
     ||||
     ||||
     ||||
--------------
`
	// this is also ugly
	height := 2 - strings.Count(tree, "\n") - strings.Count(helpView, "\n")
	if !m.falseStart {
		tree = line1 + "\n" + stage1 + "\n" + stage2 + "\n" + line2 + "\n" + yellows + "\n" + yellows + "\n" + yellows + "\n" + line3 + "\n" + greens + "\n" + bottoms + "\n" + raceTime
	} else {
		tree = line1 + "\n" + stage1 + "\n" + stage2 + "\n" + line2 + "\n" + yellows + "\n" + yellows + "\n" + yellows + "\n" + line3 + "\n" + greens + "\n" + bottoms + "\n" + raceTime

	}
	return "\n" + m.raceMsg + "\n" + tree + strings.Repeat("\n", height) + helpView
}

func (m model) setStgtime(stage int) time.Time {
	x := m.timer.Add(time.Millisecond * time.Duration(1000*stage))
	return x
}

//  ____________
// |(oo)=||=(oo)|
// |(oo)=||=(oo)|
//   ==========
//  |(O)=||=(O)|
//  |(O)=||=(O)|
//  |(O)=||=(O)|
//  |====||====|
//  |(O)=||=(O)|
//   ==========
//      ||||
//      ||||
//      ||||
//      ||||
//      ||||
//      ||||
// --------------
