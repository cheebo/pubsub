package memory

import (
	"context"
	"github.com/cheebo/pubsub"
	"github.com/pkg/errors"
	"sync"
)

type Hub interface {
	NewExchange(name string) chan<- []byte
	NewSubscriber(exchange string, unMarshaler pubsub.UnMarshaller) (pubsub.Subscriber, error)
}

type hub struct {
	sid         uint
	mutex       sync.Mutex
	exchanges   map[string]chan []byte
	subscribers map[string]map[uint]chan []byte
	stop        chan uint
}

type publisher struct {
	marshaller pubsub.Marshaller
	ch         chan<- []byte
	err        chan error
}

type subscriber struct {
	id           uint
	unMarshaller pubsub.UnMarshaller
	stop         chan<- uint
	errors       chan error
	subscribed   <-chan []byte
}

type message struct {
	unmarshaller pubsub.UnMarshaller
	body         []byte
}

func NewHub() Hub {
	hb := &hub{
		mutex:       sync.Mutex{},
		exchanges:   make(map[string]chan []byte),
		subscribers: make(map[string]map[uint]chan []byte),
		stop:        make(chan uint),
	}

	go func(mutex sync.Mutex, stop chan uint) {
		for {
			select {
			case stopId := <-stop:
				mutex.Lock()

				for name, exs := range hb.subscribers {
					for sid, ch := range exs {
						if sid == stopId {
							close(ch)
							delete(hb.subscribers[name], sid)
						}
					}
				}

				mutex.Unlock()
			}
		}
	}(hb.mutex, hb.stop)

	return hb
}

func (h *hub) NewExchange(name string) chan<- []byte {
	if ex, ok := h.exchanges[name]; ok {
		return ex
	}
	ex := make(chan []byte)
	h.exchanges[name] = ex
	stop := make(chan []byte)

	go func(name string, exchange chan []byte, stop chan []byte, mutex sync.Mutex) {
		for {
			select {
			case msg := <-exchange:
				mutex.Lock()
				lst := h.subscribers[name]
				for _, ch := range lst {
					ch <- msg
				}
				mutex.Unlock()
			}
		}
	}(name, ex, stop, h.mutex)

	return ex
}

func (h *hub) NewSubscriber(exchange string, unMarshaller pubsub.UnMarshaller) (pubsub.Subscriber, error) {
	h.sid++
	h.mutex.Lock()
	defer h.mutex.Unlock()

	_, ok := h.exchanges[exchange]
	if !ok {
		return nil, errors.New("exchange not found")
	}

	ch := make(chan []byte)
	if h.subscribers[exchange] == nil {
		h.subscribers[exchange] = make(map[uint]chan []byte)
	}
	h.subscribers[exchange][h.sid] = ch
	sub := newSubscriber(h.sid, ch, h.stop, unMarshaller)

	return sub, nil
}

func NewPublisher(ch chan<- []byte, marshaller pubsub.Marshaller) pubsub.Publisher {
	return &publisher{
		ch:         ch,
		err:        make(chan error),
		marshaller: marshaller,
	}
}

func (p *publisher) Publish(ctx context.Context, key string, msg interface{}) error {
	if p.marshaller == nil {
		return pubsub.ErrorNoMarshaller
	}
	body, err := p.marshaller(msg)
	if err != nil {
		return nil
	}
	p.ch <- body

	return nil
}

func (p *publisher) Errors() <-chan error {
	return p.err
}

func newSubscriber(id uint, subscribe <-chan []byte, stop chan<- uint, unMarshaller pubsub.UnMarshaller) pubsub.Subscriber {
	return &subscriber{
		id:           id,
		stop:         stop,
		errors:       make(chan error, 100),
		subscribed:   subscribe,
		unMarshaller: unMarshaller,
	}
}

func (s *subscriber) Start() <-chan pubsub.Message {
	output := make(chan pubsub.Message)
	go func(s *subscriber, output chan pubsub.Message) {
		defer func() {
			close(output)
		}()
		for {
			select {
			case msg := <-s.subscribed:
				output <- &message{
					unmarshaller: s.unMarshaller,
					body:         msg,
				}
			default:
				// void
			}
		}
	}(s, output)
	return output
}

func (s *subscriber) Errors() <-chan error {
	return s.errors
}

func (s *subscriber) Stop() error {
	s.stop <- s.id
	return nil
}

func (m *message) UnMarshal(msg interface{}) error {
	if m.unmarshaller == nil {
		return pubsub.ErrorNoUnMarshaller
	}
	return m.unmarshaller(m.body, msg)
}

func (m *message) Done() error {
	return nil
}
