using MediatR;
using TransactionService.DB;
using TransactionService.Features.Commands;
using TransactionService.Features.Queries;
using TransactionService.Models.Enums;
using TransactionService.Services.Interfaces;

namespace TransactionService.BgServices
{
    public class TransactionConfirmationBgService : BackgroundService
    {
        private readonly IServiceScopeFactory _scopeFactory;
        private readonly ILogger<TransactionConfirmationBgService> _logger;
        private readonly TimeSpan _pollInterval = TimeSpan.FromSeconds(15);

        public TransactionConfirmationBgService(IServiceScopeFactory scopeFactory, ILogger<TransactionConfirmationBgService> logger)
        {
            _scopeFactory = scopeFactory;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("[Transaction Service] Transaction Confirmation bg service started");

            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    await SyncPendingTransactionsAsync(stoppingToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[Transaction Service] Error occurred while syncing pending transactions");
                }

                await Task.Delay(_pollInterval, stoppingToken);
            }

            _logger.LogInformation("[Transaction Service] Transaction Confirmation bg service stopping");
        }

        private async Task SyncPendingTransactionsAsync(CancellationToken ct)
        {
            using var scope = _scopeFactory.CreateScope();
            var mediator = scope.ServiceProvider.GetRequiredService<ISender>();

            List<string> pendingTxHashes = await mediator.Send(new GetTransactionHashesQuery(TxStatus.PENDING), ct);

            if (pendingTxHashes == null || !pendingTxHashes.Any()) return;

            _logger.LogInformation("[Transaction Service] Found {Count} pending transactions to sync", pendingTxHashes.Count);

            foreach (string txHash in pendingTxHashes)
            {
                try
                {
                    await mediator.Send(new UpdateTransactionConfirmationsCommand(txHash), ct);
                }
                catch(Exception ex)
                {
                    _logger.LogWarning(ex, "[Transaction Service] Failed to update transaction {TxHash}", txHash);
                }
            }
        }
    }

}

