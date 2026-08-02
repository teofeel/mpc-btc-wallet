using Microsoft.EntityFrameworkCore;
using Steeltoe.Discovery.Eureka;
using WalletService.DB;
using WalletService.DTOs;
using WalletService.Exceptions;
using WalletService.Mappers;
using WalletService.Mappers.Interfaces;
using WalletService.Models;
using WalletService.Services;
using WalletService.Services.Interfaces;

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
builder.Services.AddScoped<IMapper<WalletAddress, WalletAddressesResponseDTO>, WalletAddressResponseMapper>();

// Domain
builder.Services.AddScoped<IWalletCreationService, BitcoinWalletCreationService>();

// Exception handler
builder.Services.AddExceptionHandler<GlobalExceptionHandler>();
builder.Services.AddProblemDetails();

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
