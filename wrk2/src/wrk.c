// Copyright (C) 2012 - Will Glozer.  All rights reserved.

#include "wrk.h"
#include "unistd.h"
#include "script.h"
#include "main.h"
#include "hdr_histogram.h"
#include "stats.h"
#include "assert.h"

// Max recordable latency of 1 day
#define MAX_LATENCY 24L * 60 * 60 * 1000000
uint64_t raw_latency[MAXTHREADS][MAXL];

static struct config {
    uint64_t num_urls;
    uint64_t threads;
    uint64_t connections;
    int dist; //0: fixed; 1: exp; 2: normal; 3: zipf
    uint64_t duration;
    uint64_t timeout;
    uint64_t pipeline;
    uint64_t rate;
    uint64_t delay_ms;
    bool     print_separate_histograms;
    bool     print_sent_requests;
    bool     latency;
    bool     dynamic;
    bool     record_all_responses;
    bool     print_all_responses;
    bool     print_realtime_latency;
    char    *script;
    SSL_CTX *ctx; //https://www.openssl.org/docs/man3.0/man3/SSL_CTX_new.html
} cfg;

static struct {
    stats **requests;
    pthread_mutex_t mutex;
} statistics;


static struct sock sock = {
    .connect  = sock_connect,
    .close    = sock_close,
    .read     = sock_read,
    .write    = sock_write,
    .readable = sock_readable
};

static struct http_parser_settings parser_settings = {
    .on_message_complete = response_complete
};

static volatile sig_atomic_t stop = 0;

static void handler(int sig) {
    stop = 1;
}

static void usage() {
    printf("Usage: wrk <options> <url>                                       \n"
           "  Options:                                                       \n"
           "    -c, --connections <N>  Connections to keep open              \n"
           "    -D, --dist        <S>  fixed, exp, norm, zipf                \n"
           "    -P                     Print each request's latency          \n"
           "    -p                     Print 99th latency every 0.2s to file \n"
           "                           under the current working directory   \n"
           "    -d, --duration    <T>  Duration of test                      \n"
           "    -t, --threads     <N>  Number of threads to use              \n"
           "                                                                 \n"
           "    -s, --script      <S>  Load Lua script file                  \n"
           "    -H, --header      <H>  Add header to request                 \n"
           "    -L, --latency          Print latency statistics              \n"
           "    -S, --separate         Print statistics on different url     \n" 
           "    -T, --timeout     <T>  Socket/request timeout [unit:s]       \n"
           "    -B, --batch_latency    Measure latency of whole              \n"
           "                           batches of pipelined ops              \n"
           "                           (as opposed to each op)               \n"
           "    -r, --requests         Show the number of sent requests      \n"
           "    -v, --version          Print version details                 \n"
           "    -R, --rate        <T>  work rate (throughput)                \n"
           "                           in requests/sec (total)               \n"
           "                           [Required Parameter]                  \n"
           "                                                                 \n"
           "                                                                 \n"
           "  Numeric arguments may include a SI unit (1k, 1M, 1G)           \n"
           "  Time arguments may include a time unit (2s, 2m, 2h)            \n");
}

int main(int argc, char **argv) {
    char *url, **headers = zmalloc(argc * sizeof(char *));
    struct http_parser_url parts = {};

    char **urls = zmalloc(argc * sizeof(url));
    struct http_parser_url *mul_parts = zmalloc(argc * sizeof(struct http_parser_url));

    if (parse_args(&cfg, &urls, &mul_parts, headers, argc, argv)) {
        usage();
        exit(1);
    }

    thread *threads = zcalloc(cfg.num_urls * cfg.threads * sizeof(thread));
    uint64_t connections = cfg.connections / cfg.threads;
    uint64_t throughput = cfg.rate / cfg.threads;
    char *time = format_time_s(cfg.duration);

    lua_State **L = zmalloc(cfg.num_urls * sizeof(lua_State *));
    pthread_mutex_init(&statistics.mutex, NULL);
    statistics.requests = zmalloc(cfg.num_urls * sizeof(stats *));

    /*statitical variables*/
    struct hdr_histogram* total_latency_histogram;
    struct hdr_histogram* total_real_latency_histogram;
    struct hdr_histogram* total_request_histogram;

    hdr_init(1, MAX_LATENCY, 3, &total_latency_histogram);
    hdr_init(1, MAX_LATENCY, 3, &total_real_latency_histogram);
    hdr_init(1, MAX_LATENCY, 3, &total_request_histogram);
    
    uint64_t *runtime_us = zcalloc(cfg.num_urls * sizeof(uint64_t));
    uint64_t total_complete  = 0;
    uint64_t total_bytes     = 0;
    uint64_t sent_requests[cfg.num_urls];
    uint64_t total_sent_requests  = 0;
    uint64_t start_urls[cfg.num_urls];
    /*statitical variables*/

    char **purls = urls;

    for(int id_url=0; id_url< cfg.num_urls; id_url++ ){
        url = *purls;
        purls++;
        parts = *mul_parts;
        mul_parts++;
        
        char *schema  = copy_url_part(url, &parts, UF_SCHEMA);
        char *host    = copy_url_part(url, &parts, UF_HOST);
        char *port    = copy_url_part(url, &parts, UF_PORT);

        char *service = port ? port : schema;

        if (!strncmp("https", schema, 5)) {
            if ((cfg.ctx = ssl_init()) == NULL) {
                fprintf(stderr, "unable to initialize SSL\n");
                ERR_print_errors_fp(stderr);
                exit(1);
            }
            sock.connect  = ssl_connect;
            sock.close    = ssl_close;
            sock.read     = ssl_read;
            sock.write    = ssl_write;
            sock.readable = ssl_readable;
        }

        signal(SIGPIPE, SIG_IGN);
        signal(SIGINT,  SIG_IGN);

        statistics.requests[id_url] = stats_alloc(10);
        hdr_init(1, MAX_LATENCY, 3, &((statistics.requests[id_url])->histogram));
        // create lua_create
        L[id_url] = script_create(cfg.script, url, headers);

        if (!script_resolve(L[id_url], host, service)) {
            char *msg = strerror(errno);
            fprintf(stderr, "unable to connect to %s:%s %s\n", host, service, msg);
            exit(1);
        }

        uint64_t stop_at = time_us() + (cfg.duration * 1000000); // check timeout

        for (uint64_t id_thread = 0; id_thread < cfg.threads; id_thread++) {
            uint64_t i = id_url * cfg.threads + id_thread;
            thread *t = &threads[i];
            t->tid           =  i;
            // aeCreateEventLoop(server.maxclients+CONFIG_FDSET_INCR)
            t->loop          = aeCreateEventLoop(10 + cfg.connections * cfg.num_urls * 3);
            t->connections   = connections;
            t->throughput    = throughput;
            t->stop_at       = stop_at;
            t->complete      = 0;
            t->sent          = 0;
            t->monitored     = 0;
            t->target        = throughput/10; //Shuang
            t->accum_latency = 0;
            t->L = script_create(cfg.script, url, headers);
            script_init(L[id_url], t, argc - optind, &argv[optind]);

            if (i == 0) {
                cfg.pipeline = script_verify_request(t->L);
                cfg.dynamic = !script_is_static(t->L);
                if (script_want_response(t->L)) {
                    printf("script_want_response\n");
                    parser_settings.on_header_field = header_field;
                    parser_settings.on_header_value = header_value;
                    parser_settings.on_body         = response_body;
                }
            }

            if (!t->loop || pthread_create(&t->thread, NULL, &thread_main, t)) {
                char *msg = strerror(errno);
                fprintf(stderr, "unable to create thread %"PRIu64" for %s: %s\n", id_thread, url, msg);
                exit(2);
            }
        }
        printf("Running %s test @ %s\n", time, url);
        printf("  %"PRIu64" threads and %"PRIu64" connections\n\n",
                cfg.threads, cfg.connections);

        start_urls[id_url] = time_us();

    }
    
    struct sigaction sa = {
        .sa_handler = handler,
        .sa_flags   = 0,
    };

    sigfillset(&sa.sa_mask);
    sigaction(SIGINT, &sa, NULL);

    uint64_t start  = time_us();

    purls = urls;
    for(uint64_t id_url=0; id_url< cfg.num_urls; id_url++ ){        
        url = *purls;
        purls++;

        for (uint64_t id_thread = 0; id_thread < cfg.threads; id_thread++) {
            thread *t = &threads[id_url*cfg.threads+id_thread];
            pthread_join(t->thread, NULL); 
        }
        // timer
        runtime_us[id_url] = time_us() -start_urls[id_url];
    }
    
    uint64_t end = time_us();
    uint64_t total_runtime_us = end - start_urls[0];

    purls = urls;

    for(uint64_t id_url=0; id_url< cfg.num_urls; id_url++ ){
        sent_requests[id_url] = 0;

        url = *purls;
        purls++;

        uint64_t complete = 0;
        uint64_t bytes    = 0;
        errors errors     = { 0 };
        // init 2 histograms with MAX_LATENCY
        struct hdr_histogram* latency_histogram;
        struct hdr_histogram* real_latency_histogram;
        hdr_init(1, MAX_LATENCY, 3, &latency_histogram);
        hdr_init(1, MAX_LATENCY, 3, &real_latency_histogram);
        for (uint64_t id_thread = 0; id_thread < cfg.threads; id_thread++) {
            uint64_t i = id_url*cfg.threads+id_thread;
            thread *t = &threads[i];
            complete += t->complete;
            bytes    += t->bytes;

            errors.connect += t->errors.connect;
            errors.read    += t->errors.read;
            errors.write   += t->errors.write;
            errors.timeout += t->errors.timeout;
            errors.status  += t->errors.status;

            hdr_add(latency_histogram, t->latency_histogram);
            hdr_add(real_latency_histogram, t->real_latency_histogram);
            if (cfg.print_all_responses) {
                char filename[10] = {0};
                sprintf(filename, "%" PRIu64 ".txt", i);
                FILE* ff = fopen(filename, "w");
                uint64_t nnum = MAXL;
                if ((t->complete) < nnum) nnum = t->complete;
                for (uint64_t j=1; j < nnum; ++j)
                    fprintf(ff, "%" PRIu64 "\n", raw_latency[i][j]);
                fclose(ff);
            }

            if(cfg.print_sent_requests){
                sent_requests[id_url] += t->sent;
            }
            
        }

        long double runtime_s   = runtime_us[id_url] / 1000000.0;
        long double req_per_s   = complete   / runtime_s;
        long double bytes_per_s = bytes      / runtime_s;

        stats *latency_stats = stats_alloc(10);
        latency_stats->min = hdr_min(latency_histogram);
        latency_stats->max = hdr_max(latency_histogram);
        latency_stats->histogram = latency_histogram;

        if(cfg.print_separate_histograms){
            printf("\n-----------------------------------------------------------------------\n");
            printf("Test Results @ %s \n", url);
            if(cfg.print_sent_requests){
                printf("Sent %ld requests\n\n", sent_requests[id_url]);
            }

            print_stats_header();
            print_stats("Latency", latency_stats, format_time_us);
            print_stats("Req/Sec", statistics.requests[id_url], format_metric);

            if (cfg.latency) {
                print_hdr_latency(latency_histogram,
                        "Recorded Latency");
                printf("-----------------------------------------------------------------------\n");
            }

            char *runtime_msg = format_time_us(runtime_us[id_url]);

            printf("  %"PRIu64" requests in %s, %sB read\n",
                complete, runtime_msg, format_binary(bytes));
            if (errors.connect || errors.read || errors.write || errors.timeout) {
                printf("  Socket errors: connect %d, read %d, write %d, timeout %d\n",
                    errors.connect, errors.read, errors.write, errors.timeout);
            }

            if (errors.status) {
                printf("  Non-2xx or 3xx responses: %d\n", errors.status);
            }

            printf("Requests/sec: %9.2Lf  \n", req_per_s);
            printf("Transfer/sec: %10sB\n", format_binary(bytes_per_s));

            if (script_has_done(L[id_url])) {
                script_summary(L[id_url], runtime_us[id_url], complete, bytes);
                script_errors(L[id_url], &errors);
                script_done(L[id_url], latency_stats, statistics.requests[id_url]);
            }
        }

        if(cfg.print_sent_requests){
            total_sent_requests += sent_requests[id_url];
        }
        total_complete += complete;
        total_bytes += bytes;
        hdr_add(total_latency_histogram, latency_histogram);
        hdr_add(total_real_latency_histogram, real_latency_histogram);
        hdr_add(total_request_histogram, (statistics.requests[id_url])->histogram);
                
    }

    if(cfg.num_urls>1){    
        printf("\n-----------------------------------------------------------------------\n");
        printf("---------------------------Overall Statistics--------------------------\n");
        printf("-----------------------------------------------------------------------\n");
        stats *latency_stats = stats_alloc(10);
        latency_stats->min = hdr_min(total_latency_histogram);
        latency_stats->max = hdr_max(total_latency_histogram);
        latency_stats->histogram = total_latency_histogram;
        stats *request_stats = stats_alloc(10);
        request_stats->min = hdr_min(total_request_histogram);
        request_stats->max = hdr_max(total_request_histogram);
        request_stats->histogram = total_request_histogram;

        if(cfg.print_sent_requests){
            printf("Sent %ld requests\n\n", total_sent_requests);
        }
        print_stats_header();
        print_stats("Latency", latency_stats, format_time_us);
        print_stats("Req/Sec", request_stats, format_metric);

        if (cfg.latency) {
            print_hdr_latency(total_latency_histogram,
                    "Recorded Overall Latency");
            printf("-----------------------------------------------------------------------\n");
        }
        long double total_runtime_s   = total_runtime_us/ 1000000.0;
        long double total_req_per_s   = total_complete   / total_runtime_s;
        long double total_bytes_per_s = total_bytes      / total_runtime_s;
        char *total_runtime_msg = format_time_us(total_runtime_us);

        printf("  Overall %"PRIu64" requests in %s, %sB read\n", total_complete, total_runtime_msg, format_binary(total_bytes));

        printf("Requests/sec: %9.2Lf\n", total_req_per_s);
        printf("Transfer/sec: %10sB\n", format_binary(total_bytes_per_s));
    }
    return 0;
}

void *thread_main(void *arg) {
    thread *thread = arg;
    /* State of an event based program */
    aeEventLoop *loop = thread->loop;
    // connection cs
    thread->cs = zcalloc(thread->connections * sizeof(connection));
    tinymt64_init(&thread->rand, time_us());
    hdr_init(1, MAX_LATENCY, 3, &thread->latency_histogram);
    hdr_init(1, MAX_LATENCY, 3, &thread->real_latency_histogram);

    char *request = NULL;
    size_t length = 0;

    if (!cfg.dynamic) {
        script_request(thread->L, &request, &length);
    }
    thread->ff = NULL;

    if ((cfg.print_realtime_latency) && ((thread->tid%cfg.threads) == 0)) {
        char filename[50];
        // snprintf(filename, 50, "/filer-01/datasets/nginx/url%" PRIu64 "thread%" PRIu64 ".txt", (thread->tid/cfg.threads), (thread->tid%cfg.threads));
        snprintf(filename, 50, "%s/url%" PRIu64 "thread%" PRIu64 ".txt", getcwd(NULL,0), (thread->tid/cfg.threads), (thread->tid%cfg.threads));
        printf("filename %s\n",filename);
        thread->ff = fopen(filename, "w");
    }

    // thread->throughput: request/thread/sec
    // c->throughput = throughput: request/connection/us
    double throughput = (thread->throughput / 1000000.0) / thread->connections;

    connection *c = thread->cs;
    // connection *pc = thread->cs;
    for (uint64_t i = 0; i < thread->connections; i++, c++) {
        c->thread     = thread;
        c->ssl        = cfg.ctx ? SSL_new(cfg.ctx) : NULL;
        c->request    = request;
        c->length     = length;
        c->interval   = 1000000*thread->connections/thread->throughput;
        c->throughput = throughput;
        c->complete   = 0;
        c->estimate   = 0;
        c->sent       = 0;
        //Stagger connects 1 msec apart within thread
        uint64_t id = thread->tid * thread->connections + i;
        aeCreateTimeEvent(loop, i, delayed_initial_connect, c, NULL);
    }

    uint64_t calibrate_delay = CALIBRATE_DELAY_MS + (thread->connections);
    uint64_t timeout_delay = TIMEOUT_INTERVAL_MS + (thread->connections);

    aeCreateTimeEvent(loop, calibrate_delay, calibrate, thread, NULL);
    aeCreateTimeEvent(loop, timeout_delay, check_timeouts, thread, NULL);

    thread->start = time_us();
    /*aeMain does the job of processing the event loop that is initialized in the previous phase.*/
    aeMain(loop);

    aeDeleteEventLoop(loop);
 
    zfree(thread->cs);

    if (cfg.print_realtime_latency && (thread->tid % cfg.threads) == 0) fclose(thread->ff);

    return NULL;
}

static int connect_socket(thread *thread, connection *c) {
    struct addrinfo *addr = thread->addr;
    struct aeEventLoop *loop = thread->loop;
    int fd, flags;

    fd = socket(addr->ai_family, addr->ai_socktype, addr->ai_protocol);

    flags = fcntl(fd, F_GETFL, 0);
    fcntl(fd, F_SETFL, flags | O_NONBLOCK);

    if (connect(fd, addr->ai_addr, addr->ai_addrlen) == -1) {
        if (errno != EINPROGRESS) goto error;
    }

    flags = 1;
    setsockopt(fd, IPPROTO_TCP, TCP_NODELAY, &flags, sizeof(flags));

    flags = AE_READABLE | AE_WRITABLE;

    if (aeCreateFileEvent(loop, fd, flags, socket_connected, c) == AE_OK) {
        c->parser.data = c;
        c->fd = fd;
        return fd;
    }
            
  error:
    thread->errors.connect++;
    close(fd);
    return -1;
}

static int reconnect_socket(thread *thread, connection *c) {
    aeDeleteFileEvent(thread->loop, c->fd, AE_WRITABLE | AE_READABLE);
    sock.close(c);
    close(c->fd);
    printf("reconnect_socket\n");
    return connect_socket(thread, c);
}

static int delayed_initial_connect(aeEventLoop *loop, long long id, void *data) {
    connection* c = data;
    c->thread_start = time_us();
    c->thread_next  = c->thread_start;
    connect_socket(c->thread, c);
    return AE_NOMORE;
}

static int calibrate(aeEventLoop *loop, long long id, void *data) {
    thread *thread = data;

    long double mean = hdr_mean(thread->latency_histogram);
    long double latency = hdr_value_at_percentile(
            thread->latency_histogram, 90.0) / 1000.0L;
    long double interval = MAX(latency * 2, 10);

    if (mean == 0) return CALIBRATE_DELAY_MS;

    thread->mean     = (uint64_t) mean;
    hdr_reset(thread->latency_histogram);

    thread->start    = time_us();
    thread->interval = interval;
    thread->requests = 0;

    uint64_t id_url = thread->tid / cfg.threads;
    uint64_t id_thread = thread->tid % cfg.threads;
    
    printf("  Thread calibration: mean lat.: %.3fms, rate sampling interval: %dms\n",
            (thread->mean)/1000.0,
            thread->interval);

    aeCreateTimeEvent(loop, thread->interval, sample_rate, thread, NULL);

    return AE_NOMORE;
}

static int check_timeouts(aeEventLoop *loop, long long id, void *data) {
    thread *thread = data;
    connection *c  = thread->cs;
    uint64_t now   = time_us();

    uint64_t maxAge = now - (cfg.timeout * 1000); // us
    // printf("check_timeouts maxAge %ld, c->start %ld\n", maxAge, c->start);
    for (uint64_t i = 0; i < thread->connections; i++, c++) {
        if (maxAge > c->start) {
            thread->errors.timeout++;
        }
    }

    if (stop || now >= thread->stop_at) {
        aeStop(loop);
    }

    return TIMEOUT_INTERVAL_MS; // call check_timeouts after 2s
}

static int sample_rate(aeEventLoop *loop, long long id, void *data) {
    thread *thread = data;

    uint64_t elapsed_ms = (time_us() - thread->start) / 1000;
    // requests here indicates: real-time throughput of thread. (req/sec/thread)
    uint64_t requests = (thread->requests / (double) elapsed_ms) * 1000;

    uint64_t id_url = thread->tid / cfg.threads; 
    pthread_mutex_lock(&statistics.mutex);
    stats_record(statistics.requests[id_url], requests);
    pthread_mutex_unlock(&statistics.mutex);

    thread->requests = 0;
    thread->start    = time_us();
    return thread->interval; // call sample_rate again after thread->interval
}

static int header_field(http_parser *parser, const char *at, size_t len) {
    connection *c = parser->data;
    if (c->state == VALUE) {
        *c->headers.cursor++ = '\0';
        c->state = FIELD;
    }
    buffer_append(&c->headers, at, len);
    return 0;
}

static int header_value(http_parser *parser, const char *at, size_t len) {
    connection *c = parser->data;
    if (c->state == FIELD) {
        *c->headers.cursor++ = '\0';
        c->state = VALUE;
    }
    buffer_append(&c->headers, at, len);
    return 0;
}

static int response_body(http_parser *parser, const char *at, size_t len) {
    connection *c = parser->data;
    buffer_append(&c->body, at, len);
    return 0;
}

uint64_t gen_zipf(connection *conn)
{
    static int first = 1;      // Static first time flag
    static double c = 0;          // Normalization constant
    static double scalar = 0;
    double z;                     // Uniform random number (0 < z < 1)
    double sum_prob;              // Sum of probabilities
    double zipf_value;            // Computed exponential value to be returned
    int n = 100;
    double alpha = 3;

    // Compute normalization constant on first call only
    if (first == 1) {
        for (int i=1; i<=n; i++)
            c = c + (1.0 / pow((double) i, alpha));
        c = 1.0 / c;
        for (int i=1; i<=n; i++) {
            double prob = c / pow((double) i, alpha);
            scalar = scalar + i*prob;
        }

        scalar = conn->interval / scalar;
        first = 0;
    }

  // Pull a uniform random number (0 < z < 1)
    do {
        z = (double)rand()/RAND_MAX;
    } while ((z == 0) || (z == 1));

    // Map z to the value
    sum_prob = 0;
    for (int i=1; i<=n; i++) {
        sum_prob = sum_prob + c / pow((double) i, alpha);
        if (sum_prob >= z) {
            zipf_value = i;
            break;
        }
    }
    return (uint64_t)(zipf_value*scalar);
}

uint64_t gen_exp(connection *c) {
    double z;
    double exp_value;
    do {
        z = (double)rand()/RAND_MAX;
    } while ((z == 0) || (z == 1));
    exp_value = (-log(z)*(c->interval));
    return (uint64_t)(exp_value);
}

uint64_t gen_next(connection *c) {
    if (cfg.dist == 0) { // FIXED
        return c->interval;
    }
    else if (cfg.dist == 1) { // EXP
        return gen_exp(c);
    }
    else if (cfg.dist == 2) {
    }
    else if (cfg.dist == 3) {
       return gen_zipf(c);
    }
    return 0;
}

static uint64_t usec_to_next_send(connection *c) {
    uint64_t now = time_us();
    //c->thread_next = c->thread_start + c->sent/c->throughput;
    if (c->estimate <= c->sent) {
        ++c->estimate;
        c->thread_next += gen_next(c);
    }
    if ((c->thread_next) > now)
        return c->thread_next - now;
    else
        return 0;
    
        
}

static int delay_request(aeEventLoop *loop, long long id, void *data) {
    connection* c = data;
    uint64_t time_usec_to_wait = usec_to_next_send(c);
    if (time_usec_to_wait) {
        return round((time_usec_to_wait / 1000.0L) + 0.5); /* don't send, wait */
    }
    aeCreateFileEvent(c->thread->loop, c->fd, AE_WRITABLE, socket_writeable, c);
    return AE_NOMORE;
}

static int response_complete(http_parser *parser) {
    connection *c = parser->data;
    thread *thread = c->thread;
    uint64_t now = time_us();
    int status = parser->status_code;

    thread->complete++;
    thread->requests++;

    if (status > 399) {
        thread->errors.status++;
    }

    if (c->headers.buffer) {
        *c->headers.cursor++ = '\0';
        script_response(thread->L, status, &c->headers, &c->body);
        c->state = FIELD;
    }

    if (now >= thread->stop_at) {
        aeStop(thread->loop);
        goto done;
    }


    // Record if needed, either last in batch or all, depending in cfg:
    if (cfg.record_all_responses) {
        assert(now > c->actual_latency_start[c->complete & MAXO] );
        uint64_t actual_latency_timing = now - c->actual_latency_start[c->complete & MAXO];
        hdr_record_value(thread->latency_histogram, actual_latency_timing);
        hdr_record_value(thread->real_latency_histogram, actual_latency_timing);

        thread->monitored++;
        thread->accum_latency += actual_latency_timing;
        // thread->target  = throughput/10; response side twice the interval.
        if (thread->monitored == thread->target) {
            if (cfg.print_realtime_latency && (thread->tid%cfg.threads) == 0) {
                fprintf(thread->ff, "%" PRId64 "\n", hdr_value_at_percentile(thread->real_latency_histogram, 99));
                fflush(thread->ff);
            }
            thread->monitored = 0;
            thread->accum_latency = 0;
            hdr_reset(thread->real_latency_histogram);
        }
        if (cfg.print_all_responses && ((thread->complete) < MAXL))
            raw_latency[thread->tid][thread->complete] = actual_latency_timing;
    }

    // Count all responses (including pipelined ones:)
    c->complete++;
    if (!http_should_keep_alive(parser)) {
        reconnect_socket(thread, c);
        goto done;
    }

    http_parser_init(parser, HTTP_RESPONSE);

  done:
    return 0;
}

static void socket_connected(aeEventLoop *loop, int fd, void *data, int mask) {
    connection *c = data;

    switch (sock.connect(c)) {
        case OK:    break;
        case ERROR: goto error;
        case RETRY: return;
    }

    http_parser_init(&c->parser, HTTP_RESPONSE);
    c->written = 0;

    aeCreateFileEvent(c->thread->loop, fd, AE_READABLE, socket_readable, c);

    aeCreateFileEvent(c->thread->loop, fd, AE_WRITABLE, socket_writeable, c);

    return;

  error:
    c->thread->errors.connect++;
    reconnect_socket(c->thread, c);

}

static void socket_writeable(aeEventLoop *loop, int fd, void *data, int mask) {
    connection *c = data;
    thread *thread = c->thread;
    if (!c->written) {
        uint64_t time_usec_to_wait = usec_to_next_send(c);
        if (time_usec_to_wait) {
            int msec_to_wait = round((time_usec_to_wait / 1000.0L) + 0.5);

            // Not yet time to send. Delay:
            aeDeleteFileEvent(loop, fd, AE_WRITABLE);
            aeCreateTimeEvent(
                    thread->loop, msec_to_wait, delay_request, c, NULL);
            return;
        }
    }

    if (!c->written && cfg.dynamic) {
        script_request(thread->L, &c->request, &c->length);
    }
    // buf: indicates the pos. where the request remaining to send is.
    char  *buf = c->request + c->written;
    // len: the length of request that haven't been sent.
    size_t len = c->length  - c->written;
    size_t n;
    switch (sock.write(c, buf, len, &n)) {
        case OK:    break;
        case ERROR: goto error;
        case RETRY: return;
    }
    if (!c->written) {
        c->start = time_us();
        c->actual_latency_start[c->sent & MAXO] = c->start;
        c->sent ++;
    }
    c->written += n;
    if (c->written == c->length) {
        c->written = 0;
        c->thread->sent ++;

        aeDeleteFileEvent(loop, fd, AE_WRITABLE);
        aeCreateFileEvent(thread->loop, c->fd, AE_WRITABLE, socket_writeable, c);
    }
    return;

  error:
    thread->errors.write++;
    reconnect_socket(thread, c);
}


static void socket_readable(aeEventLoop *loop, int fd, void *data, int mask) {
    connection *c = data;
    size_t n;

    do {
        switch (sock.read(c, &n)) {
            case OK:    break;
            case ERROR: goto error;
            case RETRY: return;
        }
        if (http_parser_execute(&c->parser, &parser_settings, c->buf, n) != n) goto error;
        c->thread->bytes += n;
    } while (n == RECVBUF && sock.readable(c) > 0);

    return;

  error:
    c->thread->errors.read++;
    reconnect_socket(c->thread, c);
}

static uint64_t time_us() {
    struct timeval t;
    gettimeofday(&t, NULL);
    return (t.tv_sec * 1000000) + t.tv_usec;
}

static char *copy_url_part(char *url, struct http_parser_url *parts, enum http_parser_url_fields field) {
    char *part = NULL;

    if (parts->field_set & (1 << field)) {
        uint16_t off = parts->field_data[field].off;
        uint16_t len = parts->field_data[field].len;
        part = zcalloc(len + 1 * sizeof(char));
        memcpy(part, &url[off], len);
    }

    return part;
}

static struct option longopts[] = {
    { "connections",    required_argument, NULL, 'c' },
    { "duration",       required_argument, NULL, 'd' },
    { "threads",        required_argument, NULL, 't' },
    { "script",         required_argument, NULL, 's' },
    { "header",         required_argument, NULL, 'H' },
    { "separate",       no_argument,       NULL, 'S' },
    { "latency",        no_argument,       NULL, 'L' },
    { "requests",       no_argument,       NULL, 'r' },
    { "batch_latency",  no_argument,       NULL, 'B' },
    { "timeout",        required_argument, NULL, 'T' },
    { "help",           no_argument,       NULL, 'h' },
    { "version",        no_argument,       NULL, 'v' },
    { "rate",           required_argument, NULL, 'R' },
    { "dist",           required_argument, NULL, 'D' },
    { NULL,             0,                 NULL,  0  }
};

static int parse_args(struct config *cfg, char ***urls, struct http_parser_url **mul_parts, char **headers, int argc, char **argv) {
    int c;
    char **header = headers;
    char ***url =   zmalloc(sizeof(urls));
    *url = *urls;
    struct http_parser_url **mul_part = zmalloc(sizeof(mul_parts));
    *mul_part = *mul_parts;

    memset(cfg, 0, sizeof(struct config));
    cfg->num_urls = 0;
    cfg->threads     = 2;
    cfg->connections = 10;
    cfg->duration    = 10;
    cfg->timeout     = SOCKET_TIMEOUT_MS;
    cfg->rate        = 0;
    cfg->record_all_responses = true;
    cfg->print_all_responses = false;
    cfg->print_realtime_latency = false;
    cfg->print_separate_histograms = false;
    cfg->print_sent_requests = false;
    cfg->dist = 0;

    while ((c = getopt_long(argc, argv, "t:c:s:d:D:H:T:R:LPrSpBv?", longopts, NULL)) != -1) {
        switch (c) {
            case 't':
                if (scan_metric(optarg, &cfg->threads)) return -1;
                break;
            case 'c':
                if (scan_metric(optarg, &cfg->connections)) return -1;
                break;
            case 'D':
                if (!strcmp(optarg, "fixed"))
                    cfg->dist = 0;
                if (!strcmp(optarg, "exp"))
                    cfg->dist = 1;
                if (!strcmp(optarg, "norm"))
                    cfg->dist = 2;
                if (!strcmp(optarg, "zipf"))
                    cfg->dist = 3;
                break;
            case 'd':
                if (scan_time(optarg, &cfg->duration)) return -1;
                break;
            case 's':
                cfg->script = optarg;
                break;
            case 'H':
                *header++ = optarg;
                break;
            case 'P': /* Shuang: print each requests's latency */
                cfg->print_all_responses = true;
                break;
            case 'p': /* Shuang: print avg latency every 0.2s */
                cfg->print_realtime_latency = true;
                break;
            case 'L':
                cfg->latency = true;
                break;
            case 'r':
                cfg->print_sent_requests = true;
                break;
            case 'S':
                cfg->print_separate_histograms = true;
                break;
            case 'B':
                cfg->record_all_responses = false;
                break;
            case 'T':
                if (scan_time(optarg, &cfg->timeout)) return -1;
                cfg->timeout *= 1000;
                break;
            case 'R':
                if (scan_metric(optarg, &cfg->rate)) return -1;
                break;
            case 'v':
                printf("wrk %s [%s] ", VERSION, aeGetApiName());
                printf("Copyright (C) 2012 Will Glozer\n");
                break;
            case 'h': 
            case '?': 
            case ':': 
            default:
                return -1;
        }
    }

    if (optind == argc || !cfg->threads || !cfg->duration) return -1;

    if (!cfg->connections || cfg->connections < cfg->threads) {
        fprintf(stderr, "number of connections must be >= threads\n");
        return -1;
    }

    if (cfg->rate == 0) {
        fprintf(stderr,
                "Throughput MUST be specified with the --rate or -R option\n");
        return -1;
    }

    for(int i = optind; i<argc; i++)
    {   
        if (!script_parse_url(argv[i], *mul_part)) {
            fprintf(stderr, "invalid URL: %s\n", argv[i]);
            return -1;
        }
        **url = argv[i];
        (*mul_part)++;
        (*url)++;
        cfg->num_urls += 1;
    }
    cfg->print_separate_histograms = (cfg->print_separate_histograms || (cfg->num_urls==1));
    *header = NULL;
    *mul_part = NULL;
    *url = NULL;
    return 0;
}

#include <sys/stat.h>
int file_size(char* filename)
{
    struct stat statbuf;
    stat(filename,&statbuf);
    int size=statbuf.st_size;
 
    return size;
}

static void print_stats_header() {
    printf("  Thread Stats%6s%11s%8s%12s\n", "Avg", "Stdev", "99%", "+/- Stdev");
}

static void print_units(long double n, char *(*fmt)(long double), int width) {
    char *msg = fmt(n);
    int len = strlen(msg), pad = 2;

    if (isalpha(msg[len-1])) pad--;
    if (isalpha(msg[len-2])) pad--;
    width -= pad;

    printf("%*.*s%.*s", width, width, msg, pad, "  ");

    free(msg);
}

static void print_stats(char *name, stats *stats, char *(*fmt)(long double)) {
    uint64_t max = stats->max;
    long double mean  = stats_summarize(stats);
    long double stdev = stats_stdev(stats, mean);

    printf("    %-10s", name);
    print_units(mean,  fmt, 8);
    print_units(stdev, fmt, 10);
    print_units(stats_percentile(stats, 99.0), fmt, 9);
    printf("%8.2Lf%%\n", stats_within_stdev(stats, mean, stdev, 1));
}

static void print_hdr_latency(struct hdr_histogram* histogram, const char* description) {
    long double percentiles[] = { 50.0, 75.0, 90.0, 99.0, 99.9, 99.99, 99.999, 100.0};
    printf("  Latency Distribution (HdrHistogram - %s)\n", description);
    for (size_t i = 0; i < sizeof(percentiles) / sizeof(long double); i++) {
        long double p = percentiles[i];
        int64_t n = hdr_value_at_percentile(histogram, p);
        printf("%7.3Lf%%", p);
        print_units(n, format_time_us, 10);
        printf("\n");
    }
    printf("\n%s\n", "  Detailed Percentile spectrum:");
    hdr_percentiles_print(histogram, stdout, 5, 1000.0, CLASSIC);
}

static void print_stats_latency(stats *stats) {
    long double percentiles[] = { 50.0, 75.0, 90.0, 99.0, 99.9, 99.99, 99.999, 100.0 };
    printf("  Latency Distribution\n");
    for (size_t i = 0; i < sizeof(percentiles) / sizeof(long double); i++) {
        long double p = percentiles[i];
        uint64_t n = stats_percentile(stats, p);
        printf("%7.3Lf%%", p);
        print_units(n, format_time_us, 10);
        printf("\n");
    }
}
