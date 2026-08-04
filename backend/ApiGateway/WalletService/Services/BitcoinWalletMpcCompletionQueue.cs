using System.Threading.Channels;
using WalletService.Events;
using WalletService.Services.Interfaces;

namespace WalletService.Services
{
    public class BitcoinWalletMpcCompletionQueue : IMpcCompletionQueue
    {
        private readonly Channel<MpcKeyGenCompletedNotification> _channel = Channel.CreateUnbounded<MpcKeyGenCompletedNotification>();

        public ValueTask QueueCompletionAsync(MpcKeyGenCompletedNotification notification, CancellationToken ct = default)
        {
            return _channel.Writer.WriteAsync(notification, ct);
        }

        public IAsyncEnumerable<MpcKeyGenCompletedNotification> ReadCompletionAsync(CancellationToken ct = default)
        {
            return _channel.Reader.ReadAllAsync(ct);
        }
    }
}
