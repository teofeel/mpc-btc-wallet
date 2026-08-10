using MediatR;
using WalletService.Events;
using WalletService.Services.Interfaces;

namespace WalletService.Features.Handlers
{
    public class MpcKeyGenCompletedHandler : INotificationHandler<MpcKeyGenCompletedNotification>
    {
        private readonly IWalletCreationService _walletCreationService;

        public MpcKeyGenCompletedHandler(IWalletCreationService walletCreationService)
        {
            _walletCreationService = walletCreationService;
        }

        public async Task Handle(MpcKeyGenCompletedNotification notification, CancellationToken cancellationToken)
        {
            await _walletCreationService.CompleteWalletSetupAsync(
                notification.WalletId,
                notification.ExtendedPublicKey,
                notification.Address,
                cancellationToken
            );
        }
    }
}
