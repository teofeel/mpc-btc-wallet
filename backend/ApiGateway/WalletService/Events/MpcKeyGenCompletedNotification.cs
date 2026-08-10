using MediatR;

namespace WalletService.Events
{
    public record MpcKeyGenCompletedNotification(
        Guid WalletId,
        string ExtendedPublicKey,
        string Address
    ) : INotification;
}
