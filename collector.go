package motel

import (
	"context"
	"sync"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type SpanCollector interface {
	Feed(sp sdktrace.ReadOnlySpan)
	Close()
}

type spanCollector struct {
	exporters []sdktrace.SpanExporter

	batched bool
	batchCh chan sdktrace.ReadOnlySpan
	batchWg *sync.WaitGroup
}

// Creates new span collector
func NewSpanCollector(
	exporters []sdktrace.SpanExporter,
	batchTime time.Duration,
	batchLimit int,
) SpanCollector {
	sc := &spanCollector{
		exporters: exporters,
		batched:   batchTime > 0,
		batchWg:   &sync.WaitGroup{},
	}

	if batchTime >= 0 {
		sc.batchWg.Add(1)
		defer sc.batchWg.Done()
		sc.batchCh = make(chan sdktrace.ReadOnlySpan)
		go sc.batcher(batchTime, batchLimit)
	}

	return sc
}

func (sc *spanCollector) batcher(dur time.Duration, limit int) {
	buffer := make([]sdktrace.ReadOnlySpan, limit)
	timer := time.NewTimer(dur)
	count := 0

	ctx := context.TODO()
	var sp sdktrace.ReadOnlySpan
	for active := true; active; {
		select {
		case sp, active = <-sc.batchCh:
			if active {
				buffer[count] = sp
				count++
			}
			if count == limit || (!active && count > 0) {
				for _, e := range sc.exporters {
					e.ExportSpans(ctx, buffer[:count])
				}
				count = 0
				timer.Reset(dur)
			}
		case <-timer.C:
			if count > 0 {
				for _, e := range sc.exporters {
					e.ExportSpans(ctx, buffer[:count])
				}
				count = 0
			}
			timer.Reset(dur)
		}
	}

	timer.Stop()
	<-timer.C
}

func (sc *spanCollector) Feed(sp sdktrace.ReadOnlySpan) {
	if sc.batched {
		sc.batchCh <- sp
	} else {
		for _, e := range sc.exporters {
			e.ExportSpans(context.TODO(), []sdktrace.ReadOnlySpan{sp})
		}
	}
}

func (sc *spanCollector) Close() {
	if sc.batched {
		close(sc.batchCh)
		sc.batchWg.Wait()
	}
	for _, e := range sc.exporters {
		e.Shutdown(context.TODO())
	}
}
