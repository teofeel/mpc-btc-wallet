using WalletService.DTOs;
using WalletService.Mappers.Interfaces;
using WalletService.Models;

namespace WalletService.Mappers
{
    public class DeriveAddressMapper : IMapper<WalletAddress, DeriveWalletAddressResponseDTO>
    {
        public WalletAddress ToEntity(DeriveWalletAddressResponseDTO data)
        {
            return null;
        }

        public DeriveWalletAddressResponseDTO ToDto(WalletAddress data)
        {
            return new DeriveWalletAddressResponseDTO
            {
                Address = data.Address,
                BalanceSatoshis = data.BalanceSatoshis,
                DateCreated = data.DateCreated,
            };
        }
    }
}
