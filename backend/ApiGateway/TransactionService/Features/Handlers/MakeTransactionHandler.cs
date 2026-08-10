using MediatR;
using TransactionService.DTOs;
using TransactionService.Features.Commands;
using TransactionService.Services.Interfaces;

namespace TransactionService.Features.Handlers;

public class MakeTransactionHandler : IRequestHandler<MakeTransactionCommand, TransactionResponseDTO>
{
    private readonly ITransactionService _transactionService;
    
    public MakeTransactionHandler(ITransactionService transactionService)
    {
        _transactionService = transactionService;
    }

    public async Task<TransactionResponseDTO> Handle(MakeTransactionCommand request, CancellationToken ct)
    {
        return await _transactionService.MakeTransactionAsync(request.Dto);
    }
}