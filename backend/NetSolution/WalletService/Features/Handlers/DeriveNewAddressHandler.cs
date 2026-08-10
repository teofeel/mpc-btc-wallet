using MediatR;
using Microsoft.AspNetCore.Razor.TagHelpers;
using NBitcoin;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Features.Commands;
using WalletService.Mappers;
using WalletService.Models;
using WalletService.Services.Interfaces;

namespace WalletService.Features.Handlers
{
    public class DeriveNewAddressHandler : IRequestHandler<DeriveNewAddressCommand, DeriveWalletAddressResponseDTO>
    {
        private readonly WalletDbContext _dbContext;
        private readonly IWalletCreationService _walletCreationService;
        private readonly DeriveAddressMapper _deriveAddressMapper;
        private readonly ILogger<CreateWalletHandler> _logger;

        public DeriveNewAddressHandler(WalletDbContext dbContext, IWalletCreationService walletCreationService, DeriveAddressMapper deriveAddressMapper, ILogger<CreateWalletHandler> logger)
        {
            _dbContext = dbContext;
            _walletCreationService = walletCreationService;
            _deriveAddressMapper = deriveAddressMapper;
            _logger = logger;
        }

        public async Task<DeriveWalletAddressResponseDTO> Handle(DeriveNewAddressCommand command, CancellationToken token)
        {
            Wallet wallet = await _dbContext.Wallets.FindAsync(command.id) ?? throw new ResourceNotFoundException($"Couldn't find wallet with this id {command.id}");


            BitcoinAddress btcAddress = _walletCreationService.DeriveNewAddress(wallet.Network, wallet.ExtendedPublicKey, wallet.NextReceiveIndex, false);

            WalletAddress address = new WalletAddress
            {
                Id = Guid.NewGuid(),
                WalletId = wallet.Id,
                Wallet = wallet,
                Address = btcAddress.ToString(),
                AddressIndex = wallet.NextReceiveIndex,
                IsChange = false,
                BalanceSatoshis = 0,
                DateCreated = DateTime.UtcNow,
            };

            wallet.Addresses.Add(address);
            wallet.NextReceiveIndex++;

            await _dbContext.WalletAddresses.AddAsync(address, token);
            await _dbContext.SaveChangesAsync(token);

            return _deriveAddressMapper.ToDto(address);
        }
    }
}
