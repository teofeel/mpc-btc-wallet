using WalletService.Services.MpcService;

namespace WalletService.Services.Interfaces
{
    public interface IWalletMpcService
    {
        Task<string> StartKeyGenerationAsync(string walletId, string idempotencyKey, string network, CancellationToken ct);
        Task<GetKeyGenStatusResponse> GetStatusAsync(string jobId, CancellationToken ct);
    }
}
