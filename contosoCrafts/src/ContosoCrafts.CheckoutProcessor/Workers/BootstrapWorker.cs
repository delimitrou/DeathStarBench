using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Steeltoe.Discovery;

namespace ContosoCrafts.CheckoutProcessor.Workers
{
    public class BootstrapWorker : BackgroundService
    {
        private readonly ILogger<BootstrapWorker> _logger;
        private readonly IHostApplicationLifetime _applicationLifetime;
        private readonly IDiscoveryClient _dicoveryClient;

        public BootstrapWorker(IDiscoveryClient dicoveryClient, IHostApplicationLifetime applicationLifetime, ILogger<BootstrapWorker> logger)
        {
            _dicoveryClient = dicoveryClient;
            _applicationLifetime = applicationLifetime;
            _logger = logger;
        }
        protected override Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("{ApplicationName} registered", "CheckoutProcessor");
            return Task.CompletedTask;
        }
    }
}