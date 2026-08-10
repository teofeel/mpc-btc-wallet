namespace TransactionService.DTOs
{
    public class TransactionResponseDTO
    {
        public Guid Id { get; set; }
        public string ReceivingAddress { get; set; } = string.Empty;
        public decimal AmountInBtc { get; set; }
        public decimal FeeInBtc { get; set; }
        public string? TxHash { get; set; }
        public int Confirmations { get; set; }
        public string Status { get; set; } = string.Empty;
        public DateTime CreatedAt { get; set; }
        public DateTime? ConfirmedAt { get; set; }
        public string? ErrorMessage { get; set; }

    }
}
