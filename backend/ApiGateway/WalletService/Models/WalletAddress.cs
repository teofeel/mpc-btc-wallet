namespace WalletService.Models
{
    public class WalletAddress
    {
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid WalletId { get; set; }
        public Wallet Wallet { get; set; } = null;
        public string Address { get; set; } = string.Empty;
        public int AddressIndex { get; set; }
        public bool IsChange { get; set; } = false;
        public long BalanceSatoshis { get; set; } = 0;
        public DateTime DateCreated { get; set; } = DateTime.UtcNow;
    }
}
