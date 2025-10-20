package chess

type eventEmitter struct {
	listeners map[string][]func(interface{})
}

func newEventEmitter() eventEmitter {
	return eventEmitter{
		listeners: map[string][]func(interface{}){},
	}
}

func (e *eventEmitter) on(event string, handler func(interface{})) {
	if event == "" || handler == nil {
		return
	}
	e.listeners[event] = append(e.listeners[event], handler)
}

func (e *eventEmitter) emit(event string, data interface{}) {
	if event == "" {
		return
	}
	handlers := e.listeners[event]
	for _, handler := range handlers {
		handler(data)
	}
}