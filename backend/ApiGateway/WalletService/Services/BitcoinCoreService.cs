using NBitcoin;
using NBitcoin.RPC;
using WalletService.Services.Interfaces;

namespace WalletService.Services
{
    public class BitcoinCoreService : IBitcoinCoreService
    {
        private readonly RPCClient _rpcClient;
        private readonly ILogger<BitcoinCoreService> _logger;

        public BitcoinCoreService(RPCClient rpcClient, ILogger<BitcoinCoreService> logger)
        {
            _rpcClient = rpcClient;
            _logger = logger;
        }

        public async Task RegisterWalletDescriptorAsync(string xpub, string network, CancellationToken ct)
        {
            Network btcNetwork = GetNetwork(network);
            ExtPubKey extPubKey = ExtPubKey.Parse(xpub, btcNetwork);

            byte[] hash160 = extPubKey.PubKey.Hash.ToBytes();
            string fingerprint = BitConverter
                .ToString(hash160.Take(4).ToArray())
                .Replace("-", "")
                .ToLowerInvariant();

            string receiveDesc = $"wpkh([{fingerprint}]{xpub}/0/*)";
            string changeDesc = $"wpkh([{fingerprint}]{xpub}/1/*)";

            _logger.LogInformation("[Wallet Service] Registering watch-only descriptors with bitcoind");
            _logger.LogDebug("[Wallet Service] Receive descriptor: {Descriptor}", receiveDesc);
            _logger.LogDebug("[Wallet Service] Change descriptor:  {Descriptor}", changeDesc);

            var descriptors = new[]
            {
                new { desc = receiveDesc, timestamp = "now", active = true },
                new { desc = changeDesc, timestamp = "now", active = true }
            };

            RPCResponse response = await _rpcClient.SendCommandAsync("importdescriptors", descriptors);

            if (response.Error is not null)
            {
                _logger.LogError("[Wallet Service] bitcoind importdescriptors RPC error: {Error}", response.Error.Message);
                throw new InvalidOperationException($"Failed to register wallet with bitcoind: {response.Error.Message}");
            }

            _logger.LogInformation("[Wallet Service] Successfully registered wallet descriptors with bitcoind (fingerprint: {Fingerprint})", fingerprint);
        }

        private static Network GetNetwork(string network)
        {
            return network?.Trim().ToLowerInvariant() switch
            {
                "mainnet" or "main" => Network.Main,
                "testnet" or "test" => Network.TestNet,
                "regtest" => Network.RegTest,
                _ => throw new ArgumentException($"Unsupported network type: '{network}'", nameof(network))
            };
        }
    }
}