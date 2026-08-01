using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using Ocelot.DependencyInjection;
using Ocelot.Middleware;
using Ocelot.Provider.Eureka;
using Steeltoe.Discovery.Client;
using System.Text;

var builder = WebApplication.CreateBuilder(args);

// Add ocelot to config
builder.Configuration.AddJsonFile("ocelot.json", optional: false, reloadOnChange: true);

// Add Steeltoe Eureka Discovery client
builder.Services.AddDiscoveryClient(builder.Configuration);

// Configure JWT for ocelot
String authenticationScheme = "Bearer";
var jwtSettings = builder.Configuration.GetSection("Jwt");
var key = Encoding.UTF8.GetBytes(jwtSettings["SecretKey"] ?? "YOUR_SUPER_SECRET_KEY_HERE_12345");

builder.Services.AddAuthentication(options =>
{
    options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
    options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
})
.AddJwtBearer(authenticationScheme, options =>
{
    options.RequireHttpsMetadata = false;
    options.SaveToken = true;
    options.TokenValidationParameters = new TokenValidationParameters
    {
        ValidateIssuerSigningKey = false,
        IssuerSigningKey = new SymmetricSecurityKey(key),
        ValidateIssuer = true,
        ValidIssuer = jwtSettings["Issuer"] ?? "http://localhost:8080",
        ValidateAudience = true,
        ValidAudience = jwtSettings["Audience"] ?? "bitcoin-wallet-users",
        ValidateLifetime = true,
        ClockSkew = TimeSpan.Zero
    };
});

// Attach eureka 
builder.Services
    .AddOcelot(builder.Configuration)
    .AddEureka();

// Add services to the container.
builder.Services.AddSwaggerGen();

var app = builder.Build();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}


app.UseHttpsRedirection();

app.UseAuthentication();
app.UseAuthorization();

await app.UseOcelot();

app.Run();
