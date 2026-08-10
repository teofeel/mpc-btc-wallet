package network

import "github.com/bnb-chain/tss-lib/v2/tss"

type Router interface {
	Register(id string, party tss.Party, errCb func(*tss.Error))
	Unregister(id string)
	GetParty(id string) tss.Party
	Stop()
	Route(msg tss.Message, from *tss.PartyID)
}
