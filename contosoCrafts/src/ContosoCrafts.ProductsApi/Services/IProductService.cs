using System.Collections.Generic;
using System.Threading.Tasks;
using ContosoCrafts.ProductsApi.Models;

namespace ContosoCrafts.ProductsApi.Services
{
    public interface IProductService
    {
        Task<IEnumerable<Product>> GetProducts(int page = 1, int limit = 20);
        Task AddRating(string productId, int rating);
        Task<Product> GetSingle(string id);
    }
}