using MediatR;
using Microsoft.EntityFrameworkCore;
using NBitcoin;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Features.Queries;
using WalletService.Models;

namespace WalletService.Features.Handlers
{
    public class GetWalletHandler : IRequestHandler<GetWalletQuery, GetWalletResponseDTO>
    {
        private readonly WalletDbContext _dbContext;
        private readonly ILogger<GetWalletHandler> _logger;

        public GetWalletHandler(WalletDbContext dbContext, ILogger<GetWalletHandler> logger)
        {
            _dbContext = dbContext;
            _logger = logger;
        }

        public async Task<GetWalletResponseDTO> Handle(GetWalletQuery query, CancellationToken token)
        {
            _logger.LogInformation("[Wallet Service] Fetching wallet {Id}", query.Id);

            Wallet? wallet = await _dbContext.Wallets
                .Include(w => w.Addresses)
                .AsNoTracking()
                .FirstOrDefaultAsync(w => w.Id == query.Id, token);

            if (wallet is null)
                throw new ResourceNotFoundException("Couldn't find wallet with this Id");

            ExtPubKey? accountXPub = TryParseAccountXPub(wallet);

            return new GetWalletResponseDTO
            {
                WalletId = wallet.Id,
                Name = wallet.Name,
                Network = wallet.Network,
                Status = wallet.Status.ToString(),
                DerivationPath = wallet.DerivationPath,
                ExtendedPublicKey = wallet.ExtendedPublicKey,
                Addresses = wallet.Addresses
                    .OrderBy(a => a.AddressIndex)
                    .Select(address => new WalletAddressInfoDTO
                    {
                        Address = address.Address,
                        AddressIndex = address.AddressIndex,
                        IsChange = address.IsChange,
                        PublicKeyHex = accountXPub != null ? DerivePubKeyHex(accountXPub, address) : string.Empty,
                    })
                    .ToList(),
            };
        }

        private ExtPubKey? TryParseAccountXPub(Wallet wallet)
        {
            if (string.IsNullOrWhiteSpace(wallet.ExtendedPublicKey))
            {
                return null;
            }

            try
            {
                return ExtPubKey.Parse(wallet.ExtendedPublicKey, GetNetwork(wallet.Network));
            }
            catch (Exception ex)
            {
                _logger.LogWarning(ex, "[Wallet Service] Failed to parse extended public key for wallet {WalletId}", wallet.Id);
                return null;
            }
        }

        /// <summary>
        /// Compressed pubkey (33 bytes, hex) that owns the given address.
        /// The first address (index 0, receive) is the keygen master address, which is
        /// the xpub's own key; every later address is a BIP84 child of the account xpub.
        /// </summary>
        private static string DerivePubKeyHex(ExtPubKey accountXPub, WalletAddress address)
        {
            PubKey pubKey = address.AddressIndex == 0 && !address.IsChange
                ? accountXPub.PubKey
                : accountXPub.Derive(address.IsChange ? 1u : 0u).Derive((uint)address.AddressIndex).PubKey;

            return pubKey.ToHex();
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
