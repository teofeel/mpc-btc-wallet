using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace WalletService.Migrations
{
    /// <inheritdoc />
    public partial class InitialCreate : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Wallets",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    OwnerId = table.Column<Guid>(type: "uuid", nullable: false),
                    Name = table.Column<string>(type: "text", nullable: false),
                    Network = table.Column<string>(type: "text", nullable: false),
                    DateCreated = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    DerivationPath = table.Column<string>(type: "text", nullable: false),
                    ExtendedPublicKey = table.Column<string>(type: "text", nullable: false),
                    EncryptedMnemonic = table.Column<string>(type: "text", nullable: true),
                    NextReceiveIndex = table.Column<int>(type: "integer", nullable: false),
                    NextChangeIndex = table.Column<int>(type: "integer", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Wallets", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "WalletAddresses",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    WalletId = table.Column<Guid>(type: "uuid", nullable: false),
                    Address = table.Column<string>(type: "text", nullable: false),
                    AddressIndex = table.Column<int>(type: "integer", nullable: false),
                    IsChange = table.Column<bool>(type: "boolean", nullable: false),
                    BalanceSatoshis = table.Column<long>(type: "bigint", nullable: false),
                    DateCreated = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_WalletAddresses", x => x.Id);
                    table.ForeignKey(
                        name: "FK_WalletAddresses_Wallets_WalletId",
                        column: x => x.WalletId,
                        principalTable: "Wallets",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateIndex(
                name: "IX_WalletAddresses_WalletId_AddressIndex_IsChange",
                table: "WalletAddresses",
                columns: new[] { "WalletId", "AddressIndex", "IsChange" },
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "WalletAddresses");

            migrationBuilder.DropTable(
                name: "Wallets");
        }
    }
}
