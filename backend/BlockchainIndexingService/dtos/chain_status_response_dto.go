package dtos

type ChainStatusResponse struct {
	Chain                string
	NodeHeight           int64
	IndexedHeight        int64
	SyncPercentage       float64
	BestBlockHash        string
	VerificationProgress float64
}
