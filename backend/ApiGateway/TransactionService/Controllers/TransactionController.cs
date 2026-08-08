using MediatR;
using Microsoft.AspNetCore.Components.Forms;
using Microsoft.AspNetCore.Mvc;
using Proto;
using TransactionService.Controllers.Interfaces;
using TransactionService.DTOs;
using TransactionService.Features.Queries;
using TransactionService.Models.Enums;
using TransactionService.Services.Interfaces;

namespace TransactionService.Controllers
{
    public class TransactionController : IController
    {
        private readonly ISender _mediator;
        private readonly ITransactionService _transactionService;

        public TransactionController(ISender mediator,  ITransactionService transactionService)
        {
            _transactionService = transactionService;
            _mediator = mediator;
        }

        [HttpGet("{id:guid}")]
        [ProducesResponseType(StatusCodes.Status200OK, Type = typeof(TransactionResponseDTO))]
        [ProducesResponseType(StatusCodes.Status404NotFound)]
        [ProducesResponseType(StatusCodes.Status500InternalServerError)]
        public async Task<IActionResult> GetTransactionById(Guid id, CancellationToken ct)
        {
            TransactionResponseDTO transaction = await _mediator.Send(new GetTransactionQuery(id), ct);

            return Ok(transaction);
        }

        [HttpGet("wallet/{walletId:guid}")]
        [ProducesResponseType(StatusCodes.Status200OK, Type = typeof(List<TransactionResponseDTO>))]
        [ProducesResponseType(StatusCodes.Status500InternalServerError)]
        public async Task<IActionResult> GetTransactionsByWalletId(
            [FromRoute] Guid walletId, 
            [FromQuery] int page = 1, 
            [FromQuery] int pageSize = 50, 
            CancellationToken ct = default)
        {
            List<TransactionResponseDTO> transactions = await _mediator.Send(new GetTransactionsByWalletIdQuery(walletId, page, pageSize), ct);

            return Ok(transactions);
        }

        [HttpGet("fee-estimation")]
        [ProducesResponseType(StatusCodes.Status200OK, Type = typeof(FeeEstimationResponseDto))]
        [ProducesResponseType(StatusCodes.Status500InternalServerError)]
        public async Task<IActionResult> GetFeeEstimation([FromQuery] FeePriority priority = FeePriority.MEDIUM)
        {
            FeeEstimationResponseDTO response = await _transactionService.GetFeeEstimationAsync(priority);

            return Ok(response);
        }
    }
}

