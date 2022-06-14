using System;
using System.Net.Mime;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.ObjectPool;
using RabbitMQ.Client;

namespace ContosoCrafts.WebSite.Services
{
    public class RabbitMQBus
    {
        private const string CHECKOUT_QUEUE_NAME = "contoso_steeltoe_checkout";
        private const string CHECKOUT_EXCHANGE_NAME = "contoso_steeltoe_cart";
        private const string CHECKOUT_ROUTING_KEY = "cart_checkout";
        private readonly ObjectPool<IModel> _rabbitBuilderPool;
        public RabbitMQBus(ObjectPool<IModel> rabbitBuilderPool)
        {
            _rabbitBuilderPool = rabbitBuilderPool;
            InitializeQueue();
        }

        protected virtual void InitializeQueue()
        {
            IModel channel = _rabbitBuilderPool.Get();

            channel.QueueDeclare(queue: CHECKOUT_QUEUE_NAME, durable: true,
                                 exclusive: false, autoDelete: false, arguments: null);

            channel.ExchangeDeclare(exchange: CHECKOUT_EXCHANGE_NAME, type: ExchangeType.Direct);

            channel.QueueBind(CHECKOUT_QUEUE_NAME, CHECKOUT_EXCHANGE_NAME, CHECKOUT_ROUTING_KEY, null);

            _rabbitBuilderPool.Return(channel);
        }

        public virtual void Publish<T>(T payload)
        {
            var json_payload = JsonSerializer.Serialize(payload);

            IModel channel = _rabbitBuilderPool.Get();

            var properties = channel.CreateBasicProperties();
            properties.Persistent = true;
            properties.ContentType = MediaTypeNames.Application.Json;
            properties.MessageId = Guid.NewGuid().ToString("N");

            channel.BasicPublish(exchange: CHECKOUT_EXCHANGE_NAME,
                                 routingKey: CHECKOUT_ROUTING_KEY,
                                 mandatory: false, //TODO: huh??
                                 basicProperties: properties,
                                 body: Encoding.UTF8.GetBytes(json_payload));

            _rabbitBuilderPool.Return(channel);
        }
    }
}