using MediatR;
using TransactionService.DTOs;

namespace TransactionService.Features.Commands;

public record MakeTransactionCommand(TransactionRequestDTO Dto) : IRequest<TransactionResponseDTO>;