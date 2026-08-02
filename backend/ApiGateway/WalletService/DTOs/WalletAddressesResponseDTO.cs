namespace WalletService.DTOs
{
    public class WalletAddressesResponseDTO
    {
        public string Name { get; set; }
        public string Network { get; set; }
        public string Address { get; set; }
        public long BalanceSatoshis { get; set; }
        public DateTime DateCreated { get; set; }
    }
}
