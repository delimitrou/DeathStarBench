using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using ContosoCrafts.ProductsApi.Models;
using MongoDB.Bson;
using MongoDB.Driver;

namespace ContosoCrafts.ProductsApi.Services
{
    public class ProductService : IProductService
    {
        private readonly IMongoClient _mongo;
        private readonly IMongoDatabase _database;
        private const string COLLECTION_NAME = "products";
        private const string DATABASE_NAME = "contosocrafts";

        public ProductService(IMongoClient mongo)
        {
            this._mongo = mongo;
            this._database = this._mongo.GetDatabase(DATABASE_NAME);
        }
        public async Task AddRating(string productId, int rating)
        {
            var collection = this._database.GetCollection<Product>(COLLECTION_NAME);
            var filter = Builders<Product>.Filter.Eq(x => x.ProductId, productId);
            var update = Builders<Product>.Update.Push(x => x.Ratings, rating);
            var result = await collection.UpdateOneAsync(filter, update);
        }

        public async Task<IEnumerable<Product>> GetProducts(int page = 1, int limit = 20)
        {
            var collection = this._database.GetCollection<Product>(COLLECTION_NAME);
            var results = await collection.Find(new BsonDocument())
                            .Skip(Convert.ToInt32((page - 1) * limit)).Limit(Convert.ToInt32(limit))
                            .ToListAsync();
            return results;
        }

        public async Task<Product> GetSingle(string id)
        {
            var collection = this._database.GetCollection<Product>(COLLECTION_NAME);
            var filter = Builders<Product>.Filter.Eq(x => x.ProductId, id);
            var cursor = await collection.FindAsync(filter);
            return cursor.SingleOrDefault();
        }
    }
}