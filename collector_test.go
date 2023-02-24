package motel

import (
	"context"
	"sync"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type mockEporter struct {
	feed chan []sdktrace.ReadOnlySpan
}

func (e *mockEporter) ExportSpans(
	ctx context.Context,
	spans []sdktrace.ReadOnlySpan,
) (err error) {
	e.feed <- spans
	return nil
}

func (e *mockEporter) Shutdown(
	ctx context.Context,
) (err error) {
	return nil
}

var _ sdktrace.SpanExporter = (*mockEporter)(nil)

func Test_NewSpanCollectorBatched(t *testing.T) {
	ex := &mockEporter{}
	sc := NewSpanCollector(
		[]sdktrace.SpanExporter{ex},
		time.Second,
		5,
	).(*spanCollector)

	if sc.batched == false {
		t.Errorf("span collector expected to be batched")
	}
	if sc.batchCh == nil {
		t.Errorf("batch channel expected to be created")
	}
	if len(sc.exporters) != 1 {
		t.Errorf("span collector expected to have 1 exporter")
	}
}

func Test_NewSpanCollectorUnbatched(t *testing.T) {
	ex := &mockEporter{}
	sc := NewSpanCollector(
		[]sdktrace.SpanExporter{ex},
		0, 0,
	).(*spanCollector)

	if sc.batched == true {
		t.Errorf("span collector not expected to be batched")
	}
	if sc.batchCh != nil {
		t.Errorf("batch channel not expected to be created")
	}
	if len(sc.exporters) != 1 {
		t.Errorf("span collector expected to have 1 exporter")
	}
}

func Test_FeedBatched(t *testing.T) {
	ch := make(chan []sdktrace.ReadOnlySpan)
	ex := &mockEporter{feed: ch}

	spans := 12
	batchSize := 5
	bursts := 3

	sc := NewSpanCollector(
		[]sdktrace.SpanExporter{ex},
		time.Millisecond*10,
		batchSize,
	).(*spanCollector)

	brstCnt := 0
	spanCnt := 0

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		tm := time.NewTimer(time.Millisecond * 50)
		for active := true; active; {
			select {
			case <-tm.C:
				t.Errorf("getting spans timed out")
				active = false
			case spns := <-ch:
				brstCnt++
				spanCnt += len(spns)

				if spanCnt == spans {
					active = false
				}
			}
		}
	}()

	sp := CreateSpan(
		"", trace.SpanKindClient, nil,
		[16]byte{}, [8]byte{}, [8]byte{}, 0x01,
		true, time.Now(), time.Now(),
	)

	for i := 0; i < spans; i++ {
		sc.Feed(sp)
	}

	wg.Wait()

	if brstCnt != bursts {
		t.Errorf("bursts expected to be %d but got %d", bursts, brstCnt)
	}
	if spans != spanCnt {
		t.Errorf("spans expected to be %d but got %d", spans, spanCnt)
	}
}

func Test_FeedUnbatched(t *testing.T) {
	ch := make(chan []sdktrace.ReadOnlySpan)
	ex := &mockEporter{feed: ch}

	spans := 12

	sc := NewSpanCollector(
		[]sdktrace.SpanExporter{ex},
		0, 0,
	).(*spanCollector)

	brstCnt := 0
	spanCnt := 0

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		tm := time.NewTimer(time.Millisecond * 50)
		for active := true; active; {
			select {
			case <-tm.C:
				t.Errorf("getting spans timed out")
				active = false
			case spns := <-ch:
				brstCnt++
				spanCnt += len(spns)

				if spanCnt == spans {
					active = false
				}
			}
		}
	}()

	sp := CreateSpan(
		"", trace.SpanKindClient, nil,
		[16]byte{}, [8]byte{}, [8]byte{}, 0x01,
		true, time.Now(), time.Now(),
	)

	for i := 0; i < spans; i++ {
		sc.Feed(sp)
	}

	wg.Wait()

	if brstCnt != spans {
		t.Errorf("bursts expected to be %d but got %d", spans, brstCnt)
	}
	if spanCnt != spans {
		t.Errorf("spans expected to be %d but got %d", spans, spanCnt)
	}
}
