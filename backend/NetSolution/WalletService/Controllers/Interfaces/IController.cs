using Microsoft.AspNetCore.Mvc;

namespace WalletService.Controllers.Interfaces
{

    [ApiController]
    [Route("api/v1/wallets")]
    public abstract class IController : ControllerBase
    {
    }
}
