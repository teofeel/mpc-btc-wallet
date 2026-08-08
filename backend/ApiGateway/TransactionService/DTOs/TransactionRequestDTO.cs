using System.ComponentModel.DataAnnotations;
using TransactionService.Models.Enums;

namespace TransactionService.DTOs
{
    public class TransactionRequestDTO
    {
        [Required(ErrorMessage = "Receiving address is required.")]
        [StringLength(90, MinimumLength = 26, ErrorMessage = "Invalid address length.")]
        public String ReceivingAddress { get; set; }

        [Range(0.00000001, 21000000.0, ErrorMessage = "Amount must be between 0.00000001 and 21,000,000 BTC.")]
        public decimal Amount { get; set; }

        public FeePriority FeePriority { get; set; } = FeePriority.MEDIUM;
    }
}
