using MediatR;
using TransactionService.DTOs;

namespace TransactionService.Features.Queries
{
    public record GetTransactionsByWalletIdQuery(Guid WalletId, int Page = 1, int PageSize = 10) : IRequest<List<TransactionResponseDTO>>;
}
