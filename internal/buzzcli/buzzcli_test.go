package buzzcli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSendMessageWithFakeBuzz(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")
	factory := shellFactory(`cat >/dev/null; echo '{"ok":true}'`)

	err := sendMessage(
		context.Background(),
		"chan-1",
		"hello\n\n---\nfestival: X\ntask: -\npath: -\ngate: n/a\n---\n",
		"ws://localhost:3000",
		time.Second,
		fakeLookPath,
		factory,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendMessageRequiresKey(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "")
	err := SendMessage(context.Background(), "c", "body", "")
	if err == nil {
		t.Fatal("expected error without key")
	}
}

func TestSendMessageHonorsEarlierCallerDeadline(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	started := time.Now()

	err := sendMessage(ctx, "chan-1", "body", "", time.Second, fakeLookPath, shellFactory("while :; do :; done"))

	assertPromptContextError(t, err, context.DeadlineExceeded, started)
}

func TestSendMessageHonorsCancellation(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()
	started := time.Now()

	err := sendMessage(ctx, "chan-1", "body", "", time.Second, fakeLookPath, shellFactory("while :; do :; done"))

	assertPromptContextError(t, err, context.Canceled, started)
}

func TestSendMessageReapsTimedOutProcessesWithoutGoroutineGrowth(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")
	before := runtime.NumGoroutine()
	commands := make([]*exec.Cmd, 0, 10)
	factory := func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "sh", "-c", "while :; do :; done")
		commands = append(commands, cmd)
		return cmd
	}

	for range 10 {
		err := sendMessage(context.Background(), "chan-1", "body", "", 10*time.Millisecond, fakeLookPath, factory)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline error, got %v", err)
		}
	}
	for _, cmd := range commands {
		if cmd.ProcessState == nil || !errors.Is(cmd.Process.Kill(), os.ErrProcessDone) {
			t.Fatal("timed-out Buzz process was not reaped")
		}
	}
	runtime.Gosched()
	if after := runtime.NumGoroutine(); after > before+1 {
		t.Fatalf("goroutines grew from %d to %d", before, after)
	}
}

func TestSendMessageWrapsSpawnAndExitFailures(t *testing.T) {
	t.Setenv("BUZZ_PRIVATE_KEY", "test-key")
	tests := []struct {
		name    string
		factory commandFactory
		want    string
	}{
		{name: "spawn", factory: shellFactoryWithBinary("/does/not/exist", ""), want: "no such file"},
		{name: "exit", factory: shellFactory("exit 23"), want: "exit status 23"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendMessage(context.Background(), "chan-1", "body", "", time.Second, fakeLookPath, tt.factory)
			if err == nil || !strings.Contains(err.Error(), "buzz messages send for channel \"chan-1\"") || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func fakeLookPath() (string, error) {
	return "buzz", nil
}

func shellFactory(script string) commandFactory {
	return shellFactoryWithBinary("sh", script)
}

func shellFactoryWithBinary(binary, script string) commandFactory {
	return func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, binary, "-c", script)
	}
}

func assertPromptContextError(t *testing.T, err error, target error, started time.Time) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("expected %v, got %v", target, err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cancellation took %s", elapsed)
	}
	if !strings.Contains(err.Error(), "buzz messages send for channel \"chan-1\"") {
		t.Fatalf("missing command context: %v", err)
	}
}
