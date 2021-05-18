#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MONGODB_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MONGODB_H_

#include <mongoc.h>
#include <bson/bson.h>

#define SERVER_SELECTION_TIMEOUT_MS 300

namespace social_network {

mongoc_client_pool_t* init_mongodb_client_pool(
    const json &config_json,
    const std::string &service_name,
    uint32_t max_size
) {
  std::string addr = config_json[service_name + "-mongodb"]["addr"];
  int port = config_json[service_name + "-mongodb"]["port"];
  std::string uri_str = "mongodb://" + addr + ":" +
      std::to_string(port) + "/?appname=" + service_name + "-service";
  uri_str += "&" MONGOC_URI_SERVERSELECTIONTIMEOUTMS "="
      + std::to_string(SERVER_SELECTION_TIMEOUT_MS);

  mongoc_init();
  bson_error_t error;
  mongoc_uri_t *mongodb_uri =
      mongoc_uri_new_with_error(uri_str.c_str(), &error);

  if (!mongodb_uri) {
    LOG(fatal) << "Error: failed to parse URI" << std::endl
              << "error message: " << std::endl
              << uri_str << std::endl
              << error.message<< std::endl;
    return nullptr;
  } else {
    if (config_json["ssl"]["enabled"]) {
      std::string ca_file = config_json["ssl"]["caPath"];

      mongoc_uri_set_option_as_bool(mongodb_uri, MONGOC_URI_TLS, true);
      mongoc_uri_set_option_as_utf8(mongodb_uri, MONGOC_URI_TLSCAFILE, ca_file.c_str());
      mongoc_uri_set_option_as_bool(mongodb_uri, MONGOC_URI_TLSALLOWINVALIDHOSTNAMES, true);
    }

    mongoc_client_pool_t *client_pool= mongoc_client_pool_new(mongodb_uri);
    mongoc_client_pool_max_size(client_pool, max_size);
    return client_pool;
  }
}

bool CreateIndex(
    mongoc_client_t *client,
    const std::string &db_name,
    const std::string &index,
    bool unique) {
  mongoc_database_t *db;
  bson_t keys;
  char *index_name;
  bson_t *create_indexes;
  bson_t reply;
  bson_error_t error;
  bool r;

  db = mongoc_client_get_database(client, db_name.c_str());
  bson_init (&keys);
  BSON_APPEND_INT32(&keys, index.c_str(), 1);
  index_name = mongoc_collection_keys_to_index_string(&keys);
  create_indexes = BCON_NEW (
      "createIndexes", BCON_UTF8(db_name.c_str()),
      "indexes", "[", "{",
          "key", BCON_DOCUMENT (&keys),
          "name", BCON_UTF8 (index_name),
          "unique", BCON_BOOL(unique),
      "}", "]");
  r = mongoc_database_write_command_with_opts (
      db, create_indexes, NULL, &reply, &error);
  if (!r) {
    LOG(error) << "Error in createIndexes: " << error.message;
  }
  bson_free (index_name);
  bson_destroy (&reply);
  bson_destroy (create_indexes);
  mongoc_database_destroy(db);

  return r;
}

} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MONGODB_H_
