using Microsoft.EntityFrameworkCore;
using Steeltoe.Discovery.Eureka;
using WalletService.BgServices;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Mappers;
using WalletService.Mappers.Interfaces;
using WalletService.Models;
using WalletService.Services;
using NBitcoin;
using NBitcoin.RPC;
using WalletService.Services.Interfaces;
using WalletService.Services.MpcService;

var builder = WebApplication.CreateBuilder(args);

// Add Steeltoe Eureka Discovery client
builder.Services.AddEurekaDiscoveryClient();

// Add services to the container
builder.Services.AddControllers();

// Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// DB context
if (builder.Environment.IsDevelopment())
{
    builder.Configuration.AddUserSecrets<Program>();
}

builder.Services.AddDbContext<WalletDbContext>(options =>
    options.UseNpgsql(builder.Configuration.GetConnectionString("WalletDb")));

// MediatR
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssembly(typeof(Program).Assembly));

// Mappers
builder.Services.AddScoped<CreateWalletMapper>();
builder.Services.AddScoped<StartWalletCreationMapper>();
builder.Services.AddScoped<DeriveAddressMapper>();
builder.Services.AddScoped<IMapper<WalletAddress, WalletAddressesResponseDTO>, WalletAddressResponseMapper>();

// Domain
builder.Services.AddScoped<IWalletCreationService, BitcoinWalletCreationService>();
builder.Services.AddScoped<IWalletMpcService, BitcoinWalletMpcService>();

// Exception handler
builder.Services.AddExceptionHandler<GlobalExceptionHandler>();
builder.Services.AddProblemDetails();

// GRPC
builder.Services.AddGrpcClient<WalletMpcService.WalletMpcServiceClient>(options =>
{
    options.Address = new Uri(builder.Configuration["MpcService:Address"] ?? "http://localhost:50051");
});

// Bitcoin Core RPC client
var btcNetworkStr = builder.Configuration["BitcoinCore:Network"] ?? "regtest";
Network btcNetwork = btcNetworkStr.Trim().ToLowerInvariant() switch
{
    "mainnet" or "main" => Network.Main,
    "testnet" or "test" => Network.TestNet,
    "regtest" => Network.RegTest,
    _ => Network.RegTest
};

var rpcHost = builder.Configuration["BitcoinCore:Host"] ?? "http://127.0.0.1:18443";
var rpcUser = builder.Configuration["BitcoinCore:User"] ?? "user";
var rpcPass = builder.Configuration["BitcoinCore:Password"] ?? "pass";

var rpcClient = new RPCClient(new System.Net.NetworkCredential(rpcUser, rpcPass), new Uri(rpcHost), btcNetwork);

builder.Services.AddSingleton(rpcClient);
builder.Services.AddScoped<IBitcoinCoreService, BitcoinCoreService>();

// MPC completion queue + background services
builder.Services.AddSingleton<IMpcCompletionQueue, BitcoinWalletMpcCompletionQueue>();
builder.Services.AddHostedService<MpcJobMonitor>();
builder.Services.AddHostedService<MpcJobPoller>();


var app = builder.Build();

app.UseExceptionHandler();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();


app.Run();
