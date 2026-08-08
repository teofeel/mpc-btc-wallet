using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Caching.Memory;
using NBitcoin;
using NBitcoin.RPC;
using Proto;
using TransactionService.DB;
using TransactionService.DTOs;
using TransactionService.Mappers;
using TransactionService.Models;
using TransactionService.Models.Enums;
using TransactionService.Services.Interfaces;

namespace TransactionService.Services
{
    public class TransactionService : ITransactionService
    {
        private readonly ChainStateService.ChainStateServiceClient _chainStateClient;
        private readonly Proto.TransactionService.TransactionServiceClient _transactionClient;

        private readonly TransactionDbContext _dbContext;
        private readonly IMemoryCache _cache;
        private readonly ILogger<TransactionService> _logger;
        private readonly Network _network;

        private readonly TransactionResponseMapper _transactionResponseMapper;

        private const int ESTIMATED_TX_VBYTES = 140;

        public TransactionService(ChainStateService.ChainStateServiceClient chainStateClient, Proto.TransactionService.TransactionServiceClient transactionClient,
            IMemoryCache cache, ILogger<TransactionService> logger, Network network, TransactionResponseMapper transactionResponseMapper)
        {
            _chainStateClient = chainStateClient;
            _transactionClient = transactionClient;
            _cache = cache;
            _logger = logger;
            _network = network;
            _transactionResponseMapper = transactionResponseMapper;
        }

        public Task<TransactionResponseDTO> MakeTransactionAsync(TransactionRequestDTO request)
        {
            return null;
        }

        public async Task<TransactionResponseDTO> GetTransactionByIdAsync(Guid id)
        {
            Models.Transaction tx = await _dbContext.Transactions.AsNoTracking().FirstOrDefaultAsync(t => t.Id == id);

            if (tx == null)
            {
                _logger.LogWarning("[Transaction Service] Transaction {Id} not found", id);
                throw new Exceptions.ResourceNotFoundException($"Transaction with ID '{id}' was not found");
            }

            if(!string.IsNullOrEmpty(tx.TxHash) && tx.Status != TxStatus.CONFIRMED)
            {
                try
                {
                    var response = await _transactionClient.GetTransactionAsync(new GetTransactionRequest{TxId = tx.TxHash});

                    if(response?.Transaction != null)
                    {
                        int liveConfirmations = (int)response.Transaction.Confirmations;
                        if (liveConfirmations != tx.Confirmations)
                        {
                            await UpdateTransactionConfirmationsAsync(tx.TxHash);
                            tx.Confirmations = liveConfirmations;
                        }
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogWarning(ex, "[Transaction Service] Failed to query live confirmations for TxHash: {TxHash}", tx.TxHash);
                }
            }

            return _transactionResponseMapper.ToDto(tx);
        }

        public async Task<FeeEstimationResponseDTO> GetFeeEstimationAsync(FeePriority priority = FeePriority.MEDIUM)
        {
            int targetBlocks = (int)priority;
            string cacheKey = $"fee_estimation_target_{targetBlocks}";

            if (_cache.TryGetValue(cacheKey, out FeeEstimationResponseDTO? cachedFeeDto) && cachedFeeDto != null)
            {
                return cachedFeeDto;
            }

            try
            {
                _logger.LogInformation("[Transaction Service] Requesting fee estimation from Go ChainStateService");

                var request = new GetFeeEstimationRequest
                {
                    TargetBlocks = targetBlocks,
                };

                var response = await _chainStateClient.GetFeeEstimationAsync(request);

                if (response?.Fee == null)
                {
                    _logger.LogWarning("ChainStateService returned null fee response");
                    return GetFallbackFee(priority);
                }

                decimal feePerVByteBtc = (decimal)response.Fee.FeeRateBtcKb / 1000m;
                decimal estimatedFeeBtc = Math.Round(feePerVByteBtc * ESTIMATED_TX_VBYTES, 8);

                var resultDto = new FeeEstimationResponseDTO
                {
                    Priority = priority.ToString(),
                    TargetBlocks = targetBlocks,
                    EstimatedFeeInBtc = estimatedFeeBtc,
                    EstimatedFeeInSatoshis = (long)(estimatedFeeBtc * 100_000_000m)
                };

                _cache.Set(cacheKey, resultDto, TimeSpan.FromSeconds(60));

                return resultDto;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Transaction Service] Failed to fetch fee estimation from gRPC service");
                return GetFallbackFee(priority);
            }
        }

        public Task<bool> IsAddressValidAsync(string bitcoinAddress)
        {
            if (string.IsNullOrWhiteSpace(bitcoinAddress))
                return Task.FromResult(false);

            try
            {
                BitcoinAddress.Create(bitcoinAddress, _network);
                return Task.FromResult(true);
            }
            catch
            {
                return Task.FromResult(false);
            }
        }

        public async Task UpdateTransactionConfirmationsAsync(string txHash)
        {
            Models.Transaction tx = await _dbContext.Transactions.FirstOrDefaultAsync(t => t.TxHash == txHash);

            if (tx == null) return;

            try
            {
                var response = await _transactionClient.GetTransactionAsync(new GetTransactionRequest { TxId = txHash });

                if (response?.Transaction != null)
                {
                    tx.Confirmations = (int)response.Transaction.Confirmations;

                    if (tx.Confirmations >= 6 && tx.Status != TxStatus.CONFIRMED)
                    {
                        tx.Status = TxStatus.CONFIRMED;
                        tx.ConfirmedAt = DateTime.UtcNow;
                    }

                    await _dbContext.SaveChangesAsync();
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Transaction Service] Error updating confirmations for {TxHash}", txHash);
            }
        }


        private static FeeEstimationResponseDTO GetFallbackFee(FeePriority priority)
        {
            decimal fallbackBtc = priority switch
            {
                FeePriority.LOW => 0.00001000m,    
                FeePriority.MEDIUM => 0.00002500m, 
                FeePriority.HIGH => 0.00005000m,   
                _ => 0.00002500m
            };

            return new FeeEstimationResponseDTO
            {
                Priority = priority.ToString(),
                TargetBlocks = (int)priority,
                EstimatedFeeInBtc = fallbackBtc,
                EstimatedFeeInSatoshis = (long)(fallbackBtc * 100_000_000m)
            };
        }
    }
}
