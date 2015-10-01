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

TODO:

*  Ability to resend previous event using the history
*  Ability to send events to a specific client or group
*  Auto-resend previous event if not successfully sent the first time
*  Write test cases. I've **never** done this before, so this'll be fun
*  Figure out how to allow importee code to have access to a list of clients