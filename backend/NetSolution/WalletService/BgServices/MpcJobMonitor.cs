using MediatR;
using WalletService.Services.Interfaces;

namespace WalletService.BgServices
{
    public class MpcJobMonitor : BackgroundService
    {
        private readonly IMpcCompletionQueue _queue;
        private readonly IServiceScopeFactory _scopeFactory;
        private readonly ILogger<MpcJobMonitor> _logger;

        public MpcJobMonitor(IMpcCompletionQueue queue, IServiceScopeFactory scopeFactory, ILogger<MpcJobMonitor> logger)
        {
            _queue = queue;
            _scopeFactory = scopeFactory;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            await foreach(var notification in _queue.ReadCompletionAsync(cancellationToken))
            {
                try
                {
                    using var scope = _scopeFactory.CreateScope();
                    var mediator = scope.ServiceProvider.GetRequiredService<IMediator>();

                    await mediator.Publish(notification, cancellationToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[Wallet Service] Failed to process MPC completion for Wallet {WalletId}", notification.WalletId);
                }
            }
        }
    }
}
