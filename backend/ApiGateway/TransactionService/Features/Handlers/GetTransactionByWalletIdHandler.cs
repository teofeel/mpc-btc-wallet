using MediatR;
using Microsoft.EntityFrameworkCore;
using TransactionService.DB;
using TransactionService.DTOs;
using TransactionService.Features.Queries;
using TransactionService.Mappers;
using TransactionService.Models;

namespace TransactionService.Features.Handlers
{
    public class GetTransactionByWalletIdHandler : IRequestHandler<GetTransactionsByWalletIdQuery, List<TransactionResponseDTO>>
    {
        private readonly TransactionDbContext _dbContext;
        private readonly TransactionResponseMapper _mapper;

        public GetTransactionByWalletIdHandler(TransactionDbContext dbContext, TransactionResponseMapper mapper)
        {
            _dbContext = dbContext;
            _mapper = mapper;
        }

        public async Task<List<TransactionResponseDTO>> Handle(GetTransactionsByWalletIdQuery request, CancellationToken ct)
        {
            List<Transaction> transactions = await _dbContext.Transactions.AsNoTracking()
                .Where(t => t.WalletId == request.WalletId)
                .OrderByDescending(t => t.CreatedAt)
                .Skip((request.Page - 1) * request.PageSize)
                .Take(request.PageSize)
                .ToListAsync(ct);

            return transactions.Select(_mapper.ToDto).ToList();
        }
    }
}
