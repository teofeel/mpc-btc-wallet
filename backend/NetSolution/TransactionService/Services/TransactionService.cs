using System.Numerics;
using Google.Protobuf;
using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Caching.Memory;
using NBitcoin;
using NBitcoin.Crypto;
using Mpc;
using Proto;
using TransactionService.DB;
using TransactionService.DTOs;
using TransactionService.Mappers;
using TransactionService.Models;
using TransactionService.Models.Enums;
using TransactionService.Services.Interfaces;
using WalletInfo;
using Transaction = TransactionService.Models.Transaction;
using BitcoinTransaction = NBitcoin.Transaction;

namespace TransactionService.Services
{
    public class TransactionService : ITransactionService
    {
        private readonly ChainStateService.ChainStateServiceClient _chainStateClient;
        private readonly TransactionMpcService.TransactionMpcServiceClient _transactionMpcClient;
        private readonly Proto.TransactionService.TransactionServiceClient _transactionClient;
        private readonly AddressService.AddressServiceClient _addressClient;
        private readonly WalletServiceGrpc.WalletServiceGrpcClient _walletClient;

        private readonly TransactionDbContext _dbContext;
        private readonly IMemoryCache _cache;
        private readonly ILogger<TransactionService> _logger;
        private readonly Network _network;

        private readonly TransactionResponseMapper _transactionResponseMapper;

        private const int ESTIMATED_TX_VBYTES = 140;
        private const long DUST_SATOSHIS = 546;

        public TransactionService(ChainStateService.ChainStateServiceClient chainStateClient, TransactionMpcService.TransactionMpcServiceClient transactionMpcClient, Proto.TransactionService.TransactionServiceClient transactionClient,
            AddressService.AddressServiceClient addressClient, WalletServiceGrpc.WalletServiceGrpcClient walletClient,
            TransactionDbContext dbContext, IMemoryCache cache, ILogger<TransactionService> logger, Network network, TransactionResponseMapper transactionResponseMapper)
        {
            _chainStateClient = chainStateClient;
            _transactionMpcClient = transactionMpcClient;
            _transactionClient = transactionClient;
            _addressClient = addressClient;
            _walletClient = walletClient;
            _dbContext = dbContext;
            _cache = cache;
            _logger = logger;
            _network = network;
            _transactionResponseMapper = transactionResponseMapper;
        }

        private string GetDefaultDerivationPath(string network)
        {
            return network.Equals("Regtest", StringComparison.OrdinalIgnoreCase)
                   || network.Equals("Testnet", StringComparison.OrdinalIgnoreCase)
                ? "m/84'/1'/0'/0/0"
                : "m/84'/0'/0'/0/0";
        }

        public async Task<TransactionResponseDTO> MakeTransactionAsync(TransactionRequestDTO request)
        {
            FeeEstimationResponseDTO feeEstimate = await GetFeeEstimationAsync(request.FeePriority);

            Transaction transaction = new Models.Transaction
            {
                Id = Guid.NewGuid(),
                WalletId = request.WalletId,
                ReceivingAddress = request.ReceivingAddress,
                Amount = request.Amount,
                Fee = feeEstimate.EstimatedFeeInBtc,
                Status = TxStatus.PENDING,
                CreatedAt = DateTime.UtcNow
            };

            await _dbContext.Transactions.AddAsync(transaction);
            await _dbContext.SaveChangesAsync();

            try
            {
                _logger.LogInformation("[Transaction Service] Assembling and signing transaction {Id}...", transaction.Id);
                string signedHexTx = await AssembleSignedRawTransactionHex(request, transaction);

                _logger.LogInformation("[Transaction Service] Broadcasting transaction for: {Id}", transaction.Id);

                BroadcastTransactionRequest broadcastTransactionRequest = new BroadcastTransactionRequest { HexTx = signedHexTx };
                BroadcastTransactionResponse broadcastTransactionResponse = await _transactionClient.BroadcastTransactionAsync(broadcastTransactionRequest);

                transaction.TxHash = broadcastTransactionResponse.TxId;
                transaction.Status = TxStatus.BROADCASTED;
                transaction.BroadcastedAt = DateTime.UtcNow;

                await _dbContext.SaveChangesAsync();

                _logger.LogInformation("[Transaction Service] Transaction ID {Id} successfully broadcasted with TxHash {TxHash}", transaction.Id, transaction.TxHash);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[Transaction Service] Signing transaction {TxHash} failed", transaction.TxHash);
                transaction.Status = TxStatus.FAILED;
                await _dbContext.SaveChangesAsync();

                throw;
            }

            return _transactionResponseMapper.ToDto(transaction);
        }

        public async Task<TransactionResponseDTO> GetTransactionByIdAsync(Guid id)
        {
            Transaction? tx = await _dbContext.Transactions.AsNoTracking().FirstOrDefaultAsync(t => t.Id == id);

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

                // On regtest (and other low-activity chains) estimatesmartfee returns no fee
                // rate, which the Go service forwards as 0. Fall back to a fixed estimate so
                // the transaction clears the node's minimum relay fee instead of being
                // rejected with "min relay fee not met".
                if (ToSatoshis(estimatedFeeBtc) < ESTIMATED_TX_VBYTES)
                {
                    _logger.LogWarning("ChainStateService fee estimate too low ({FeeRateBtcKb} BTC/kB), using fallback", response.Fee.FeeRateBtcKb);
                    return GetFallbackFee(priority);
                }

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
            Transaction? tx = await _dbContext.Transactions.FirstOrDefaultAsync(t => t.TxHash == txHash);

            if (tx == null) return;

            try
            {
                var response = await _transactionClient.GetTransactionAsync(new GetTransactionRequest { TxId = txHash });

                if (response?.Transaction != null)
                {
                    int newConfirmations = (int)response.Transaction.Confirmations;
                    tx.Confirmations = newConfirmations;
                    
                    if (newConfirmations < 0)
                    {
                        _logger.LogWarning("[Transaction Service] Transaction {TxHash} Failed", txHash);
                        tx.Status = TxStatus.FAILED;
                    }

                    if (newConfirmations >= 6 && tx.Status != TxStatus.CONFIRMED && tx.Status != TxStatus.FAILED)
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
        
        private async Task<string> AssembleSignedRawTransactionHex(TransactionRequestDTO request, Models.Transaction transaction)
        {
            GetWalletResponse wallet = await GetWalletOrThrowAsync(request.WalletId);

            WalletAddressDto sendingAddress = wallet.Addresses.FirstOrDefault()
                ?? throw new ArgumentException($"Wallet '{request.WalletId}' has no addresses to spend from");
            
            if (string.IsNullOrWhiteSpace(sendingAddress.PublicKeyHex))
                throw new ArgumentException($"Wallet '{request.WalletId}' has no public key for its sending address");

            byte[] pubKeyBytes = Convert.FromHexString(sendingAddress.PublicKeyHex);
            PubKey pubKey = new PubKey(pubKeyBytes);

            BitcoinAddress sendingBtcAddress = BitcoinAddress.Create(sendingAddress.Address, _network);
            Script scriptPubKey = sendingBtcAddress.ScriptPubKey;
            if (!scriptPubKey.Equals(new PubKey(pubKeyBytes).GetScriptPubKey(ScriptPubKeyType.Segwit)))
                throw new ArgumentException("Wallet public key does not match its sending address");
            
            Script scriptCode = pubKey.Hash.ScriptPubKey;

            GetAddressUTXOsResponse utxoResponse = await _addressClient.GetAddressUTXOsAsync(new GetAddressUTXOsRequest { Address = sendingAddress.Address });
            List<UTXO> utxos = utxoResponse.Utxos.Where(u => u.Confirmations > 0).ToList();
            if (utxos.Count == 0)
                throw new ArgumentException($"No confirmed UTXOs found for sending address {sendingAddress.Address}");

            long amountSats = ToSatoshis(request.Amount);
            long feeSats = ToSatoshis(transaction.Fee);
            long totalInputSats = utxos.Sum(u => u.AmountSats);
            long changeSats = totalInputSats - amountSats - feeSats;
            
            if (changeSats < 0)
            {
                throw new InvalidOperationException($"Insufficient funds");
            }

            string derivationPath = string.IsNullOrWhiteSpace(request.DerivationPath)
                ? GetDefaultDerivationPath(request.Network)
                : request.DerivationPath;

            BitcoinTransaction tx = BitcoinTransaction.Create(_network);
            tx.Version = 2;

            foreach (UTXO utxo in utxos)
            {
                tx.Inputs.Add(new TxIn(new OutPoint(uint256.Parse(utxo.TxId), utxo.Vout)));
            }
            

            tx.Outputs.Add(Money.Satoshis(amountSats), BitcoinAddress.Create(request.ReceivingAddress, _network));
            if (changeSats >= DUST_SATOSHIS)
                tx.Outputs.Add(Money.Satoshis(changeSats), sendingBtcAddress);

            transaction.SendingAddress = sendingAddress.Address;
            
            for (int i = 0; i < utxos.Count; i++)
            {
                UTXO utxo = utxos[i];
                Money inputAmount = Money.Satoshis(utxo.AmountSats);

                uint256 sighash = tx.GetSignatureHash(scriptCode, i, SigHash.All, new TxOut(inputAmount, scriptPubKey), HashVersion.WitnessV0);

                SignRequest signRequest = new SignRequest
                {
                    WalletId = request.WalletId.ToString(),
                    IdempotencyKey = $"{transaction.Id}:{i}",
                    Network = request.Network,
                    TxHash = ByteString.CopyFrom(sighash.ToBytes()),
                    DerivationPath = derivationPath,
                };

                SignResponse signResponse = await _transactionMpcClient.SignAsync(signRequest, deadline: DateTime.UtcNow.AddSeconds(45));

                byte[] derSignature = EncodeDerSignature(signResponse.R.ToByteArray(), signResponse.S.ToByteArray());
                TransactionSignature txSignature = new TransactionSignature(new ECDSASignature(derSignature), SigHash.All);

                tx.Inputs[i].WitScript = new WitScript(new byte[][] { txSignature.ToBytes(), pubKeyBytes });
            }

            return tx.ToHex();
        }

        private async Task<GetWalletResponse> GetWalletOrThrowAsync(Guid walletId)
        {
            try
            {
                return await _walletClient.GetWalletAsync(new GetWalletRequest { WalletId = walletId.ToString() });
            }
            catch (RpcException ex) when (ex.StatusCode == StatusCode.NotFound)
            {
                throw new Exceptions.ResourceNotFoundException($"Wallet '{walletId}' was not found");
            }
        }
        
        private static byte[] EncodeDerSignature(byte[] rBytes, byte[] sBytes)
        {
            byte[] r = EncodeDerInteger(rBytes);
            byte[] s = EncodeDerInteger(ToLowS(sBytes));

            byte[] der = new byte[6 + r.Length + s.Length];
            der[0] = 0x30;
            der[1] = (byte)(4 + r.Length + s.Length);
            der[2] = 0x02;
            der[3] = (byte)r.Length;
            Buffer.BlockCopy(r, 0, der, 4, r.Length);
            der[4 + r.Length] = 0x02;
            der[5 + r.Length] = (byte)s.Length;
            Buffer.BlockCopy(s, 0, der, 6 + r.Length, s.Length);

            ECDSASignature ecdsa = ECDSASignature.FromDER(der);
            return ecdsa.MakeCanonical().ToDER();
        }

        private static byte[] EncodeDerInteger(byte[] value)
        {
            int start = 0;
            while (start < value.Length - 1 && value[start] == 0) start++;

            byte[] trimmed = value[start..];
            bool needsPad = (trimmed[0] & 0x80) != 0;

            byte[] result = new byte[trimmed.Length + (needsPad ? 1 : 0)];
            if (needsPad) result[0] = 0x00;
            Buffer.BlockCopy(trimmed, 0, result, needsPad ? 1 : 0, trimmed.Length);

            return result;
        }

        private static readonly BigInteger Secp256k1N = BigInteger.Parse(
            "00FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141",
            System.Globalization.NumberStyles.HexNumber);
        
        private static byte[] ToLowS(byte[] sBytes)
        {
            BigInteger s = new BigInteger(sBytes, isUnsigned: true, isBigEndian: true);
            if (s > Secp256k1N / 2)
                s = Secp256k1N - s;

            return s.ToByteArray(isUnsigned: true, isBigEndian: true);
        }

        private static long ToSatoshis(decimal btc) => (long)(btc * 100_000_000m);

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
