package crier

type history struct {
	size   int // size of the history to be kept
	history      map[int]*event
	lastId int
}

type event struct {
	id      int
	message string
	command string    // publish, broadcast, whisper
	target  *listener // non-nil when command is whisper
	group   string    // non-nil when command is publish
}

func (h *history) addBroadcast(message string) {
	e := &event{
		id: 		h.getNextEventId(),
		command:	"broadcast",
		message:	message,
	}
	h.history[e.id] = e
}

func (h *history) addPublish(message, group string) {
	e := &event{
		id: 		h.getNextEventId(),
		command:	"publish",
		group:		group,
		message:	message,
	}
	h.history[e.id] = e
}

func (h *history) addWhisper(message string, l *listener) {
	e := &event{
		id: 		h.getNextEventId(),
		command:	"whisper",
		target:		l,
		message:	message,
	}
	h.history[e.id] = e
}

func (h *history) getNextEventId() int {
	h.lastId += 1
	return h.lastId
}

