using TransactionService.DTOs;
using TransactionService.Mappers.Interfaces;
using TransactionService.Models;

namespace TransactionService.Mappers
{
    public class TransactionResponseMapper : IMapper<Transaction, TransactionResponseDTO>
    {

        public Transaction ToEntity(TransactionResponseDTO data)
        {
            throw new NotImplementedException("ToEntity is not implemented yet.");
        }

        public TransactionResponseDTO ToDto(Transaction data)
        {
            if (data == null) return null!;

            return new TransactionResponseDTO
            {
                Id = data.Id,
                ReceivingAddress = data.ReceivingAddress,
                AmountInBtc = data.Amount,
                FeeInBtc = data.Fee,
                TxHash = data.TxHash,
                Confirmations = data.Confirmations,
                Status = data.Status.ToString(),
                CreatedAt = data.CreatedAt,
                ConfirmedAt = data.ConfirmedAt
            };
        }
    }
}
