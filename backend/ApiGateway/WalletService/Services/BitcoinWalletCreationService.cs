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
        private readonly StartWalletCreationMapper _startWalletCreationMapper;
        private readonly IWalletMpcService _walletMpcService;
        private readonly IBitcoinCoreService _bitcoinCoreService;
        private readonly ILogger<BitcoinWalletCreationService> _logger;

        public BitcoinWalletCreationService(WalletDbContext dbContext, StartWalletCreationMapper startWalletCreationMapper,
            CreateWalletMapper createWalletMapper, IWalletMpcService walletMpcService,
            IBitcoinCoreService bitcoinCoreService, ILogger<BitcoinWalletCreationService> logger)
        {
            _dbContext = dbContext;
            _startWalletCreationMapper = startWalletCreationMapper;
            _createWalletMapper = createWalletMapper;
            _walletMpcService = walletMpcService;
            _bitcoinCoreService = bitcoinCoreService;
            _logger = logger;
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

        public async Task<StartWalletCreationResponseDTO> GenerateWalletAsync(CreateWalletRequestDTO request, CancellationToken ct)
        {
            Network network = GetNetwork(request.Network);
            string coinType = network == Network.Main ? "0'" : "1'";
            KeyPath keyPath = new KeyPath($"m/84'/{coinType}/0'");

            Wallet wallet = new Wallet
            {
                Id = Guid.NewGuid(),
                Name = request.Name,
                Network = network.Name,
                DerivationPath = keyPath.ToString(),
                Status = Models.Enums.WalletStatus.PENDING_KEY_GEN,
                Addresses = new List<WalletAddress>()
            };

            string idempotencyKey = Guid.NewGuid().ToString();
            string jobId = await _walletMpcService.StartKeyGenerationAsync(wallet.Id.ToString(), idempotencyKey, request.Network, ct);

            wallet.MpcJobId = jobId;


            await _dbContext.Wallets.AddAsync(wallet);
            await _dbContext.SaveChangesAsync();

            return _startWalletCreationMapper.ToDto(wallet);
        }

        public async Task<CreateWalletResponseDTO> CompleteWalletSetupAsync(Guid walletId, string extendedPublicKey, string derivedAddress, CancellationToken ct = default)
        {
            var wallet = await _dbContext.Wallets.FindAsync(new object[] { walletId }, ct) ?? throw new KeyNotFoundException($"[Wallet Service] Wallet {walletId} not found");

            wallet.ExtendedPublicKey = extendedPublicKey;
            wallet.Status = Models.Enums.WalletStatus.ACTIVE;

            var address = new WalletAddress
            {
                Wallet = wallet,
                Address = derivedAddress,
                AddressIndex = wallet.NextReceiveIndex,
                IsChange = false
            };

            wallet.NextReceiveIndex++;
            wallet.Addresses.Add(address);

            await _dbContext.WalletAddresses.AddAsync(address);
            await _dbContext.SaveChangesAsync();

            try
            {
                await _bitcoinCoreService.RegisterWalletDescriptorAsync(
                    extendedPublicKey, wallet.Network, ct);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex,
                    "[Wallet Service] Wallet {WalletId} activated but descriptor registration with bitcoind failed " +
                    "The node will not watch this wallet until the descriptors are imported",
                    walletId);
            }

            return _createWalletMapper.ToDto(wallet);
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

    }
}
