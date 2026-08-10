using WalletService.Events;

namespace WalletService.Services.Interfaces
{
    public interface IMpcCompletionQueue
    {
        ValueTask QueueCompletionAsync(MpcKeyGenCompletedNotification notification, CancellationToken ct = default);
        IAsyncEnumerable<MpcKeyGenCompletedNotification> ReadCompletionAsync(CancellationToken ct = default);
    }
}
