using MediatR;
using WalletService.DTOs;

namespace WalletService.Features.Queries
{
    public record GetWalletQuery(Guid Id) : IRequest<GetWalletResponseDTO>;
}
