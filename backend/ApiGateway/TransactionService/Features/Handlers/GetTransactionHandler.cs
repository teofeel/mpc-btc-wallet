using MediatR;
using TransactionService.DTOs;
using TransactionService.Features.Queries;
using TransactionService.Services.Interfaces;

namespace TransactionService.Features.Handlers
{
    public class GetTransactionHandler : IRequestHandler<GetTransactionQuery, TransactionResponseDTO?>
    {
        private readonly ITransactionService _transactionService;
        private readonly ILogger<GetTransactionHandler> _logger;

        public GetTransactionHandler(ITransactionService transactionService,ILogger<GetTransactionHandler> logger)
        {
            _transactionService = transactionService;
            _logger = logger;
        }

        public async Task<TransactionResponseDTO?> Handle(GetTransactionQuery request, CancellationToken cancellationToken)
        {
            _logger.LogInformation("[Transaction Service] Handling get transaction for ID: {TransactionId}", request.Id);

            return await _transactionService.GetTransactionByIdAsync(request.Id);
        }
    }
}
