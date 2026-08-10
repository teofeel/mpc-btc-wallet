using MediatR;
using WalletService.DTOs;

namespace WalletService.Features.Commands
{
    public record DeriveNewAddressCommand(Guid id) : IRequest<DeriveWalletAddressResponseDTO>
    {
    }
}
