I'm learning Go, so the code may be rough.

To use this library:

Server Side:

	crier := crier.NewCrier()
	http.Handle("/events/", crier)
	crier.Broadcast(fmt.Sprintf("data: Message: %d - the time is %v\n\n", i, time.Now()))

Client Side:

	<script type="text/javascript">
        var source = new EventSource('/events/');
        source.onmessage = function(e) {
            document.body.innerHTML += e.data + '<br>';
        };
    </script>


Much love taken from https://github.com/kljensen/golang-html5-sse-example/blob/master/server.go and https://godoc.org/github.com/donovanhide/eventsource