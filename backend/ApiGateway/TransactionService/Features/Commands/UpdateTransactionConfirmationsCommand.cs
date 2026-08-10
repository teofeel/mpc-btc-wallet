using MediatR;

namespace TransactionService.Features.Commands
{
    public record UpdateTransactionConfirmationsCommand(string TxHash) : IRequest;
}
