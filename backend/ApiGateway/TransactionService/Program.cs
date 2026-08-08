using Steeltoe.Discovery.Eureka;

var builder = WebApplication.CreateBuilder(args);


// Add Steeltoe Eureka Discovery client
builder.Services.AddEurekaDiscoveryClient();

// Add services to the container.
builder.Services.AddControllers();

// Global exception handling
builder.Services.AddExceptionHandler<TransactionService.Exceptions.GlobalExceptionHandler>();
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
