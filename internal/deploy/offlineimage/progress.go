package deployofflineimage

import (
	"io"
	"sync/atomic"

	"github.com/containers/image/v5/types"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type progressPrinter struct {
	label  string
	last   int32
	closed atomic.Bool
	paused bool
}

const progressStep = 5

func newProgressPrinter(label string) *progressPrinter {
	paused := util.SpinPause()
	util.PrintProgress(label, 0, false)
	return &progressPrinter{label: label, last: -1, paused: paused}
}

func (p *progressPrinter) Update(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	prev := atomic.LoadInt32(&p.last)
	if percent != 100 && prev >= 0 && percent-int(prev) < progressStep {
		return
	}
	if atomic.LoadInt32(&p.last) == int32(percent) {
		return
	}
	atomic.StoreInt32(&p.last, int32(percent))
	util.PrintProgress(p.label, percent, false)
}

func (p *progressPrinter) Close() {
	if p == nil || !p.closed.CompareAndSwap(false, true) {
		return
	}
	// Print final 100% with newline (don't call Update to avoid duplicate)
	util.PrintProgress(p.label, 100, true)
	if p.paused {
		util.SpinUnpause()
	}
}

func startProgressTracker(label string, progressCh chan types.ProgressProperties) func() {
	if progressCh == nil {
		return func() {}
	}
	printer := newProgressPrinter(label)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for prop := range progressCh {
			total := prop.Artifact.Size
			if total <= 0 {
				continue
			}
			percent := int(float64(prop.Offset) / float64(total) * 100)
			if prop.Event == types.ProgressEventDone {
				percent = 100
			}
			// Don't print 100% here - let Close() handle it to avoid duplicates
			if percent < 100 {
				printer.Update(percent)
			}
		}
	}()
	return func() {
		<-done
		printer.Close()
	}
}

type progressReader struct {
	reader  io.Reader
	total   int64
	read    int64
	printer *progressPrinter
}

func newProgressReader(r io.Reader, total int64, label string) *progressReader {
	return &progressReader{
		reader:  r,
		total:   total,
		printer: newProgressPrinter(label),
	}
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.reader.Read(b)
	if n > 0 && p.total > 0 {
		p.read += int64(n)
		percent := int(float64(p.read) / float64(p.total) * 100)
		p.printer.Update(percent)
	}
	// Don't call Update(100) on EOF - let Close() handle the final print
	return n, err
}

func (p *progressReader) Close() {
	if p == nil {
		return
	}
	p.printer.Close()
}
