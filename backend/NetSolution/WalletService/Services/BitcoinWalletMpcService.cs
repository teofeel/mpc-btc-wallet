using Grpc.Core;
using NBitcoin;
using WalletService.Models;
using WalletService.Services.Interfaces;
using WalletService.Services.MpcService;

namespace WalletService.Services
{
    public class BitcoinWalletMpcService : IWalletMpcService
    {
        private readonly WalletMpcService.WalletMpcServiceClient _client;
        private readonly ILogger<BitcoinWalletMpcService> _logger;

        public BitcoinWalletMpcService(WalletMpcService.WalletMpcServiceClient client, ILogger<BitcoinWalletMpcService> logger)
        {
            _client = client;
            _logger = logger;
        }

        public async Task<string> StartKeyGenerationAsync(string walletId, string idempotencyKey, string network, CancellationToken ct)
        {
            try
            {
                var request = new StartKeyGenRequest
                {
                    WalletId = walletId,
                    IdempotencyKey = idempotencyKey,
                    Network = network,
                };

                var response = await _client.StartKeyGenAsync(request, cancellationToken: ct);

                return response.JobId;
            }
            catch (RpcException ex) when (ex.StatusCode == StatusCode.InvalidArgument)
            {
                _logger.LogWarning("Go server rejected invalid input: {Detail}", ex.Status.Detail);
                throw new ArgumentException(ex.Status.Detail, ex);
            }
            catch (RpcException ex)
            {
                _logger.LogError(ex, "[Wallet Service] gRPC error during StartKeyGen: {Status}", ex.Status);
                throw;
            }
            catch (Exception ex) 
            {
                _logger.LogError("[Wallet Service] gRPC error during StartKeyGen: {Status}", ex.Message);
                throw;
            }
        }

        public async Task<GetKeyGenStatusResponse> GetStatusAsync(string jobId, CancellationToken ct)
        {
            try
            {
                var request = new GetKeyGenStatusRequest{JobId = jobId};

                return await _client.GetKeyGenStatusAsync(request, cancellationToken: ct);
            }
            catch (RpcException ex)
            {
                _logger.LogError(ex, "[Wallet Service] gRPC error fetching status for JobID: {JobId}", jobId);
                throw;
            }
        }
    }
}
