package report

import (
	"fmt"
	"sync"
	"time"
)

const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorBlue  = "\033[34m"
	ColorCyan  = "\033[36m"
)

type LiveLogger struct {
	spinnerDone chan struct{}
	mu          sync.Mutex
}

func NewLiveLogger() *LiveLogger {
	return &LiveLogger{}
}

func (l *LiveLogger) StartSpinner(message string) {
	l.spinnerDone = make(chan struct{})
	frames := []string{
		"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f",
	}

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0

		for {
			select {
			case <-l.spinnerDone:
				fmt.Println("\r\033[K")
				return
			case <-ticker.C:
				l.mu.Lock()
				fmt.Printf("\r%s%s%s %s", ColorCyan, frames[i%len(frames)], ColorReset, message)
				l.mu.Unlock()
				i++
			}

		}
	}()
}

func (l *LiveLogger) StopSpinner() {
	if l.spinnerDone != nil {
		close(l.spinnerDone)
	}
}

func (l *LiveLogger) LogPrint(statusColor, prefix, text string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Printf("\r\033[K%s[%s]%s %s\n", statusColor, prefix, ColorReset, text)
}
