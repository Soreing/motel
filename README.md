# Manual Open Telemetry
Motel is a simple library that lets you create Open Telemetry ReadOnlySpan objects with a constructor function instead of a tracer. The library batches and sends spans to one or more exporters.

## Usage

Create a Motel Span Collector that will accept Open Telemetry ReadOnlySpans. You need to specify a list of exporters and a batch interval with batch limit. Spans will be sent to exporters periodically specified by the batch interval or when the number of batched spans reach the limit. Setting the batch interval to 0 will disable batching and skip creating the resources for it.
```golang
coll := motel.NewSpanCollector(
    []trace.SpanExporter{},
    time.Second * 5,
    10,
)
```

You can submit spans to the collector with the Feed method.
```golang
coll.Feed([]trace.ReadOnlySpan{/* spans */})
```

To stop using the collector, use the Close method. The collector will close the channel used for batching and wait for the goroutine to stop before calling Shutdown on each exporter.
```golang
coll.Close(context.TODO())
```

## Creating Spans

You can create Spans that implement ReadOnlySpan with a constructor function. You will need to provide required information from a traceparent header and additional details about the span, such name, as id, uccess, start and end time. 
```golang
// W3C traceparent Header (ver-tid-pid-flg)
trc := "00-6aa68cc5f3ec62cc7311d38af7fb4176-b5e0e5ec613d07ab-01"
// Span Id
sId := "a4dfd4db502cf69a"

tid := [16]byte{ /* Convert "6aa68cc5f3ec62cc7311d38af7fb4176" */}
pid := [16]byte{ /* Convert "b5e0e5ec613d07ab" */}
sid := [16]byte{ /* Convert "a4dfd4db502cf69a" */}
flg := 0x01
success := true
start, end := time.Now(), time.Now().Add(time.Second)

// Create span with details
span := motel.CreateSpan(
	"Span Name", trace.SpanKindServer,
	resource.Default{},
	tid, pid, sid, flg, success
	startTime, endTime,
) 
```

