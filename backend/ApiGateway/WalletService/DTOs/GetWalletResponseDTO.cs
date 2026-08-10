namespace WalletService.DTOs
{
    public class GetWalletResponseDTO
    {
        public Guid WalletId { get; set; }
        public string Name { get; set; } = string.Empty;
        public string Network { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public string DerivationPath { get; set; } = string.Empty;
        public string ExtendedPublicKey { get; set; } = string.Empty;
        public List<WalletAddressInfoDTO> Addresses { get; set; } = new List<WalletAddressInfoDTO>();
    }
}
