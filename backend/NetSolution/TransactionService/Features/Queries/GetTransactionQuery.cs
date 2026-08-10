using MediatR;
using TransactionService.DTOs;

namespace TransactionService.Features.Queries
{
    public record GetTransactionQuery(Guid Id) : IRequest<TransactionResponseDTO>;
}
