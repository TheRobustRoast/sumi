package runner

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// OutputMsg carries a single stdout/stderr line from a running task.
type OutputMsg struct{ Line string }

// DoneMsg signals a task has finished, carrying any error.
type DoneMsg struct{ Err error }

// Stream holds the channels for an active task. Use Func or Cmd to create one,
// then call Next() in Update to read lines one at a time.
type Stream struct {
	lines chan string
	done  chan error
}

// Next returns a tea.Cmd that reads the next line from the stream.
// Returns OutputMsg while lines are available, then DoneMsg when done.
func (s *Stream) Next() tea.Cmd {
	return func() tea.Msg {
		line, ok := <-s.lines
		if !ok {
			return DoneMsg{Err: <-s.done}
		}
		return OutputMsg{Line: line}
	}
}

// Func creates a Stream from a user-provided function. fn receives a "send"
// callback to write output lines. Call RunCmd inside fn to run subprocesses
// and stream their output. The stream ends when fn returns.
func Func(fn func(send func(string)) error) (*Stream, tea.Cmd) {
	s := &Stream{
		lines: make(chan string, 512),
		done:  make(chan error, 1),
	}
	go func() {
		defer close(s.lines)
		s.done <- fn(func(line string) { s.lines <- line })
	}()
	return s, s.Next()
}

// Cmd starts a single external command as a stream.
func Cmd(args ...string) (*Stream, tea.Cmd) {
	return Func(func(send func(string)) error {
		return RunCmd(send, args...)
	})
}

// RunCmd runs a command and streams its combined stdout+stderr to send.
// Use this inside Func callbacks.
func RunCmd(send func(string), args ...string) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close()
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		send(scanner.Text())
	}
	pr.Close()
	return cmd.Wait()
}

// WriteAsSudo writes content to path using "sudo tee". Any error output is
// forwarded to send.
func WriteAsSudo(send func(string), path, content string) error {
	cmd := exec.Command("sudo", "tee", path)
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		send(string(out))
	}
	return err
}

// RunCmdDir is like RunCmd but runs the command in the given directory.
func RunCmdDir(send func(string), dir string, args ...string) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Dir = dir
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close()
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		send(scanner.Text())
	}
	pr.Close()
	return cmd.Wait()
}

// RunCmdWithStdin is like RunCmd but pipes the given string to the command's stdin.
// Used for password piping (e.g., LUKS cryptsetup) where the password must not touch disk.
func RunCmdWithStdin(send func(string), stdin string, args ...string) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return err
	}
	pw.Close()
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		send(scanner.Text())
	}
	pr.Close()
	return cmd.Wait()
}

// ReadFile reads a file, falling back to "sudo cat" if permission is denied.
func ReadFile(path string) (string, error) {
	if data, err := os.ReadFile(path); err == nil {
		return string(data), nil
	}
	out, err := exec.Command("sudo", "cat", path).Output()
	return string(out), err
}
