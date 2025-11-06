package chess

import "sync"

type eventSubscriber struct {
	ch  chan interface{}
	ack chan struct{}
}

func newEventSubscriber(handler func(any)) *eventSubscriber {
	sub := &eventSubscriber{
		ch:  make(chan interface{}),
		ack: make(chan struct{}),
	}

	go func() {
		for data := range sub.ch {
			handler(data)
			sub.ack <- struct{}{}
		}
	}()

	return sub
}

type eventHub struct {
	mu        sync.RWMutex
	listeners map[string][]*eventSubscriber
}

func newEventHub() *eventHub {
	return &eventHub{
		listeners: make(map[string][]*eventSubscriber),
	}
}

func (h *eventHub) emit(e string, dta any) {
	if h == nil || e == "" {
		return
	}

	h.mu.RLock()
	subs := append([]*eventSubscriber(nil), h.listeners[e]...)
	h.mu.RUnlock()

	for _, sub := range subs {
		sub.ch <- dta
		<-sub.ack
	}
}

func (h *eventHub) on(e string, hndlr func(any)) {
	if h == nil || e == "" || hndlr == nil {
		return
	}

	sub := newEventSubscriber(hndlr)
	h.mu.Lock()
	h.listeners[e] = append(h.listeners[e], sub)
	h.mu.Unlock()
}

type KingThreatEvent struct {
	AttackingSquare *Square
	KingSquare      *Square
}

type MoveEvent struct {
	Algebraic              string
	CapturedPiece          *Piece
	Castle                 bool
	EnPassant              bool
	Piece                  *Piece
	PostSquare             *Square
	PrevSquare             *Square
	Promotion              bool
	RookSource             *Square
	RookDestination        *Square
	EnPassantCaptureSquare *Square
	hashCode               string
	prevMoveCount          int
	simulate               bool
	undone                 bool
}
