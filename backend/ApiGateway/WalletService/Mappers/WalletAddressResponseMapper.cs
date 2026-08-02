using WalletService.DTOs;
using WalletService.Mappers.Interfaces;
using WalletService.Models;

namespace WalletService.Mappers
{
    public class WalletAddressResponseMapper : IMapper<WalletAddress, WalletAddressesResponseDTO>
    {
        public WalletAddress ToEntity(WalletAddressesResponseDTO data)
        {
            return null;
        }

        public WalletAddressesResponseDTO ToDto(WalletAddress data)
        {
            return new WalletAddressesResponseDTO
            {
                Address = data.Address,
                BalanceSatoshis = data.BalanceSatoshis,
                DateCreated = data.DateCreated
            }; 
        }
    }
}
