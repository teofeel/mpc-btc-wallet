using WalletService.DTOs;
using MediatR;

namespace WalletService.Features.Commands
{
    public record CreateWalletCommand(CreateWalletRequestDTO request) : IRequest<CreateWalletResponseDTO>
    {
    }
}
