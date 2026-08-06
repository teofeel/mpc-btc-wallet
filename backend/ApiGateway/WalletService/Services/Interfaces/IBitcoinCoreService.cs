namespace WalletService.Services.Interfaces
{
    public interface IBitcoinCoreService
    {
        Task RegisterWalletDescriptorAsync(string xpub, string network, CancellationToken ct);
    }
}
