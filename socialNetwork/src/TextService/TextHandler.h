#ifndef SOCIAL_NETWORK_MICROSERVICES_TEXTHANDLER_H
#define SOCIAL_NETWORK_MICROSERVICES_TEXTHANDLER_H

#include <iostream>
#include <string>
#include <regex>
#include <future>


#include "../../gen-cpp/TextService.h"
#include "../../gen-cpp/ComposePostService.h"
#include "../../gen-cpp/UserMentionService.h"
#include "../../gen-cpp/UrlShortenService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"

namespace social_network {

class TextHandler : public TextServiceIf {
 public:
  TextHandler(
      ClientPool<ThriftClient<ComposePostServiceClient>> *,
      ClientPool<ThriftClient<UrlShortenServiceClient>> *,
      ClientPool<ThriftClient<UserMentionServiceClient>> *);
  ~TextHandler() override = default;

  void UploadText(int64_t, const std::string &,
      const std::map<std::string, std::string> &) override;
 private:
  ClientPool<ThriftClient<ComposePostServiceClient>> *_compose_client_pool;
  ClientPool<ThriftClient<UrlShortenServiceClient>> *_url_client_pool;
  ClientPool<ThriftClient<UserMentionServiceClient>> *_user_mention_client_pool;
};

TextHandler::TextHandler(
    ClientPool<ThriftClient<ComposePostServiceClient>> *compose_client_pool,
    ClientPool<ThriftClient<UrlShortenServiceClient>> *url_client_pool,
    ClientPool<ThriftClient<UserMentionServiceClient>> *user_mention_client_pool) {
  _compose_client_pool = compose_client_pool;
  _url_client_pool = url_client_pool;
  _user_mention_client_pool = user_mention_client_pool;
}

void TextHandler::UploadText(
    int64_t req_id,
    const std::string &text,
    const std::map<std::string, std::string> & carrier) {

  // Initialize a span
  TextMapReader reader(carrier);
  std::map<std::string, std::string> writer_text_map;
  TextMapWriter writer(writer_text_map);
  auto parent_span = opentracing::Tracer::Global()->Extract(reader);
  auto span = opentracing::Tracer::Global()->StartSpan(
      "UploadText",
      { opentracing::ChildOf(parent_span->get()) });
  opentracing::Tracer::Global()->Inject(span->context(), writer);

  std::vector<std::string> user_mentions;
  std::smatch m;
  std::regex e("@[a-zA-Z0-9-_]+");
  auto s = text;
  while (std::regex_search(s, m, e)){
    auto user_mention = m.str();
    user_mention = user_mention.substr(1, user_mention.length());
    user_mentions.emplace_back(user_mention);
    s = m.suffix().str();
  }

  std::vector<std::string> urls;
  e = "(http://|https://)([a-zA-Z0-9_!~*'().&=+$%-]+)";
  s = text;
  while (std::regex_search(s, m, e)){
    auto url = m.str();
    urls.emplace_back(url);
    s = m.suffix().str();
  }

  std::future<std::vector<std::string>> shortened_urls_future = std::async(
      std::launch::async, [&](){
        auto url_client_wrapper = _url_client_pool->Pop();
        if (!url_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to url-shorten-service";
          throw se;
        }
        std::vector<std::string> return_urls;
        auto url_client = url_client_wrapper->GetClient();
        try {
          url_client->UploadUrls(return_urls, req_id, urls, writer_text_map);
        } catch (...) {
          LOG(error) << "Failed to upload urls to url-shorten-service";
          _url_client_pool->Push(url_client_wrapper);
          throw;
        }    
        
        _url_client_pool->Push(url_client_wrapper);
        return return_urls;
      });

  std::future<void> user_mention_future = std::async(
      std::launch::async, [&](){
        auto user_mention_client_wrapper = _user_mention_client_pool->Pop();
        if (!user_mention_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to user-mention-service";
          throw se;
        }
        std::vector<std::string> urls;
        auto user_mention_client = user_mention_client_wrapper->GetClient();
        try {
          user_mention_client->UploadUserMentions(req_id, user_mentions,
                                                  writer_text_map);
        } catch (...) {
          LOG(error) << "Failed to upload user_mentions to user-mention-service";
          _user_mention_client_pool->Push(user_mention_client_wrapper);
          throw;
        }

        _user_mention_client_pool->Push(user_mention_client_wrapper);

      });

  std::vector<std::string> shortened_urls;
  try {
    shortened_urls = shortened_urls_future.get();
  } catch (...) {
    LOG(error) << "Failed to get shortened urls from url-shorten-service";
    throw;
  }

  std::string updated_text;
  if (!urls.empty()) {
    s = text;
    int idx = 0;
    while (std::regex_search(s, m, e)){
      auto url = m.str();
      urls.emplace_back(url);
      updated_text += m.prefix().str() + shortened_urls[idx];
      s = m.suffix().str();
      idx++;
    }
  } else {
    updated_text = text;
  }

  
  std::future<void> upload_text_future = std::async(
      std::launch::async, [&]() {
        // Upload to compose post service
        auto compose_post_client_wrapper = _compose_client_pool->Pop();
        if (!compose_post_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to compose-post-service";
          throw se;
        }
        auto compose_post_client = compose_post_client_wrapper->GetClient();
        try {
          compose_post_client->UploadText(req_id, updated_text, writer_text_map);
        } catch (...) {
          LOG(error) << "Failed to upload text to compose-post-service";
          _compose_client_pool->Push(compose_post_client_wrapper);
          throw;
        }          
        _compose_client_pool->Push(compose_post_client_wrapper);
      });

  try {
    user_mention_future.get();
  } catch (...) {
    LOG(error) << "Failed to upload user mentions to user-mention-service";
    throw;
  }

  try {
    upload_text_future.get();
  } catch (...) {
    LOG(error) << "Failed to upload text to compose-post-service";
    throw;
  }

  span->Finish();
}

} //namespace social_network





#endif //SOCIAL_NETWORK_MICROSERVICES_TEXTHANDLER_H
