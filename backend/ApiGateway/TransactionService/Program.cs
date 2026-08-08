using Microsoft.EntityFrameworkCore;
using NBitcoin;
using Steeltoe.Discovery.Eureka;
using TransactionService.BgServices;
using TransactionService.DB;
using TransactionService.Exceptions;
using TransactionService.Mappers;
using TransactionService.Services;
using TransactionService.Services.Interfaces;

var builder = WebApplication.CreateBuilder(args);


// Add Steeltoe Eureka Discovery client
builder.Services.AddEurekaDiscoveryClient();

// Add services to the container.
builder.Services.AddControllers();

// Global exception handling
builder.Services.AddExceptionHandler<GlobalExceptionHandler>();
builder.Services.AddProblemDetails();


// Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

if (builder.Environment.IsDevelopment())
{
    builder.Configuration.AddUserSecrets<Program>();
}

// MediatR
builder.Services.AddMediatR(cfg =>
    cfg.RegisterServicesFromAssembly(typeof(Program).Assembly));

// DB context
builder.Services.AddDbContext<TransactionDbContext>(options =>
    options.UseNpgsql(builder.Configuration.GetConnectionString("TransactionDb")));

// Caching (fee estimation)
builder.Services.AddMemoryCache();

// Bitcoin network (dev stack runs against bitcoind regtest)
builder.Services.AddSingleton(
    Network.GetNetwork(builder.Configuration["BitcoinNetwork"] ?? "regtest") ?? Network.RegTest);

// Mappers
builder.Services.AddScoped<TransactionResponseMapper>();

// Domain
builder.Services.AddScoped<ITransactionService, TransactionService.Services.TransactionService>();

// gRPC clients (both services are served by the BlockchainIndexingService)
builder.Services.AddGrpcClient<Proto.ChainStateService.ChainStateServiceClient>(options =>
    options.Address = new Uri(builder.Configuration["IndexerService:Address"] ?? "http://localhost:50052"));
builder.Services.AddGrpcClient<Proto.TransactionService.TransactionServiceClient>(options =>
    options.Address = new Uri(builder.Configuration["IndexerService:Address"] ?? "http://localhost:50052"));

// Background services
builder.Services.AddHostedService<TransactionConfirmationBgService>();

var app = builder.Build();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseExceptionHandler();

app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();

app.Run();