package workforce

import (
	"sync"
	"testing"
	"time"
)

func TestInterruptionFlow(t *testing.T) {
	var states []RunState
	var messages []string
	var mu sync.Mutex

	coordinator := NewCoordinator(
		func(s RunState) {
			mu.Lock()
			states = append(states, s)
			mu.Unlock()
		},
		func(sender, content string) {
			mu.Lock()
			messages = append(messages, content)
			mu.Unlock()
		},
	)

	// Start run
	err := coordinator.Transition(StateRunning)
	if err != nil {
		t.Fatalf("Failed to transition to running: %v", err)
	}

	stepChan := make(chan int)
	doneChan := make(chan bool)

	// Simulating the workforce loop
	go func() {
		for i := 1; i <= 3; i++ {
			if !coordinator.CheckStepBoundary() {
				doneChan <- false
				return
			}
			stepChan <- i
			// Simulate some work time
			time.Sleep(10 * time.Millisecond)
		}
		doneChan <- true
	}()

	// 1. First step should execute immediately
	select {
	case step := <-stepChan:
		if step != 1 {
			t.Errorf("Expected step 1, got %d", step)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for step 1")
	}

	// Wait a tiny bit for the goroutine to finish step 1 and call check boundary for step 2
	time.Sleep(5 * time.Millisecond)

	// 2. Call Interrupt
	err = coordinator.Interrupt()
	if err != nil {
		t.Fatalf("Failed to interrupt: %v", err)
	}

	// Poll until state becomes INTERRUPTED
	ok := false
	for i := 0; i < 50; i++ {
		if coordinator.GetState() == StateInterrupted {
			ok = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ok {
		t.Errorf("Expected state to be INTERRUPTED, got %s", coordinator.GetState())
	}

	// 3. Resume with message
	err = coordinator.Resume("Proceed with coder")
	if err != nil {
		t.Fatalf("Failed to resume: %v", err)
	}

	// Step 2 should now execute
	select {
	case step := <-stepChan:
		if step != 2 {
			t.Errorf("Expected step 2, got %d", step)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for step 2")
	}

	mu.Lock()
	if len(messages) != 1 || messages[0] != "Proceed with coder" {
		t.Errorf("Expected message 'Proceed with coder', got %v", messages)
	}
	mu.Unlock()

	// Wait a tiny bit for step 2 to complete and loop to run CheckStepBoundary for step 3
	time.Sleep(5 * time.Millisecond)

	// 4. Test abort/cancel flow
	err = coordinator.Interrupt()
	if err != nil {
		t.Fatalf("Failed to interrupt: %v", err)
	}

	// Poll until state becomes INTERRUPTED
	ok = false
	for i := 0; i < 50; i++ {
		if coordinator.GetState() == StateInterrupted {
			ok = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ok {
		t.Fatalf("Expected state to be INTERRUPTED for abort check, got %s", coordinator.GetState())
	}

	err = coordinator.Abort()
	if err != nil {
		t.Fatalf("Failed to abort: %v", err)
	}

	select {
	case success := <-doneChan:
		if success {
			t.Error("Expected execution to be cancelled, but it succeeded")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for loop termination")
	}

	mu.Lock()
	hasCancelledState := false
	for _, s := range states {
		if s == StateCancelled {
			hasCancelledState = true
			break
		}
	}
	mu.Unlock()

	if !hasCancelledState {
		t.Error("Expected coordinator to reach CANCELLED state")
	}
}
