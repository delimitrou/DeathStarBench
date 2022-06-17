using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using ContosoCrafts.WebSite.Events;
using ContosoCrafts.WebSite.Models;
using ContosoCrafts.WebSite.Services;
using EventAggregator.Blazor;
using Microsoft.AspNetCore.Components;
using Microsoft.Extensions.Caching.Distributed;

namespace ContosoCrafts.WebSite.Components
{
    public class ProductListBase : ComponentBase
    {
        [Inject]
        protected IProductService ProductService { get; set; }

        [Inject]
        private IEventAggregator EventAggregator { get; set; }

        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        [Inject]
        private IDistributedCache Cache { get; set; }
        
        protected IEnumerable<Product> products = null;
        protected Product selectedProduct;
        protected string selectedProductId;

        protected override async Task OnInitializedAsync()
        {
            if (products == null)
                products = await ProductService.GetProducts();
        }
        protected async Task SelectProduct(string productId)
        {
            selectedProductId = productId;
            selectedProduct = (await ProductService.GetProducts()).First(x => x.Id == productId);
        }

        protected async Task SubmitRating(int rating)
        {
            await ProductService.AddRating(selectedProductId, rating);
            await SelectProduct(selectedProductId);
            StateHasChanged();
        }

        protected async Task AddToCart(string productId, string title)
        {            
            Dictionary<string, CartItem> state = null;
            
            // Check for exisiting cart data
            var cartData = await Cache.GetStringAsync(Constants.CART_CACHE_KEY);
            if (cartData == null)
            {
                // Empty cart
                state = new Dictionary<string, CartItem> { [productId] = new CartItem { Title = title, Quantity = 1 } };
            }
            else
            {
                state = JsonSerializer.Deserialize<Dictionary<string, CartItem>>(cartData);
                if (state.ContainsKey(productId))
                {
                    // Product already in cart
                    CartItem selectedItem = state[productId];
                    selectedItem.Quantity++;
                    state[productId] = selectedItem;
                }
                else
                {
                    // Add product to car
                    state[productId] = new CartItem { Title = title, Quantity = 1 };
                }
            }

            // Persist new state
            var payload = JsonSerializer.Serialize(state);

            await Cache.SetStringAsync(Constants.CART_CACHE_KEY, payload);
            await EventAggregator.PublishAsync(new ShoppingCartUpdated { ItemCount = state.Keys.Count });
        }
    }
}