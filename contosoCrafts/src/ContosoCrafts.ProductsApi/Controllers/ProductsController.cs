using System.Threading.Tasks;
using ContosoCrafts.ProductsApi.Services;
using Microsoft.AspNetCore.Mvc;

namespace ContosoCrafts.ProductsApi.Controllers
{
    [ApiController]
    [Route("[controller]")]
    public class ProductsController : ControllerBase
    {
        public ProductsController(IProductService productService)
        {
            _productService = productService;
        }

        private readonly IProductService _productService;

        [HttpGet]
        public async Task<ActionResult> GetList(int page = 1, int limit = 20)
        {
            var result = await _productService.GetProducts(page, limit);
            return Ok(result);
        }

        [HttpGet("{id}")]
        public async Task<ActionResult> GetSingle(string id)
        {
            var result = await _productService.GetSingle(id);
            return Ok(result);
        }

        [HttpPatch]
        public async Task<ActionResult> Patch(RatingRequest request)
        {
            await _productService.AddRating(request.ProductId, request.Rating);
            return Ok();
        }

        public class RatingRequest
        {
            public string ProductId { get; set; }
            public int Rating { get; set; }
        }
    }
}
