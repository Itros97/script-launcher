package services

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Itros97/script-launcher/internal/domain"
)

// watch reads the view's frames and messages until both pipes close, then records how it ended.
func (service *PreviewService) watch(stdout, stderr io.Reader) {
	var group sync.WaitGroup
	group.Add(2)
	go func() { defer group.Done(); service.readFrames(stdout) }()
	go func() { defer group.Done(); service.readMessages(stderr) }()
	group.Wait()
	service.finish(service.command.Wait())
}

// readFrames swaps each complete frame in as the latest, reusing the previous one's pixels for the next read.
func (service *PreviewService) readFrames(stdout io.Reader) {
	var spare []byte
	for {
		frame, _, err := domain.ReadFrame(stdout, spare)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			service.note(fmt.Sprintf("La vista de la cámara se ha cortado: %v", err))
			_, _ = io.Copy(io.Discard, stdout) // keep the pipe drained so the process can end
			return
		}
		service.mutex.Lock()
		spare = service.latest.RGB
		service.latest = frame
		service.sequence++
		service.mutex.Unlock()
		service.onChange()
	}
}

// readMessages copies what the script writes for the user into the log.
func (service *PreviewService) readMessages(stderr io.Reader) {
	reader := bufio.NewReader(stderr)
	for {
		text, err := reader.ReadString('\n')
		if text != "" {
			service.append(messageLine(cleanLine(text)))
		}
		if err != nil {
			return
		}
	}
}

// finish marks the view as closed, drops its last frame and says how it ended, all at once so a
// view started right after never sees this one's ending.
func (service *PreviewService) finish(waitErr error) {
	service.mutex.Lock()
	ending := "■ Cámara parada"
	if !service.stopping && waitErr != nil {
		ending = fmt.Sprintf("✗ La cámara se ha cerrado (código %d)", service.command.ProcessState.ExitCode())
	}
	service.log.Append(domain.Line{Text: ending, Stream: domain.System})
	service.running, service.latest = false, domain.Frame{}
	service.mutex.Unlock()
	service.onChange()
}

// messageLine marks the script's progress lines (they start with ▸) as normal output and the rest as errors.
func messageLine(text string) domain.Line {
	if strings.HasPrefix(text, "▸") {
		return domain.Line{Text: text, Stream: domain.Stdout}
	}
	return domain.Line{Text: text, Stream: domain.Stderr}
}

// note adds a message of the launcher's own to the log.
func (service *PreviewService) note(text string) {
	service.append(domain.Line{Text: text, Stream: domain.System})
}

func (service *PreviewService) append(line domain.Line) {
	service.mutex.Lock()
	service.log.Append(line)
	service.mutex.Unlock()
	service.onChange()
}
