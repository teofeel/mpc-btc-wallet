
namespace WalletService.Models
{
    public class Wallet
    {
        public Guid Id { get; set; } = Guid.NewGuid();

        public Guid OwnerId { get; set; } = Guid.Empty;

        public string Name { get; set; } = "Primary Wallet";
        public string Network { get; set; } = "Main";
        public DateTime DateCreated { get; set; } = DateTime.UtcNow;

        public string DerivationPath { get; set; } = "m/84'/0'/0'";
        public string ExtendedPublicKey { get; set; } = string.Empty;
        public string? EncryptedMnemonic { get; set; }

        public int NextReceiveIndex { get; set; } = 0;
        public int NextChangeIndex { get; set; } = 0;

        public ICollection<WalletAddress> Addresses { get; set; } = new List<WalletAddress>();
    }
}
