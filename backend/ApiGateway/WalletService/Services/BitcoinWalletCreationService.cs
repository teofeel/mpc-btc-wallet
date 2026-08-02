using NBitcoin;
using System.Threading.Tasks;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Mappers;
using WalletService.Models;
using WalletService.Services.Interfaces;

namespace WalletService.Services
{
    public class BitcoinWalletCreationService : IWalletCreationService
    {
        private readonly WalletDbContext _dbContext;
        private readonly CreateWalletMapper _createWalletMapper;

        public BitcoinWalletCreationService(WalletDbContext dbContext, CreateWalletMapper createWalletMapper)
        {
            _dbContext = dbContext;
            _createWalletMapper = createWalletMapper;
        }

        private Network GetNetwork(string network)
        {
            return network?.Trim().ToLowerInvariant() switch
            {
                "mainnet" or "main" => Network.Main,
                "testnet" or "test" => Network.TestNet,
                "regtest" => Network.RegTest,
                _ => throw new ArgumentException($"Unsupported network type: '{network}'", nameof(network))
            };
        }

        public async Task<CreateWalletResponseDTO> GenerateWalletAsync(CreateWalletRequestDTO request)
        {
            Network network = GetNetwork(request.Network);
            string coinType = network == Network.Main ? "0'" : "1'";

            Mnemonic mnemonic = GenerateMnemonic(Wordlist.English, WordCount.Twelve);
            ExtKey masterKey = mnemonic.DeriveExtKey();

            KeyPath keyPath = new KeyPath($"m/84'/{coinType}/0'");
            ExtPubKey pubKey = masterKey.Derive(keyPath).Neuter();

            // mnemonic should be stored encrypted 
            Wallet wallet = new Wallet
            {
                Name = request.Name,
                Network = network.Name,
                ExtendedPublicKey = pubKey.ToString(network),
                EncryptedMnemonic = mnemonic.ToString(), // TODO encrypt it
                DerivationPath = keyPath.ToString()
            };

            BitcoinAddress btcAddress = this.DeriveNewAddress(request.Network, wallet.ExtendedPublicKey, wallet.NextReceiveIndex);

            WalletAddress address = new WalletAddress
            {
                Wallet = wallet,
                Address = btcAddress.ToString(),
                AddressIndex = wallet.NextReceiveIndex,
                IsChange = false
            };

            wallet.NextReceiveIndex++;

            await _dbContext.Wallets.AddAsync(wallet);
            await _dbContext.WalletAddresses.AddAsync(address);
            await _dbContext.SaveChangesAsync();

            CreateWalletResponseDTO response = _createWalletMapper.ToDto(wallet);
            response.Address = address.Address;

            return response;
        }

        public BitcoinAddress DeriveNewAddress(string networkType, string extPublicKey, int nextReceiveIndex, bool isChange = false)
        {
            Network network = GetNetwork(networkType);

            ExtPubKey accountXPub = ExtPubKey.Parse(extPublicKey, network);

            uint changePath = isChange ? 1u : 0;

            BitcoinAddress receiveAddress = accountXPub
                .Derive(changePath)
                .Derive((uint)nextReceiveIndex)
                .PubKey
                .GetAddress(ScriptPubKeyType.Segwit, network);

            return receiveAddress;
        }

        public Mnemonic GenerateMnemonic(Wordlist wordlist, WordCount count)
        {
            return new Mnemonic(wordlist, count);
        }

    }
}
