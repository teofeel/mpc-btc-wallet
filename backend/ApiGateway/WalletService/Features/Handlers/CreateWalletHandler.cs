using MediatR;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Features.Commands;
using WalletService.Mappers;
using WalletService.Services;
using WalletService.Services.Interfaces;

namespace WalletService.Features.Handlers
{
    public class CreateWalletHandler : IRequestHandler<CreateWalletCommand, StartWalletCreationResponseDTO>
    {
        private readonly IWalletCreationService _walletCreationService;
        private readonly ILogger<CreateWalletHandler> _logger;

        public CreateWalletHandler(IWalletCreationService walletCreationService, ILogger<CreateWalletHandler> logger)
        {
            _walletCreationService = walletCreationService;
            _logger = logger;
        }

        public async Task<StartWalletCreationResponseDTO> Handle(CreateWalletCommand command, CancellationToken token)
        {
            _logger.LogInformation("[Wallet Service] Creating new wallet....");
            return await _walletCreationService.GenerateWalletAsync(command.request, token);
        }
    }
}
