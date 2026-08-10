using MediatR;
using Microsoft.EntityFrameworkCore;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Features.Commands;
using WalletService.Features.Queries;
using WalletService.Mappers.Interfaces;
using WalletService.Models;

namespace WalletService.Features.Handlers
{
    public class GetWalletAddressesHandler : IRequestHandler<GetWalletAddressesQuery, List<WalletAddressesResponseDTO>>
    {
        private readonly WalletDbContext _dbContext;
        private readonly IMapper<WalletAddress, WalletAddressesResponseDTO> _mapper;
        private readonly ILogger<GetWalletAddressesHandler> _logger;

        public GetWalletAddressesHandler(WalletDbContext dbContext, IMapper<WalletAddress, WalletAddressesResponseDTO> mapper, ILogger<GetWalletAddressesHandler> logger)
        {
            _dbContext = dbContext;
            _mapper = mapper;
            _logger = logger;
        }

        public async Task<List<WalletAddressesResponseDTO>> Handle(GetWalletAddressesQuery query, CancellationToken token)
        {
            _logger.LogInformation("[Wallet Service] Fetching wallet addresses for {Id}", query.id);


            Wallet? wallet = await _dbContext.Wallets
                .Include(w => w.Addresses)
                .AsNoTracking()
                .FirstOrDefaultAsync(w => w.Id == query.id, token);

            if (wallet is null)
                throw new ResourceNotFoundException("Couldn't find wallet with this Id");


            return wallet.Addresses.Select(address =>
            {
                WalletAddressesResponseDTO dto = _mapper.ToDto(address);
                dto.Name = wallet.Name;
                dto.Network = wallet.Network;
                return dto;
            }).ToList();
        }
    }
}
