using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace WalletService.Exceptions
{
    public class GlobalExceptionHandler : IExceptionHandler
    {
        private readonly ILogger<GlobalExceptionHandler> _logger;

        public GlobalExceptionHandler(ILogger<GlobalExceptionHandler> logger)
        {
            _logger = logger;
        }

        public async ValueTask<bool> TryHandleAsync(HttpContext httpContext, Exception exception, CancellationToken cancellationToken)
        {
            var (statusCode, title) = exception switch
            {
                ArgumentException => (StatusCodes.Status400BadRequest, exception.Message),
                KeyNotFoundException => (StatusCodes.Status404NotFound, exception.Message),
                ResourceNotFoundException => (StatusCodes.Status404NotFound, exception.Message),
                _ => (StatusCodes.Status500InternalServerError, "An unexpected error occurred")
            };

            _logger.LogError("[Wallet Service] Unhandled exception: {Message}", exception.Message);

            httpContext.Response.StatusCode = statusCode;
            await httpContext.Response.WriteAsJsonAsync(new ProblemDetails
            {
                Status = statusCode,
                Title = title,
                Instance = httpContext.Request.Path
            }, cancellationToken);

            return true;
        }
    }
}
