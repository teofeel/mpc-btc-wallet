using WalletService.DTOs;
using WalletService.Mappers.Interfaces;
using WalletService.Models;

namespace WalletService.Mappers
{
    public class StartWalletCreationMapper : IMapper<Wallet, StartWalletCreationResponseDTO>
    {
        public Wallet ToEntity(StartWalletCreationResponseDTO data)
        {
            return null;
        }

        public StartWalletCreationResponseDTO ToDto(Wallet data)
        {
            return new StartWalletCreationResponseDTO
            {
                Id = data.Id,
                Name = data.Name,
                Network = data.Network,
                JobId = data.MpcJobId
            };
        }
    }
}
