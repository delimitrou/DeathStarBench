rs.initiate(
{
  "_id": "user-mongodb-config",
  "configsvr": true,
  "version": 1,
  "members": [
    {
      "_id": 1,
      "host": "user-mongodb-config:27017"
    }
  ]
}
)