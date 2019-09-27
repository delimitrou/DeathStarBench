rs.initiate(
{
  "_id": "movie-id-mongodb-config",
  "configsvr": "true",
  "version": 1,
  "members": [
    {
      "_id": "movie-id-mongodb-config",
      "host": "movie-id-mongodb-config:27017"
    }
  ]
}
)