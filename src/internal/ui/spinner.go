package ui

import (
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/spinner"
)

// Get the Dot spinner type
var dotSpinner = spinner.Dot

// liveSpinnerModel is a custom bubbletea model that supports live-updating the title.
type liveSpinnerModel struct {
	spin      spinner.Model
	title     string
	mu        sync.RWMutex
	done      chan error
	actionDone chan struct{}
}

func (m *liveSpinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spin.Tick,
		tea.Cmd(func() tea.Msg {
			<-m.actionDone
			return doneMsg{}
		}),
	)
}

func (m *liveSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		return m, tea.Quit
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Interrupt
		}
	}
	var cmd tea.Cmd
	m.spin, cmd = m.spin.Update(msg)
	return m, cmd
}

func (m *liveSpinnerModel) View() tea.View {
	m.mu.RLock()
	title := m.title
	m.mu.RUnlock()
	return tea.NewView(m.spin.View() + " " + title)
}

// UpdateTitle updates the spinner's title dynamically.
func (m *liveSpinnerModel) UpdateTitle(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.title = title
}

type doneMsg struct{}

// RunLiveSpinner runs an action with a live-updating spinner.
// The action function receives a callback to update the spinner's title.
// The spinner displays the title dynamically as the action progresses.
func RunLiveSpinner(title string, action func(updateTitle func(string)) error) error {
	m := &liveSpinnerModel{
		spin:       spinner.New(spinner.WithSpinner(dotSpinner)),
		title:      title,
		done:       make(chan error, 1),
		actionDone: make(chan struct{}),
	}

	// Run the action in a goroutine
	go func() {
		err := action(func(newTitle string) {
			m.UpdateTitle(newTitle)
		})
		m.done <- err
		close(m.actionDone)
	}()

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Wait for the action to complete and get the error
	actionErr := <-m.done
	
	// Check if the final model has an error
	if finalModel != nil {
		if mm, ok := finalModel.(*liveSpinnerModel); ok {
			_ = mm
		}
	}
	
	return actionErr
}

// RunStaticSpinner executes an action with a simple spinner.
// Use this for operations where you don't have progress feedback.
func RunStaticSpinner(title string, action func() error) error {
	m := &liveSpinnerModel{
		spin:       spinner.New(),
		title:      title,
		done:       make(chan error, 1),
		actionDone: make(chan struct{}),
	}

	go func() {
		err := action()
		m.done <- err
		close(m.actionDone)
	}()

	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return err
	}

	return <-m.done
}
