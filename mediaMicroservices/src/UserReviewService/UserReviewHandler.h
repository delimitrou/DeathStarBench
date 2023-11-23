#ifndef MEDIA_MICROSERVICES_USERREVIEWHANDLER_H
#define MEDIA_MICROSERVICES_USERREVIEWHANDLER_H

#include <iostream>
#include <string>

#include <mongoc.h>
#include <bson/bson.h>

#include "../../gen-cpp/UserReviewService.h"
#include "../../gen-cpp/ReviewStorageService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../ClientPool.h"
#include "../RedisClient.h"
#include "../ThriftClient.h"

namespace media_service
{
  class UserReviewHandler : public UserReviewServiceIf
  {
  public:
    UserReviewHandler( ClientPool<RedisClient> *, mongoc_client_pool_t *, ClientPool<ThriftClient<ReviewStorageServiceClient>> *);
    
    ~UserReviewHandler() override = default;

    void UploadUserReview(int64_t, int64_t, int64_t, int64_t,
                          const std::map<std::string, std::string> &) override;
    void ReadUserReviews(std::vector<Review> &_return, int64_t req_id,
                         int64_t user_id, int32_t start, int32_t stop,
                         const std::map<std::string, std::string> &carrier) override;

  private:
    ClientPool<RedisClient> *_redis_client_pool;
    mongoc_client_pool_t *_mongodb_client_pool;
    ClientPool<ThriftClient<ReviewStorageServiceClient>> *_review_client_pool;
  };

  UserReviewHandler::UserReviewHandler( ClientPool<RedisClient> *redis_client_pool, mongoc_client_pool_t *mongodb_pool, ClientPool<ThriftClient<ReviewStorageServiceClient>> *review_storage_client_pool)
  {
    _redis_client_pool    = redis_client_pool;
    _mongodb_client_pool  = mongodb_pool;
    _review_client_pool   = review_storage_client_pool;
  }


  void UserReviewHandler::UploadUserReview(
      int64_t req_id,
      int64_t user_id,
      int64_t review_id,
      int64_t timestamp,
      const std::map<std::string, std::string> &carrier)
  {

    // Initialize local variable for MongoDB
    mongoc_client_t* mongodb_client = nullptr;
    mongoc_cursor_t* cursor         = nullptr;
    bson_error_t error;
    bson_t const* doc_iter          = nullptr;
    bson_t reply                    = BSON_INITIALIZER;
    bool ok                         = true; // Will be reused by different code snippet


    // Init timing stuff
    TextMapReader reader(carrier);
    std::map<std::string, std::string> writer_text_map;
    TextMapWriter writer(writer_text_map);
    
    // ─── Launch Timing ───────────────────────────────────────────
    auto parent_span = opentracing::Tracer::Global()->Extract(reader);
    auto span = opentracing::Tracer::Global()->StartSpan( "UploadUserReview", {opentracing::ChildOf(parent_span->get())});
    opentracing::Tracer::Global()->Inject(span->context(), writer);
    // ─────────────────────────────────────────────────────────────

    
    // ─── Init Connection To Server ───────────────────────────────

    mongodb_client = mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client)
    {
      ServiceException se{};
      se.errorCode  = ErrorCode::SE_MONGODB_ERROR;
      se.message    = "Failed to pop a client from MongoDB pool";
      throw se;
    }

    // ─── Accessing Collections ───────────────────────────────────

    // Access the "user-review" database and the "users" and "reviews" collections
    mongoc_collection_t* users_collection   = mongoc_client_get_collection(mongodb_client, "user-review", "users");
    mongoc_collection_t* reviews_collection = mongoc_client_get_collection(mongodb_client, "user-review", "reviews");

    if (!users_collection || !reviews_collection)
    {
      ServiceException se{};
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to get access a collection from a MongoDB pool";

      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

      mongoc_collection_destroy(users_collection);
      mongoc_collection_destroy(reviews_collection);
      throw se;
    }

    // ─── Buidling Documents ──────────────────────────────────────
    // Or the mongoDB queries, sort of
    
    bson_t* user_doc = BCON_NEW(
      "user_id", BCON_INT64(user_id)
      );

    bson_t* review_doc = BCON_NEW(
      "review_id",  BCON_INT64(review_id),
      "timestamp",  BCON_INT64(timestamp),
      "user_id",    BCON_INT64(user_id)
      );


    // ─── Db Find Tracing ─────────────────────────────────────────
    auto find_span = opentracing::Tracer::Global()->StartSpan( "MongoFindUser", {opentracing::ChildOf(&span->context())}); 

    // Seleting a user
    cursor = mongoc_collection_find_with_opts(users_collection, user_doc, NULL, NULL);

    find_span->Finish();
    // ─────────────────────────────────────────────────────────────

  
    // ─── Db Insert Tracing ───────────────────────────────────────
    auto insert_span = opentracing::Tracer::Global()->StartSpan( "MongoTotalInsert", {opentracing::ChildOf(&span->context())});

    // ─── Adding The A User If Does Not Exist ─────────────────────

    ok = mongoc_cursor_next(cursor, &doc_iter);
    if(!ok)
    {
      
        if (!mongoc_collection_insert_one(users_collection, user_doc, NULL, NULL, &error)) 
        {
          ServiceException se{};
          se.errorCode = ErrorCode::SE_MONGODB_ERROR;
          se.message = error.message;

          //! If you wonder why those lines are here, and also duplicated to
          //! other parts of the source code It's because someone had the mad
          //! idea to mix C and C++. Therefore there is no good way to
          //! disalocate memory using a C-like `goto;` or with a `unique_ptr<>`
          //! in C++ when anything goes out of scope. Also the compiler isn't
          //! keen with any goto  Well done lads.
          bson_destroy(user_doc);
          bson_destroy(review_doc);
          mongoc_cursor_destroy(cursor);
          mongoc_collection_destroy(users_collection);
          mongoc_collection_destroy(reviews_collection);
          mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        } 
    }

    // ─── Db Update Tracing ───────────────────────────────────────
    auto update_span = opentracing::Tracer::Global()->StartSpan( "MongoReviewInsert", {opentracing::ChildOf(&span->context())});

    // ─── Insert Review Id Anyway ─────────────────────────────────

    if (!mongoc_collection_insert_one(reviews_collection, review_doc, NULL, NULL, &error)) {
      ServiceException se{};
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = error.message;

      bson_destroy(user_doc);
      bson_destroy(review_doc);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(users_collection);
      mongoc_collection_destroy(reviews_collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    }
    
    update_span->Finish();
    insert_span->Finish();
    // ─────────────────────────────────────────────────────────────

    bson_destroy(user_doc);
    bson_destroy(review_doc);
    mongoc_cursor_destroy(cursor);
    mongoc_collection_destroy(users_collection);
    mongoc_collection_destroy(reviews_collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);


    // STARTING REDIS
    
    auto redis_client_wrapper = _redis_client_pool->Pop();
    if (!redis_client_wrapper)
    {
      ServiceException se{};
      se.errorCode  = ErrorCode::SE_REDIS_ERROR;
      se.message    = "Cannot connected to Redis server";
      throw se;
    }
    auto redis_client = redis_client_wrapper->GetClient();

    // ─── Redis Timing ────────────────────────────────────────────
    auto redis_span = opentracing::Tracer::Global()->StartSpan("RedisUpdate", {opentracing::ChildOf(&span->context())});
    // ─────────────────────────────────────────────────────────────

    // ─── Sync Redis I Guess ──────────────────────────────────────
    auto num_reviews = redis_client->zcard(std::to_string(user_id));
    redis_client->sync_commit();
    auto num_reviews_reply = num_reviews.get();

    std::vector<std::string> options{"NX"};
    if (num_reviews_reply.ok() && num_reviews_reply.as_integer())
    {
      std::multimap<std::string, std::string> value = {{std::to_string(timestamp), std::to_string(review_id)}};
      redis_client->zadd(std::to_string(user_id), options, value);
      redis_client->sync_commit();
    }

    _redis_client_pool->Push(redis_client_wrapper);

    // ─── Stop Timer ──────────────────────────────────────────────
    redis_span->Finish();

    // DO NOT FORGET TO FREE RESSOURCES
    // redis ?

    span->Finish();
    // ─────────────────────────────────────────────────────────────


  }

  void UserReviewHandler::ReadUserReviews(
      std::vector<Review> &_return, 
      int64_t req_id,
      int64_t user_id, 
      int32_t start,  // start index
      int32_t stop,   // stop index
      const std::map<std::string, 
      std::string> &carrier)
  {

    // Init timing stuff
    TextMapReader reader(carrier);
    std::map<std::string, std::string> writer_text_map;
    TextMapWriter writer(writer_text_map);
    
    // ─── Launch Timing ───────────────────────────────────────────
    auto parent_span = opentracing::Tracer::Global()->Extract(reader);
    auto span = opentracing::Tracer::Global()->StartSpan( "ReadUserReviews", {opentracing::ChildOf(parent_span->get())}); 
    opentracing::Tracer::Global()->Inject(span->context(), writer);
    // ─────────────────────────────────────────────────────────────

    // Early return if index does not fall into expected range
    // but, i mean..., who is using signed int as index <:^)
    if (stop <= start || start < 0) return;

    // REDIS STUFF
  
    auto redis_client_wrapper = _redis_client_pool->Pop();
    if (!redis_client_wrapper)
    {
      ServiceException se;
      se.errorCode = ErrorCode::SE_REDIS_ERROR;
      se.message = "Cannot connected to Redis server";
      throw se;
    }

    auto redis_client = redis_client_wrapper->GetClient();
    auto redis_span = opentracing::Tracer::Global()->StartSpan( "RedisFind", {opentracing::ChildOf(&span->context())});
    auto review_ids_future = redis_client->zrevrange( std::to_string(user_id), start, stop - 1);
    redis_client->commit();
    redis_span->Finish();

    cpp_redis::reply review_ids_reply;
    try
    {
      review_ids_reply = review_ids_future.get();
    }
    catch (...)
    {
      LOG(error) << "Failed to read review_ids from user-review-redis";
      _redis_client_pool->Push(redis_client_wrapper);
      throw;
    }

    _redis_client_pool->Push(redis_client_wrapper);
    
    std::vector<int64_t> review_ids;
    auto review_ids_reply_array = review_ids_reply.as_array();
    for (auto &review_id_reply : review_ids_reply_array)
    {
      review_ids.emplace_back(std::stoul(review_id_reply.as_string()));
    }
    // END OF REDIS STUFF




    // ─────────────────────────────────────────────────────────────
    // ─────────────────────────────────────────────────────────────
    // ─── MongoDB Stuff ───────────────────────────────────────────
    // ─────────────────────────────────────────────────────────────
    // ─────────────────────────────────────────────────────────────

    // Initialize local variable for MongoDB
    mongoc_collection_t*  reviews_collection  = nullptr;
    mongoc_client_t* mongodb_client           = nullptr;
    mongoc_cursor_t* cursor                   = nullptr;
    bson_t const* doc_el                      = nullptr;
    bson_error_t error;
    bson_iter_t iter;
    bson_t reply                    = BSON_INITIALIZER;
    bool ok                         = true;

    std::multimap<std::string, std::string> redis_update_map;

    unsigned int mongo_start = start + review_ids.size();
  
    // If the index are not broken
    if (mongo_start < stop)
    {

      // ─── Init Connection To Server ───────────────────────────────

      mongodb_client = mongoc_client_pool_pop(_mongodb_client_pool);
      if (!mongodb_client)
      {
        ServiceException se{};
        se.errorCode  = ErrorCode::SE_MONGODB_ERROR;
        se.message    = "Failed to pop a client from MongoDB pool";
        throw se;
      }

      // ─── Accessing Collections ───────────────────────────────────

      // Access the "user-review" database and the "users" and "reviews" collections
      // mongoc_collection_t* users_collection   = mongoc_client_get_collection(mongodb_client, "user-review", "users");
      reviews_collection = mongoc_client_get_collection(mongodb_client, "user-review", "reviews");

      if (!reviews_collection)
      {
        ServiceException se{};
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = "Failed to get access a collection from a MongoDB pool";

        // mongoc_collection_destroy(users_collection);
        mongoc_collection_destroy(reviews_collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

        throw se;
      }

    }

    // ─── Buidling Documents ──────────────────────────────────────
    // Or the mongoDB queries, sort of
      
    bson_t* user_doc = BCON_NEW(
      "user_id", BCON_INT64(user_id)
      );

    bson_t* opts = BCON_NEW(
      "limit", BCON_INT32(stop),
      "skip", BCON_INT32(0),
      "sort", "{ timestamp: -1 }" 
      );


    // ─────────────────────────────────────────────────────────────
    // ─── Gather Result From Db To Update Redis ───────────────────
    // ─────────────────────────────────────────────────────────────

    // Cursor  and cursor index counter
    size_t cursor_idx = 0; 

    // ─── Start Trace ─────────────────────────────────────────────    
    auto find_span = opentracing::Tracer::Global()->StartSpan( "MongoFindUserReviews", {opentracing::ChildOf(&span->context())});

    // Run the find query
    cursor = mongoc_collection_find_with_opts(reviews_collection, user_doc, opts, NULL);

    find_span->Finish();
    // ─────────────────────────────────────────────────────────────

    // ─── Iterrate Over All Gathered Item Of The Collection ───────
    while (mongoc_cursor_next(cursor, &doc_el))
    {
      // Create placeholder for review element
      int32_t review_id;
      int64_t timestamp;

      // Find in the BSON the entries we want
      if (bson_iter_init_find(&iter, doc_el, "review_id") && BSON_ITER_HOLDS_INT32(&iter)) 
        review_id = bson_iter_int32(&iter);
      else
       throw new ServiceException;

      if (bson_iter_find(&iter, "timestamp") && BSON_ITER_HOLDS_INT64(&iter)) 
        timestamp = bson_iter_int64(&iter);
      else
        throw new ServiceException;

      //? If the position of the entries (index) are in certain range we add them
      //? to vectors/map that will be used to update Redis
      if (cursor_idx >= mongo_start) review_ids.emplace_back(review_id);
      redis_update_map.insert({std::to_string(timestamp), std::to_string(review_id)});

      // Increment cursor after iteration
      ++cursor_idx;
    }


    find_span->Finish();

    bson_destroy(opts);
    bson_destroy(user_doc);
    mongoc_cursor_destroy(cursor);
    // mongoc_collection_destroy(users_collection);
    mongoc_collection_destroy(reviews_collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

    // ─────────────────────────────────────────────────────────────
    // ─── Back To Redis Shit ──────────────────────────────────────
    // ─────────────────────────────────────────────────────────────

    std::future<std::vector<Review>> review_future = std::async(std::launch::async, [&]() {

      auto review_client_wrapper = _review_client_pool->Pop();
      
      
      if (!review_client_wrapper) {
        ServiceException se;
        se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
        se.message = "Failed to connected to review-storage-service";
        throw se;
      }
      
      std::vector<Review> _return_reviews;
      
      auto review_client = review_client_wrapper->GetClient();
      
      
      try {
        review_client->ReadReviews( _return_reviews, req_id, review_ids, writer_text_map);
      } catch (...) {
        _review_client_pool->Push(review_client_wrapper);
        LOG(error) << "Failed to read review from review-storage-service";
        throw;
      }

      _review_client_pool->Push(review_client_wrapper);
      return _return_reviews; 
      });

    std::future<cpp_redis::reply> zadd_reply_future;

    if (!redis_update_map.empty())
    { 
      // Update Redis
      redis_client_wrapper = _redis_client_pool->Pop();
      if (!redis_client_wrapper)
      {
        ServiceException se;
        se.errorCode = ErrorCode::SE_REDIS_ERROR;
        se.message = "Cannot connected to Redis server";
        throw se;
      }

      redis_client = redis_client_wrapper->GetClient();
      auto redis_update_span = opentracing::Tracer::Global()->StartSpan( "RedisUpdate", {opentracing::ChildOf(&span->context())});
      
      redis_client->del(std::vector<std::string>{std::to_string(user_id)});
      std::vector<std::string> options{"NX"};
      zadd_reply_future = redis_client->zadd( std::to_string(user_id), options, redis_update_map);
      
      redis_client->commit();
      
      redis_update_span->Finish();
    }

    try
    {
      _return = review_future.get();
    }
    catch (...)
    {
      LOG(error) << "Failed to get review from review-storage-service";
      if (!redis_update_map.empty())
      {
        try
        {
          zadd_reply_future.get();
        }
        catch (...)
        {
          LOG(error) << "Failed to Update Redis Server";
        }
        _redis_client_pool->Push(redis_client_wrapper);
      }
      throw;
    }

    if (!redis_update_map.empty())
    {
      try
      {
        zadd_reply_future.get();
      }
      catch (...)
      {
        LOG(error) << "Failed to Update Redis Server";
        _redis_client_pool->Push(redis_client_wrapper);
        throw;
      }
      _redis_client_pool->Push(redis_client_wrapper);
    }

    span->Finish();
  }

} // namespace media_service

#endif // MEDIA_MICROSERVICES_USERREVIEWHANDLER_H
