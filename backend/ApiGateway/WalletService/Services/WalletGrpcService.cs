using Grpc.Core;
using MediatR;
using WalletInfo;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Features.Queries;

namespace WalletService.Services
{
    public class WalletGrpcService : WalletServiceGrpc.WalletServiceGrpcBase
    {
        private readonly IMediator _mediator;

        public WalletGrpcService(IMediator mediator)
        {
            _mediator = mediator;
        }

        public override async Task<GetWalletResponse> GetWallet(GetWalletRequest request, ServerCallContext context)
        {
            if (!Guid.TryParse(request.WalletId, out Guid walletId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "wallet_id must be a valid UUID"));
            }

            GetWalletResponseDTO wallet;
            try
            {
                wallet = await _mediator.Send(new GetWalletQuery(walletId), context.CancellationToken);
            }
            catch (ResourceNotFoundException)
            {
                throw new RpcException(new Status(StatusCode.NotFound, $"Wallet '{request.WalletId}' was not found"));
            }

            GetWalletResponse response = new GetWalletResponse
            {
                WalletId = wallet.WalletId.ToString(),
                Name = wallet.Name,
                Network = wallet.Network,
                Status = wallet.Status,
                DerivationPath = wallet.DerivationPath,
                ExtendedPublicKey = wallet.ExtendedPublicKey,
            };

            foreach (WalletAddressInfoDTO address in wallet.Addresses)
            {
                response.Addresses.Add(new WalletAddressDto
                {
                    Address = address.Address,
                    AddressIndex = address.AddressIndex,
                    IsChange = address.IsChange,
                    PublicKeyHex = address.PublicKeyHex,
                });
            }

            return response;
        }
    }
}
