using Microsoft.Extensions.ObjectPool;
using RabbitMQ.Client;

namespace ContosoCrafts.CheckoutProcessor.Services
{
    public class RabbitModelPooledObjectPolicy : IPooledObjectPolicy<IModel>
    {
        private readonly IConnection _connection;

        public RabbitModelPooledObjectPolicy(IConnectionFactory connectionFactory)
        {
            _connection = connectionFactory.CreateConnection(Constants.APPLICATION_NAME);
        }

        public IModel Create() => _connection.CreateModel();

        public bool Return(IModel obj)
        {
            if (obj.IsOpen)
            {
                return true;
            }
            else
            {
                obj?.Close();
                obj?.Dispose();
                return false;
            }
        }
    }
}