using Microsoft.EntityFrameworkCore;
using TransactionService.Models;

namespace TransactionService.DB
{
    public class TransactionDbContext : DbContext
    {
        public TransactionDbContext(DbContextOptions<TransactionDbContext> options) : base(options) { }

        public DbSet<Transaction> Transactions => Set<Transaction>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);

            modelBuilder.Entity<Transaction>(entity =>
            {
                entity.HasKey(t => t.Id);

                entity.Property(t => t.WalletId)
                .IsRequired();

                entity.Property(t => t.ReceivingAddress)
                .IsRequired()
                .HasMaxLength(90);

                entity.Property(t => t.SendingAddress)
                .HasMaxLength(90);

                entity.Property(t => t.TxHash)
                .HasMaxLength(64);

                entity.Property(t => t.Amount)
                .HasPrecision(18, 8)
                .IsRequired();

                entity.Property(t => t.Fee)
                .HasPrecision(18, 8)
                .IsRequired();

                entity.Property(t => t.Status)
                .HasConversion<string>()
                .HasMaxLength(20);

                entity.HasIndex(t => t.TxHash);
                entity.HasIndex(t => t.Status);
            });
        }

    }
}
