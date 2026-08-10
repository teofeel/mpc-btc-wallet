using MediatR;
using Microsoft.EntityFrameworkCore;
using TransactionService.DB;
using TransactionService.Features.Queries;
using TransactionService.Models.Enums;

namespace TransactionService.Features.Handlers
{
    public class GetTransactionHashesHandler : IRequestHandler<GetTransactionHashesQuery, List<string>>
    {
        private readonly TransactionDbContext _dbContext;

        public GetTransactionHashesHandler(TransactionDbContext dbContext)
        {
            _dbContext = dbContext;
        }

        public async Task<List<string>> Handle(GetTransactionHashesQuery request, CancellationToken cancellationToken)
        {
            return await _dbContext.Transactions
                .AsNoTracking()
                .Where(t => t.Status == request.Status && !string.IsNullOrEmpty(t.TxHash))
                .Select(t => t.TxHash!)
                .ToListAsync(cancellationToken);
        }
    }
}
