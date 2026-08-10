using WalletService.Models.Enums;

namespace WalletService.DTOs
{
    public class StartWalletCreationResponseDTO
    {
        public Guid Id { get; set; }
        public string? Name { get; set; }
        public string? Network { get; set; }
        public string? JobId { get; set; }
    }
}
