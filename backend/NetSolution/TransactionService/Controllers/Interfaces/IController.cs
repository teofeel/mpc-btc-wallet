using Microsoft.AspNetCore.Mvc;

namespace TransactionService.Controllers.Interfaces
{
    [ApiController]
    [Route("api/v1/transactions")]
    public abstract class IController : ControllerBase
    {
    }
}
