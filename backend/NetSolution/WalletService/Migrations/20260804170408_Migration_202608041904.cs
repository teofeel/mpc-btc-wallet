using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace WalletService.Migrations
{
    /// <inheritdoc />
    public partial class Migration_202608041904 : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "EncryptedMnemonic",
                table: "Wallets",
                newName: "MpcJobId");

            migrationBuilder.AddColumn<int>(
                name: "Status",
                table: "Wallets",
                type: "integer",
                nullable: false,
                defaultValue: 0);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "Status",
                table: "Wallets");

            migrationBuilder.RenameColumn(
                name: "MpcJobId",
                table: "Wallets",
                newName: "EncryptedMnemonic");
        }
    }
}
