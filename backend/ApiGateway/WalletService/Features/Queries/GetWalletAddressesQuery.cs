using MediatR;
using WalletService.DTOs;

namespace WalletService.Features.Queries
{
    public record GetWalletAddressesQuery(Guid id) : IRequest<List<WalletAddressesResponseDTO>>;
}
