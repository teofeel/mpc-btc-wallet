using MediatR;
using System.Transactions;
using TransactionService.Models.Enums;

namespace TransactionService.Features.Queries
{
    public record GetTransactionHashesQuery(TxStatus Status) : IRequest<List<string>>;
}
