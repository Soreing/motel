package motel

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type SpanCollector struct {
	exporters []sdktrace.SpanExporter
	resource  *resource.Resource

	batched bool
	batchCh chan *Span
	batchWg *sync.WaitGroup
}

// Creates new span collector
func NewSpanCollector(
	exporters []sdktrace.SpanExporter,
	resource *resource.Resource,
	batchTime time.Duration,
	batchLimit int,
) *SpanCollector {
	sc := &SpanCollector{
		exporters: exporters,
		resource:  resource,
		batched:   batchTime > 0,
		batchWg:   &sync.WaitGroup{},
	}

	if batchTime >= 0 {
		sc.batchCh = make(chan *Span)
		sc.batchWg.Add(1)
		go sc.batcher(batchTime, batchLimit)
	}

	return sc
}

func (sc *SpanCollector) batcher(dur time.Duration, limit int) {
	buffer := make([]sdktrace.ReadOnlySpan, limit)
	timer := time.NewTimer(dur)
	count := 0

	var sp *Span
	ctx := context.TODO()
	for active := true; active; {
		select {
		case sp, active = <-sc.batchCh:
			buffer[count] = sp
			count++
			if count == limit || (!active && count > 0) {
				for _, e := range sc.exporters {
					e.ExportSpans(ctx, buffer[:])
				}
				count = 0
				timer.Reset(dur)
			}
		case <-timer.C:
			if count > 0 {
				for _, e := range sc.exporters {
					e.ExportSpans(ctx, buffer[:])
				}
				count = 0
			}
			timer.Reset(dur)
		}
	}

	sc.batchWg.Done()
}

func (sc *SpanCollector) GetResource() *resource.Resource {
	return sc.resource
}

func (sc *SpanCollector) Feed(sp Span) {
	if sc.batched {
		sc.batchCh <- &sp
	} else {
		for _, e := range sc.exporters {
			e.ExportSpans(context.TODO(), []sdktrace.ReadOnlySpan{&sp})
		}
	}
}

func (sc *SpanCollector) Close() {
	if sc.batched {
		close(sc.batchCh)
		sc.batchWg.Wait()
	}
	for _, e := range sc.exporters {
		e.Shutdown(context.TODO())
	}
}

