namespace TransactionService.DTOs
{
    public class FeeEstimationResponseDTO
    {
        public string Priority { get; init; } = string.Empty;
        public int TargetBlocks { get; init; }
        public decimal EstimatedFeeInBtc { get; init; }
        public long EstimatedFeeInSatoshis { get; init; }
    }
}
