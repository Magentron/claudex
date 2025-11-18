package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	colorReset  = "\x1b[0m"
	colorCyan   = "\x1b[36m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
)

// SetupPatterns configures all the pattern matching rules
func SetupPatterns(interceptor *Interceptor) error {
	// INPUT RULES - Only checked when user presses ENTER

	// Rule 1: Append custom text to "hello" pattern in user input
	err := interceptor.AddInputRule(`(?i)hello`, func(input string, writer io.Writer) bool {
		// Simply append the custom text (no clearing needed)
		appendedText := ", this is a custom text"

		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[CLAUDEX] Appending: '%s'\n", appendedText)
		}

		writer.Write([]byte(appendedText))

		// Return false to let the original ENTER go through
		return false
	})
	if err != nil {
		return err
	}

	// Rule 2: Capture any substring that would autocomplete to /BMad:agents:dev
	// This matches: /d, /de, /dev, /b, /bm, /bmad, /a, /ag, /agents, etc.
	err = interceptor.AddInputRule(`(?i)^/(d|de|dev|b|bm|bma|bmad|a|ag|age|agen|agent|agents)$`, func(input string, writer io.Writer) bool {
		// Append custom text when this pattern is detected
		appendedText := " - BMad agents development mode activated"

		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[CLAUDEX] BMad command pattern '%s' matched, appending: '%s'\n", input, appendedText)
		}

		writer.Write([]byte(appendedText))

		// Return false to let the original ENTER go through
		return false
	})
	if err != nil {
		return err
	}

	// Rule 3: Replace "goodbye" with test message in user input
	err = interceptor.AddInputRule(`(?i)goodbye`, func(input string, writer io.Writer) bool {
		customMessage := fmt.Sprintf("\n%s[Claudex]%s %sIntercepted \"goodbye\" - sending different message to Claude...%s\n",
			colorYellow, colorReset, colorGreen, colorReset)
		// Write to stderr for user notification
		os.Stderr.WriteString(customMessage)

		// Send the replacement message to Claude via PTY
		replacementMessage := "the test we are doing is working\r"
		writer.Write([]byte(replacementMessage))

		return true // Block original, we sent replacement
	})
	if err != nil {
		return err
	}

	// OUTPUT RULES - Checked continuously on Claude's output

	// PROGRAMMATIC ENTER TEST: Simple test to understand how to submit programmatically
	err = interceptor.AddOutputRule(`TEST-TRIGGER`, func(input string, writer io.Writer) bool {
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "\n========== PROGRAMMATIC ENTER TEST START ==========\n")
			fmt.Fprintf(interceptor.GetLogFile(), "[TEST] Detected TEST-TRIGGER in output\n")
		}

		ptyWriter := interceptor.GetPtyWriter()
		if ptyWriter == nil {
			if interceptor.GetLogFile() != nil {
				fmt.Fprintf(interceptor.GetLogFile(), "[TEST ERROR] PTY writer not available!\n")
			}
			return false
		}

		// Type a simple test message
		testMessage := "this is a programmatic test"
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[TEST] Typing: '%s'\n", testMessage)
		}
		ptyWriter.Write([]byte(testMessage))

		// Try different ways to send Enter
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[TEST] Attempting Enter submission methods:\n")
		}

		// Method 1: Just byte 13
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[TEST] Method 1: []byte{13}\n")
		}
		n, _ := ptyWriter.Write([]byte{13})
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[TEST] Method 1 result: wrote %d bytes\n", n)
		}

		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "========== PROGRAMMATIC ENTER TEST END ==========\n\n")
		}

		return false
	})
	if err != nil {
		return err
	}

	// Rule 4: Detect when /BMad:agents:dev is running and interrupt it
	err = interceptor.AddOutputRule(`/BMad:agents:dev is running`, func(input string, writer io.Writer) bool {
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "\n========== BMAD OUTPUT INTERCEPTION START ==========\n")
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Detected '/BMad:agents:dev is running' in output\n")
		}

		ptyWriter := interceptor.GetPtyWriter()
		if ptyWriter == nil {
			if interceptor.GetLogFile() != nil {
				fmt.Fprintf(interceptor.GetLogFile(), "[BMAD ERROR] PTY writer not available!\n")
			}
			return false
		}

		// Send ESC to interrupt Claude
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Sending ESC to interrupt\n")
		}
		ptyWriter.Write([]byte{0x1B})

		// Wait for interrupt to take effect (needs more time when Claude is actively processing)
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Waiting 500ms for interrupt to complete...\n")
		}
		time.Sleep(500 * time.Millisecond)

		// Type the modified command (all at once, like TEST-TRIGGER which worked!)
		modifiedCommand := "/BMad:agents:dev - BMad agents development mode activated"
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Typing: '%s'\n", modifiedCommand)
		}
		ptyWriter.Write([]byte(modifiedCommand))

		// Wait for the text to be processed before sending Enter
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Waiting 1000ms before sending Enter...\n")
		}
		time.Sleep(1000 * time.Millisecond)

		// Send Enter - try multiple approaches
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Sending Enter attempts...\n")
		}

		// Attempt 1: Single \r (byte 13)
		n1, _ := ptyWriter.Write([]byte{13})
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Attempt 1: sent \\r (byte 13), wrote %d bytes\n", n1)
		}

		time.Sleep(100 * time.Millisecond)

		// Attempt 2: Send \r again
		n2, _ := ptyWriter.Write([]byte{13})
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Attempt 2: sent \\r (byte 13) again, wrote %d bytes\n", n2)
		}

		time.Sleep(100 * time.Millisecond)

		// Attempt 3: Try \n (byte 10)
		n3, _ := ptyWriter.Write([]byte{10})
		if interceptor.GetLogFile() != nil {
			fmt.Fprintf(interceptor.GetLogFile(), "[BMAD] Attempt 3: sent \\n (byte 10), wrote %d bytes\n", n3)
			fmt.Fprintf(interceptor.GetLogFile(), "========== BMAD OUTPUT INTERCEPTION END ==========\n\n")
		}

		return false
	})
	if err != nil {
		return err
	}

	// Rule 5: Detect "hello world" in output
	err = interceptor.AddOutputRule(`(?i)hello world`, func(input string, writer io.Writer) bool {
		customMessage := fmt.Sprintf("\n%s[Claudex]%s %s🎉 Hello World detected in OUTPUT!%s\n",
			colorYellow, colorReset, colorCyan, colorReset)
		os.Stderr.WriteString(customMessage)
		return false // Don't block, just notify
	})
	if err != nil {
		return err
	}

	return nil
}
