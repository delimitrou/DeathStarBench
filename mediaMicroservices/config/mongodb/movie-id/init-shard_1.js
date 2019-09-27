rs.initiate(
{
  "_id": "movie-id-mongodb-shard-1",
  "version": 1,
  "members": [
    {
      "_id": "movie-id-mongodb-shard-1_0",
      "host": "movie-id-mongodb-shard-1_0:27017"
    }
  ]
}
)