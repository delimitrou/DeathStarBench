using System;
using System.Threading.Channels;
using System.Threading.Tasks;
using RabbitMQ.Client;

namespace ContosoCrafts.CheckoutProcessor.Services
{
    public class AsyncChannelConsumer : AsyncDefaultBasicConsumer
    {
        private const int MAXIMUM_CHANNEL_MESSAGES = 1000;
        private readonly Channel<RabbitmMQMessage> _channel;
        public ChannelReader<RabbitmMQMessage> Reader => _channel.Reader;

        public AsyncChannelConsumer(IModel model) : base(model)
        {
            var options = new BoundedChannelOptions(MAXIMUM_CHANNEL_MESSAGES)
            {
                SingleWriter = true,
                SingleReader = false
            };

            _channel = Channel.CreateBounded<RabbitmMQMessage>(options);
        }

        public override async Task HandleBasicDeliver(
            string consumerTag,
            ulong deliveryTag, bool redelivered,
            string exchange, string routingKey,
            IBasicProperties properties,
            ReadOnlyMemory<byte> body)
        {
            var message = new RabbitmMQMessage(consumerTag, deliveryTag, redelivered, exchange, routingKey, properties, body);
            await _channel.Writer.WaitToWriteAsync();

            int retryCounter = 0;
            while (!_channel.Writer.TryWrite(message))
            {
                retryCounter++;
                if (retryCounter > 3) break;
            }
        
            if (retryCounter > 3) {
                //TODO: logg something
                //properties.MessageId ??
            }                        
        }
    }
}