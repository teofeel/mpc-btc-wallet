package dkg

import pb "mpc/gen/mpc"

type Job struct {
	JobID          string
	WalletID       string
	Status         pb.GetKeyGenStatusResponse_Status
	PublicKeyHex   string
	ChainCodeHex   string
	XPub           string
	BitcoinAddress string
	ErrMsg         string
}
