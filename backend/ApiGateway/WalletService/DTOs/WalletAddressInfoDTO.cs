namespace WalletService.DTOs
{
    public class WalletAddressInfoDTO
    {
        public string Address { get; set; } = string.Empty;
        public int AddressIndex { get; set; }
        public bool IsChange { get; set; }
        public string PublicKeyHex { get; set; } = string.Empty;
    }
}
