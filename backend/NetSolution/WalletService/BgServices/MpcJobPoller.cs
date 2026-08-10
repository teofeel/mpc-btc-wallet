using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using WalletService.DB;
using WalletService.Events;
using WalletService.Models;
using WalletService.Models.Enums;
using WalletService.Services.Interfaces;
using WalletService.Services.MpcService;

namespace WalletService.BgServices
{
    public class MpcJobPoller : BackgroundService
    {
        private static readonly TimeSpan PollInterval = TimeSpan.FromSeconds(5);
        private static readonly TimeSpan PendingStaleAfter = TimeSpan.FromMinutes(5);

        private readonly IServiceScopeFactory _scopeFactory;
        private readonly ILogger<MpcJobPoller> _logger;

        public MpcJobPoller(IServiceScopeFactory scopeFactory, ILogger<MpcJobPoller> logger)
        {
            _scopeFactory = scopeFactory;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken cancellationToken)
        {
            while (!cancellationToken.IsCancellationRequested)
            {
                try
                {
                    await PollOnceAsync(cancellationToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[Wallet Service] MPC job poller error");
                }

                await Task.Delay(PollInterval, cancellationToken);
            }
        }

        private async Task PollOnceAsync(CancellationToken cancellationToken)
        {
            using var scope = _scopeFactory.CreateScope();

            var dbContext = scope.ServiceProvider.GetRequiredService<WalletDbContext>();
            var mpcService = scope.ServiceProvider.GetRequiredService<IWalletMpcService>();
            var queue = scope.ServiceProvider.GetRequiredService<IMpcCompletionQueue>();

            var staleCutoff = DateTime.UtcNow - PendingStaleAfter;
            var staleWallets = await dbContext.Wallets
                .Where(w => w.Status == WalletStatus.PENDING_KEY_GEN && w.MpcJobId != null && w.DateCreated < staleCutoff)
                .ToListAsync(cancellationToken);

            foreach (var staleWallet in staleWallets)
            {
                _logger.LogWarning("[Wallet Service] Marking wallet {WalletId} FAILED: pending for over {Minutes} minutes (MPC job lost or wedged)", staleWallet.Id, PendingStaleAfter.TotalMinutes);
                staleWallet.Status = WalletStatus.FAILED;
            }

            if (staleWallets.Count > 0)
            {
                await dbContext.SaveChangesAsync(cancellationToken);
            }

            var pendingWallets = await dbContext.Wallets
                .Where(w => w.Status == WalletStatus.PENDING_KEY_GEN && w.MpcJobId != null)
                .ToListAsync(cancellationToken);

            foreach (var wallet in pendingWallets)
            {
                try
                {
                    var status = await mpcService.GetStatusAsync(wallet.MpcJobId!, cancellationToken);

                    switch (status.Status)
                    {
                        case GetKeyGenStatusResponse.Types.Status.Completed:
                            await queue.QueueCompletionAsync(new MpcKeyGenCompletedNotification(wallet.Id, status.Xpub, status.BitcoinAddress), cancellationToken);
                            break;

                        case GetKeyGenStatusResponse.Types.Status.Failed:
                            _logger.LogError("[Wallet Service] MPC keygen FAILED for wallet {WalletId}: {Error}",wallet.Id, status.ErrorMessage);

                            wallet.Status = WalletStatus.FAILED;
                            await dbContext.SaveChangesAsync(cancellationToken);

                            break;
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[Wallet Service] Failed to poll MPC status for wallet {WalletId}", wallet.Id);
                }
            }
        }
    }
}
