using MediatR;

namespace WalletService.Events
{
    public record MpcKeyGenCompletedNotification(
        Guid WalletId,
        string PublicKeyHex,
        string Address
    ) : INotification;
}
