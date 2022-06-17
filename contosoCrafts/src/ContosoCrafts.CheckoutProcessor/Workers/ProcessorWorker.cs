using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using ContosoCrafts.CheckoutProcessor.Models;
using ContosoCrafts.CheckoutProcessor.Services;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.ObjectPool;
using RabbitMQ.Client;

namespace ContosoCrafts.CheckoutProcessor.Workers
{
    public class ProcessorWorker : BackgroundService
    {
        private const string CHECKOUT_QUEUE_NAME = "contoso_steeltoe_checkout";
        private const string CHECKOUT_EXCHANGE_NAME = "contoso_steeltoe_cart";
        private const string CHECKOUT_ROUTING_KEY = "cart_checkout";
        private readonly ILogger<ProcessorWorker> _logger;
        private readonly ObjectPool<IModel> _rabbitBuilderPool;

        public ProcessorWorker(ObjectPool<IModel> rabbitBuilderPool, ILogger<ProcessorWorker> logger)
        {
            _rabbitBuilderPool = rabbitBuilderPool;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("Worker running {WorkerName}", nameof(ProcessorWorker));

            InitializeQueue();
            await ProcessMessages(stoppingToken);

            _logger.LogInformation("Worker stopped {WorkerName}", nameof(ProcessorWorker));
        }

        private async Task ProcessMessages(CancellationToken stoppingToken)
        {
            IModel channel = _rabbitBuilderPool.Get();

            var consumer = new AsyncChannelConsumer(channel);
            channel.BasicConsume(CHECKOUT_QUEUE_NAME, false, consumer);

            await foreach (RabbitmMQMessage msg in consumer.Reader.ReadAllAsync(stoppingToken))
            {
                _logger.LogInformation("Message Received on the channel");

                var json_payload = Encoding.UTF8.GetString(msg.Body.ToArray());
                var cartItems = JsonSerializer.Deserialize<IEnumerable<CartItem>>(json_payload);

                _logger.LogInformation("Received {OrderItemCount} items", cartItems.Count());
                channel.BasicAck(msg.DeliveryTag, false);
            };

            if (stoppingToken.IsCancellationRequested) _logger.LogWarning("Operation cancelled {operation}", nameof(ProcessMessages));
            _logger.LogWarning("Unable to read channel");
        }

        private void InitializeQueue()
        {
            _logger.LogInformation($"Initializing queue:{CHECKOUT_QUEUE_NAME} and exchange:{CHECKOUT_EXCHANGE_NAME}");

            IModel channel = _rabbitBuilderPool.Get();
            channel.QueueDeclare(queue: CHECKOUT_QUEUE_NAME, durable: true,
                                 exclusive: false, autoDelete: false, arguments: null);

            channel.ExchangeDeclare(CHECKOUT_EXCHANGE_NAME, ExchangeType.Direct);
            channel.QueueBind(CHECKOUT_QUEUE_NAME, CHECKOUT_EXCHANGE_NAME, CHECKOUT_ROUTING_KEY, null);

            _rabbitBuilderPool.Return(channel);
        }
    }
}