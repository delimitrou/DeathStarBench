using System.Text.Json.Serialization;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace ContosoCrafts.ProductsApi.Models
{
    public class Product
    {
        [BsonId]
        [BsonElement("_id")]
        [JsonIgnore]
        [BsonRepresentation(BsonType.ObjectId)]
        public string RecId { get; set; }

        [BsonRepresentation(BsonType.String)]
        [BsonElement("Id")]
        [JsonPropertyName("Id")]
        public string ProductId { get; set; }
        public string Maker { get; set; }
        public string Image { get; set; }
        public string Url { get; set; }
        public string Title { get; set; }
        public string Description { get; set; }
        public int[] Ratings { get; set; }
    }
}