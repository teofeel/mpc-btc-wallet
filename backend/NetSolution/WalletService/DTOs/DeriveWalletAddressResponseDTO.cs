namespace WalletService.DTOs
{
    public class DeriveWalletAddressResponseDTO
    {
        public string Address { get; set; }
        public long BalanceSatoshis { get; set; }
        public DateTime DateCreated { get; set; }
    }
}
