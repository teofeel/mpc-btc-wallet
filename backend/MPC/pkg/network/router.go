package network

import (
	"github.com/bnb-chain/tss-lib/v2/tss"
	"log"
	"sync"
)

type Router interface {
	Register(id string, party tss.Party, errCb func(*tss.Error))
	Unregister(id string)
	GetParty(id string) tss.Party
	Stop()
	Route(msg tss.Message, from *tss.PartyID)
}

type MessageRouter struct {
	mu      sync.RWMutex
	parties map[string]tss.Party
	inbox   map[string]chan wireMsg
}

type wireMsg struct {
	data        []byte
	from        *tss.PartyID
	isBroadcast bool
}

func NewMemoryRouter() *MessageRouter {
	return &MessageRouter{
		parties: make(map[string]tss.Party),
		inbox:   make(map[string]chan wireMsg),
	}
}

func processInbox(id string, party tss.Party, ch <-chan wireMsg, errCb func(*tss.Error)) {
	for m := range ch {
		ok, err := party.UpdateFromBytes(m.data, m.from, m.isBroadcast)

		if !ok {
			log.Printf("Party %s failed to register: %v\n", id, err)

		}

		if (!ok || err != nil) && errCb != nil {
			errCb(err)
		}
	}
}

func (r *MessageRouter) processMessage(wireData []byte, routing *tss.MessageRouting, destinations []*tss.PartyID, from *tss.PartyID) {
	for id, ch := range r.inbox {
		if id == from.Id {
			continue
		}

		if destinations == nil || containsPartyId(destinations, id) {
			select {
			case ch <- wireMsg{wireData, routing.From, routing.IsBroadcast}:
			default:
				log.Printf("Party %s failed to route message: %v\n", id, routing)
			}
		}
	}
}

func containsPartyId(list []*tss.PartyID, id string) bool {
	for _, pid := range list {
		if pid.Id == id {
			return true
		}
	}

	return false
}

func (r *MessageRouter) Register(id string, party tss.Party, errCb func(*tss.Error)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.parties[id] = party

	ch := make(chan wireMsg, 100)

	r.inbox[id] = ch

	go processInbox(id, party, ch, errCb)
}

func (r *MessageRouter) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.parties, id)

	if ch, exists := r.inbox[id]; exists {
		close(ch)
		delete(r.inbox, id)
	}
}

func (r *MessageRouter) GetParty(id string) tss.Party {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.parties[id]
}

func (r *MessageRouter) Stop() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, ch := range r.inbox {
		close(ch)
		delete(r.inbox, id)
	}
}

func (r *MessageRouter) Route(msg tss.Message, from *tss.PartyID) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wireData, routing, err := msg.WireBytes()
	if err != nil {
		return
	}

	r.processMessage(wireData, routing, msg.GetTo(), from)
}
