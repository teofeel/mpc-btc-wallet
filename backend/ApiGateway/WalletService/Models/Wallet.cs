
using WalletService.Models.Enums;

namespace WalletService.Models
{
    public class Wallet
    {
        public Guid Id { get; set; } = Guid.NewGuid();
        public Guid OwnerId { get; set; } = Guid.Empty;
        public string Name { get; set; } = "Primary Wallet";
        public string Network { get; set; } = "Main";
        public DateTime DateCreated { get; set; } = DateTime.UtcNow;
        public WalletStatus Status { get; set; } = WalletStatus.PENDING_KEY_GEN;
        public string? MpcJobId { get; set; }
        public string DerivationPath { get; set; } = "m/84'/0'/0'";
        public string ExtendedPublicKey { get; set; } = string.Empty;
        public int NextReceiveIndex { get; set; } = 0;
        public int NextChangeIndex { get; set; } = 0;

        public ICollection<WalletAddress> Addresses { get; set; } = new List<WalletAddress>();
    }
}
