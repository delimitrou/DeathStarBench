using ContosoCrafts.WebSite.Services;
using EventAggregator.Blazor;
using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.ObjectPool;
using RabbitMQ.Client;
using Serilog;
using StackExchange.Redis;
using Steeltoe.Common.Http.Discovery;
using Steeltoe.Management.Tracing;

namespace ContosoCrafts.WebSite
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        public void ConfigureServices(IServiceCollection services)
        {
            services.AddRazorPages();
            services.AddServerSideBlazor();

            services.AddHttpClient("discovery",
                c =>
                {
                    c.DefaultRequestHeaders.Add("Accept", "application/json");
                    c.DefaultRequestHeaders.Add("User-Agent", "contoso-app");
                })
                .AddHttpMessageHandler<DiscoveryHttpMessageHandler>()
                .AddTypedClient<IProductService, SteeltoeProductService>();

            services.AddStackExchangeRedisCache(options =>
            {
                options.ConfigurationOptions = new ConfigurationOptions
                {
                    Password = "S0m3P@$$w0rd",
                    EndPoints = { "redis_service:6379" }
                };

                // This allows partitioning a single backend cache for use with multiple apps/services.
                options.InstanceName = "ContosoRedis";
            });

            services.TryAddSingleton<ObjectPoolProvider, DefaultObjectPoolProvider>();

            services.AddSingleton<IConnectionFactory, ConnectionFactory>(provider =>
            {
                return new ConnectionFactory
                {
                    VirtualHost = Constants.RABBITMQ_VHOST,
                    HostName = "rabbitmq_service",
                    UserName = "demo",
                    Password = "demo"
                };
            });

            services.AddSingleton<ObjectPool<IModel>>(provider =>
            {
                var poolProvider = provider.GetRequiredService<ObjectPoolProvider>();
                var cf = provider.GetRequiredService<IConnectionFactory>();
                var policy = new RabbitModelPooledObjectPolicy(cf);

                return poolProvider.Create(policy);
            });

            services.AddTransient<RabbitMQBus>();

            services.AddHealthChecks();
            services.AddControllers();
            services.AddScoped<IEventAggregator, EventAggregator.Blazor.EventAggregator>();
            
            services.AddDistributedTracing(Configuration, builder => builder.UseZipkinWithTraceOptions(services));
        }

        public void Configure(IApplicationBuilder app)
        {
            app.UseExceptionHandler("/Error");

            app.UseHsts();
            app.UseStaticFiles();

            app.UseSerilogRequestLogging();
            app.UseRouting();
            app.UseEndpoints(endpoints =>
            {
                endpoints.MapHealthChecks("/health");
                endpoints.MapRazorPages();
                endpoints.MapControllers();
                endpoints.MapBlazorHub();
            });
        }
    }
}
