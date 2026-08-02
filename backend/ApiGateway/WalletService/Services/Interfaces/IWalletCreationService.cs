using NBitcoin;
using WalletService.DTOs;
using WalletService.Models;

namespace WalletService.Services.Interfaces
{
    public interface IWalletCreationService
    {
        Task<CreateWalletResponseDTO> GenerateWalletAsync(CreateWalletRequestDTO request);

        BitcoinAddress DeriveNewAddress(string networkType, string extPublicKey, int NextReceiveIndex, bool isChange);

        Mnemonic GenerateMnemonic(Wordlist wordlist, WordCount count);
    }
}
