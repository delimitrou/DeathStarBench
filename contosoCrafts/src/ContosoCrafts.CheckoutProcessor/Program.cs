using System;
using System.IO;
using ContosoCrafts.CheckoutProcessor.Services;
using ContosoCrafts.CheckoutProcessor.Workers;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.ObjectPool;
using RabbitMQ.Client;
using Serilog;
using Steeltoe.Discovery.Client;
using Steeltoe.Management.Tracing;

namespace ContosoCrafts.CheckoutProcessor
{
    public class Program
    {
        public static IConfiguration Configuration { get; } = new ConfigurationBuilder()
            .SetBasePath(Directory.GetCurrentDirectory())
            .AddJsonFile("appsettings.json", optional: false, reloadOnChange: true)
            .AddJsonFile($"appsettings.{Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT") ?? "Production"}.json", optional: true)
            .AddEnvironmentVariables()
            .Build();
        public static int Main(string[] args)
        {
            Log.Logger = new LoggerConfiguration()
                .ReadFrom.Configuration(Configuration, "Serilog")
                .CreateLogger();

            try
            {
                Log.ForContext<Program>().Information("Starting host");
                CreateHostBuilder(args).Build().Run();
                return 0;
            }
            catch (Exception ex)
            {
                Log.Fatal(ex, "Host terminated unexpectedly");
                return 1;
            }
            finally
            {
                Log.Information("Host stopped");
                Log.CloseAndFlush();
            }
        }

        public static IHostBuilder CreateHostBuilder(string[] args) =>
            Host.CreateDefaultBuilder(args)
                .UseSerilog()
                .AddServiceDiscovery()
                .ConfigureServices((hostContext, services) =>
                {
                    services.TryAddSingleton<ObjectPoolProvider, DefaultObjectPoolProvider>();

                    // RabbitMQ services
                    services.AddSingleton<IConnectionFactory, ConnectionFactory>(provider =>
                    {
                        return new ConnectionFactory
                        {
                            VirtualHost = Constants.RABBITMQ_VHOST,
                            HostName = "rabbitmq_service",
                            UserName = "demo",
                            Password = "demo",
                            DispatchConsumersAsync = true
                        };
                    });

                    services.AddSingleton<ObjectPool<IModel>>(provider =>
                    {
                        var poolProvider = provider.GetRequiredService<ObjectPoolProvider>();
                        var cf = provider.GetRequiredService<IConnectionFactory>();
                        var policy = new RabbitModelPooledObjectPolicy(cf);

                        return poolProvider.Create(policy);
                    });

                    services.AddDistributedTracing(Configuration, 
                        builder => builder.UseZipkinWithTraceOptions(services));

                    // Worker services
                    services.AddHostedService<BootstrapWorker>();
                    services.AddHostedService<ProcessorWorker>();
                });
    }
}
