using WalletService.DTOs;
using WalletService.Mappers.Interfaces;
using WalletService.Models;

namespace WalletService.Mappers
{
    public class CreateWalletMapper : IMapper<Wallet, CreateWalletResponseDTO>
    {
        public Wallet ToEntity(CreateWalletResponseDTO data)
        {
            return null;
        }

        public CreateWalletResponseDTO ToDto(Wallet data)
        {
            return new CreateWalletResponseDTO
            {
                Name = data.Name,
                Network = data.Network
            };
        }
    }
}
