using MediatR;
using TransactionService.Features.Commands;
using TransactionService.Services.Interfaces;

namespace TransactionService.Features.Handlers
{
    public class UpdateTransactionConfirmationsHandler : IRequestHandler<UpdateTransactionConfirmationsCommand>
    {
        private readonly ITransactionService _transactionService;

        public UpdateTransactionConfirmationsHandler(ITransactionService transactionService)
        {
            _transactionService = transactionService;
        }

        public async Task Handle(UpdateTransactionConfirmationsCommand request, CancellationToken cancellationToken)
        {
            await _transactionService.UpdateTransactionConfirmationsAsync(request.TxHash);
        }
    }
}
