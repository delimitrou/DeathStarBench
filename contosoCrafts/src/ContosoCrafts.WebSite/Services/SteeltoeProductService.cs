using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using ContosoCrafts.WebSite.Models;
using Microsoft.Extensions.Logging;

namespace ContosoCrafts.WebSite.Services
{
    public class SteeltoeProductService : IProductService
    {
        private const string PRODUCTS_API_SERVICE_NAME = "contoso-productsapi";
        private readonly string PRODUCTS_REQUEST_URI = $"http://{PRODUCTS_API_SERVICE_NAME}/products";
        private readonly HttpClient _httpClient;
        private readonly ILogger<SteeltoeProductService> _logger;

        public SteeltoeProductService(HttpClient httpClient, ILogger<SteeltoeProductService> logger)
        {
            _logger = logger;
            _httpClient = httpClient;
        }
        public async Task AddRating(string productId, int rating)
        {
            var payload = JsonSerializer.Serialize(new { productId, rating });
            var content = new StringContent(payload, Encoding.UTF8, "application/json");

            var resp = await _httpClient.PatchAsync(PRODUCTS_REQUEST_URI, content);
        }

        public async Task<IEnumerable<Product>> GetProducts()
        {
            var resp = await _httpClient.GetAsync(PRODUCTS_REQUEST_URI);
            if (!resp.IsSuccessStatusCode) //TODO: Successful if service not registered
            {
                // probably log some stuff here
                return Enumerable.Empty<Product>();
            }
            var contentStream = await resp.Content.ReadAsStreamAsync();
            var products = await JsonSerializer.DeserializeAsync<IEnumerable<Product>>(contentStream,
                           new JsonSerializerOptions { PropertyNameCaseInsensitive = true });

            return products;
        }
    }
}