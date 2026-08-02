using MediatR;
using Microsoft.AspNetCore.Mvc;
using WalletService.Controllers.Interfaces;
using WalletService.DTOs;
using WalletService.Features.Commands;

namespace WalletService.Controllers
{
    public class WalletController : IController
    {
        private readonly IMediator _mediator;
        private readonly ILogger<WalletController> _logger;

        public WalletController(IMediator mediator, ILogger<WalletController> logger)
        {
            _mediator = mediator;
            _logger = logger;
        }

        [HttpGet("{id}")]
        public async Task<IActionResult> GetWallet()
        {
            return StatusCode(StatusCodes.Status501NotImplemented);
        }


        [HttpGet]
        public async Task<IActionResult> GetAllWallets()
        {
            return StatusCode(StatusCodes.Status501NotImplemented);
        }


        [HttpPost]
        [ProducesResponseType(typeof(CreateWalletResponseDTO), StatusCodes.Status201Created)]
        [ProducesResponseType(StatusCodes.Status400BadRequest)]
        [ProducesResponseType(StatusCodes.Status500InternalServerError)]
        public async Task<IActionResult> CreateWallet([FromBody] CreateWalletRequestDTO request)
        {
            if (!ModelState.IsValid)
            {
                return BadRequest(ModelState);
            }

            _logger.LogInformation("[Wallet Service] Received HTTP POST request to create wallet: {Name} on network: {Network}", request.Name, request.Network);

            CreateWalletCommand command = new CreateWalletCommand(request);
            CreateWalletResponseDTO response = await _mediator.Send(command);

            return StatusCode(StatusCodes.Status201Created, response);
        }
    }
}
