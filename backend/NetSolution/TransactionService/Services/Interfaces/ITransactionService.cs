using System.Threading.Tasks;
using TransactionService.DTOs;
using TransactionService.Models;
using TransactionService.Models.Enums;

namespace TransactionService.Services.Interfaces
{
    public interface ITransactionService
    {
        Task<TransactionResponseDTO> MakeTransactionAsync(TransactionRequestDTO request);
        Task<TransactionResponseDTO> GetTransactionByIdAsync(Guid id);
        Task<FeeEstimationResponseDTO> GetFeeEstimationAsync(FeePriority priority = FeePriority.MEDIUM);
        Task<bool> IsAddressValidAsync(string bitcoinAddress);
        Task UpdateTransactionConfirmationsAsync(string txHash);
    }
}
