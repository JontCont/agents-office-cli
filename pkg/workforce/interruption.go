package workforce

import (
	"errors"
	"fmt"
	"sync"
)

type Coordinator struct {
	state          RunState
	mu             sync.RWMutex
	resumeChan     chan string
	abortChan      chan struct{}
	startChan      chan string
	onStateChange  func(RunState)
	onMessage      func(sender, content string)
}

func NewCoordinator(onStateChange func(RunState), onMessage func(sender, content string)) *Coordinator {
	return &Coordinator{
		state:         StateQueued,
		resumeChan:    make(chan string, 1),
		abortChan:     make(chan struct{}, 1),
		startChan:     make(chan string, 1),
		onStateChange: onStateChange,
		onMessage:     onMessage,
	}
}

func (c *Coordinator) GetState() RunState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Coordinator) StartRun(prompt string) error {
	c.mu.Lock()
	if c.state == StateCompleted || c.state == StateFailed || c.state == StateCancelled {
		c.state = StateQueued
		c.resumeChan = make(chan string, 1)
		c.abortChan = make(chan struct{}, 1)
		for len(c.startChan) > 0 {
			<-c.startChan
		}
	}
	if c.state != StateQueued {
		c.mu.Unlock()
		return fmt.Errorf("cannot start run: state is not QUEUED")
	}
	c.mu.Unlock()
	c.startChan <- prompt
	return nil
}

func (c *Coordinator) GetStartPromptChan() chan string {
	return c.startChan
}

func (c *Coordinator) Transition(to RunState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	valid := false
	switch c.state {
	case StateQueued:
		valid = (to == StateRunning)
	case StateRunning:
		valid = (to == StateInterrupting || to == StateCompleted || to == StateFailed || to == StateInterrupted)
	case StateInterrupting:
		valid = (to == StateInterrupted || to == StateCancelled)
	case StateInterrupted:
		valid = (to == StateResuming || to == StateCancelled)
	case StateResuming:
		valid = (to == StateRunning)
	}

	if !valid {
		return fmt.Errorf("invalid transition from %s to %s", c.state, to)
	}

	c.state = to
	if c.onStateChange != nil {
		c.onStateChange(to)
	}
	return nil
}

func (c *Coordinator) Interrupt() error {
	return c.Transition(StateInterrupting)
}

func (c *Coordinator) Resume(feedback string) error {
	if c.GetState() != StateInterrupted {
		return errors.New("cannot resume: state is not INTERRUPTED")
	}
	c.resumeChan <- feedback
	return nil
}

func (c *Coordinator) Abort() error {
	state := c.GetState()
	if state != StateInterrupted && state != StateInterrupting {
		return errors.New("cannot abort: state must be INTERRUPTED or INTERRUPTING")
	}
	c.abortChan <- struct{}{}
	return nil
}

// AskHuman is called by an agent to trigger a breakpoint.
func (c *Coordinator) AskHuman() error {
	return c.Transition(StateInterrupted)
}

// CheckStepBoundary halts the run loop if an interruption was requested.
// Returns true if execution should continue, false if aborted.
func (c *Coordinator) CheckStepBoundary() bool {
	c.mu.Lock()
	if c.state == StateInterrupting {
		c.mu.Unlock()
		if err := c.Transition(StateInterrupted); err != nil {
			return false
		}
	} else {
		c.mu.Unlock()
	}

	if c.GetState() == StateInterrupted {
		// Wait for resume or abort signal
		select {
		case msg := <-c.resumeChan:
			if err := c.Transition(StateResuming); err != nil {
				return false
			}
			if c.onMessage != nil {
				c.onMessage("Supervisor", msg)
			}
			if err := c.Transition(StateRunning); err != nil {
				return false
			}
			return true
		case <-c.abortChan:
			_ = c.Transition(StateCancelled)
			return false
		}
	}
	return true
}
