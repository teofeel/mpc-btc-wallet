using Microsoft.EntityFrameworkCore;
using WalletService.Models;
namespace WalletService.DB
{
    public class WalletDbContext : DbContext
    {
        public WalletDbContext(DbContextOptions<WalletDbContext> options) : base(options) { }

        public DbSet<Wallet> Wallets => Set<Wallet>();
        public DbSet<WalletAddress> WalletAddresses => Set<WalletAddress>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            modelBuilder.Entity<Wallet>()
                .HasMany(w => w.Addresses)
                .WithOne(a => a.Wallet)
                .HasForeignKey(a => a.WalletId);

            modelBuilder.Entity<WalletAddress>()
                .HasIndex(a => new { a.WalletId, a.AddressIndex, a.IsChange })
                .IsUnique(); 

            base.OnModelCreating(modelBuilder);
        }
    }
}
