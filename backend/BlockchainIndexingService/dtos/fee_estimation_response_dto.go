package dtos

type FeeEstimationResponse struct {
	TargetBlocks int64
	FeeRateSatVB float64
	FeeRateBtcKB float64
}
