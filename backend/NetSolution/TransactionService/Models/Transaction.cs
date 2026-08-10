using System.Transactions;
using TransactionService.Models.Enums;

namespace TransactionService.Models
{
    public class Transaction
    {
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid WalletId { get; set; }
        public string ReceivingAddress { get; set; } = string.Empty;
        public string? SendingAddress { get; set; }
        public decimal Amount { get; set; }
        public decimal Fee { get; set; }
        public string? TxHash { get; set; }
        public int Confirmations { get; set; } = 0;
        public TxStatus Status { get; set; } = TxStatus.PENDING;
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime? BroadcastedAt { get; set; }
        public DateTime? ConfirmedAt { get; set; }
    }
}
