using NBitcoin;
using WalletService.DTOs;
using WalletService.Models;

namespace WalletService.Services.Interfaces
{
    public interface IWalletCreationService
    {
        Task<StartWalletCreationResponseDTO> GenerateWalletAsync(CreateWalletRequestDTO request, CancellationToken ct);
        Task<CreateWalletResponseDTO> CompleteWalletSetupAsync(Guid walletId, string extendedPublicKey, string address, CancellationToken ct = default);
        BitcoinAddress DeriveNewAddress(string networkType, string extPublicKey, int NextReceiveIndex, bool isChange);
    }
}
