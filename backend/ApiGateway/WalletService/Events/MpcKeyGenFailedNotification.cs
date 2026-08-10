using MediatR;

namespace WalletService.Events
{
    public record MpcKeyGenFailedNotification(
        Guid WalletId,
        string ErrorMessage
    ) : INotification;
}
