namespace WalletService.DTOs
{
    public class CreateWalletResponseDTO
    {
        public Guid Id { get; set; }
        public string? Name { get; set; }
        public string? Network { get; set; }
        public string? Address { get; set; }
    }
}
