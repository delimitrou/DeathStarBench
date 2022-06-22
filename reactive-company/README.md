# [projects](http://idugalic.github.io/projects)/reactive-company ![Java CI with Maven](https://github.com/idugalic/reactive-company/workflows/Java%20CI%20with%20Maven/badge.svg?branch=master) [![GitPitch](https://gitpitch.com/assets/badge.svg)](https://gitpitch.com/idugalic/reactive-company/master?grs=github&t=white)

This project is intended to demonstrate best practices for building a reactive web application with Spring 5 platform.

## Table of Contents

   * [Reactive programming and Reactive systems](#reactive-programming-and-reactive-systems)
       * [Why now?](#why-now)
       * [Spring WebFlux (web reactive) module](#spring-webflux-web-reactive-module)
          * [Server side](#server-side)
             * [Annotation based](#annotation-based)
             * [Functional](#functional)
          * [Client side](#client-side)
       * [Spring Reactive data](#spring-reactive-data)
    * [CI with Travis](#ci-with-travis)
    * [Running instructions](#running-instructions)
       * [Run the application by Maven:](#run-the-application-by-maven)
       * [Run the application on Cloud Foundry](#run-the-application-on-cloud-foundry)
       * [Run the application by Docker](#run-the-application-by-docker)
          * [Manage docker swarm with Portainer](#manage-docker-swarm-with-portainer)
          * [Manage docker swarm with CLI](#manage-docker-swarm-with-cli)
             * [List docker services](#list-docker-services)
             * [Scale docker services](#scale-docker-services)
             * [Browse docker service logs](#browse-docker-service-logs)
          * [Swarm mode load balancer](#swarm-mode-load-balancer)
       * [Browse the application:](#browse-the-application)
    * [Load testing with Gatling](#load-testing-with-gatling)
    * [Log output](#log-output)
    * [References and further reading](#references-and-further-reading)


## Reactive programming and Reactive systems

In plain terms reactive programming is about [non-blocking](http://www.reactivemanifesto.org/glossary#Non-Blocking) applications that are [asynchronous](http://www.reactivemanifesto.org/glossary#Asynchronous) and [message-driven](http://www.reactivemanifesto.org/glossary#Message-Driven) and require a small number of threads to [scale](http://www.reactivemanifesto.org/glossary#Scalability) vertically (i.e. within the JVM) rather than horizontally (i.e. through clustering).

A key aspect of reactive applications is the concept of backpressure which is a mechanism to ensure producers don’t overwhelm consumers. For example in a pipeline of reactive components extending from the database to the HTTP response when the HTTP connection is too slow the data repository can also slow down or stop completely until network capacity frees up.

Reactive programming also leads to a major shift from imperative to declarative async composition of logic. It is comparable to writing blocking code vs using the CompletableFuture from Java 8 to compose follow-up actions via lambda expressions.

For a longer introduction check the blog series [“Notes on Reactive Programming”](https://spring.io/blog/2016/06/07/notes-on-reactive-programming-part-i-the-reactive-landscape) by Dave Syer.

"We look at Reactive Programming as one of the methodologies or pieces of the puzzle for Reactive [Systems] as a broader term." Please read the ['Reactive Manifesto'](http://www.reactivemanifesto.org/) and ['Reactive programming vs. Reactive systems'](https://www.oreilly.com/ideas/reactive-programming-vs-reactive-systems) for more informations.

### Why now?

What is driving the rise of Reactive in Enterprise Java? Well, it’s not (all) just a technology fad — people jumping on the bandwagon with the shiny new toys. The driver is efficient resource utilization, or in other words, spending less money on servers and data centres. The promise of Reactive is that you can do more with less, specifically you can process higher loads with fewer threads. This is where the intersection of Reactive and non-blocking, asynchronous I/O comes to the foreground. For the right problem, the effects are dramatic. For the wrong problem, the effects might go into reverse (you actually make things worse). Also remember, even if you pick the right problem, there is no such thing as a free lunch, and Reactive doesn’t solve the problems for you, it just gives you a toolbox that you can use to implement solutions.


### Spring WebFlux (web reactive) module

Spring Framework 5 includes a new spring-webflux module. The module contains support for reactive HTTP and WebSocket clients as well as for reactive server web applications including REST, HTML browser, and WebSocket style interactions.

#### Server side
On the server-side WebFlux supports 2 distinct programming models:

- Annotation-based with @Controller and the other annotations supported also with Spring MVC
- Functional, Java 8 lambda style routing and handling

##### Annotation based
```java
@RestController
public class BlogPostController {

	private final BlogPostRepository blogPostRepository;

	public BlogPostController(BlogPostRepository blogPostRepository) {
		this.blogPostRepository = blogPostRepository;
	}

	@PostMapping("/blogposts")
	Mono<Void> create(@RequestBody Publisher<BlogPost> blogPostStream) {
		return this.blogPostRepository.save(blogPostStream).then();
	}

	@GetMapping("/blogposts")
	Flux<BlogPost> list() {
		return this.blogPostRepository.findAll();
	}

	@GetMapping("/blogposts/{id}")
	Mono<BlogPost> findById(@PathVariable String id) {
		return this.blogPostRepository.findOne(id);
	}
}
```
##### Functional

Functional programming model is not implemented within this application. I am not sure if it is posible to have both models in one application.

Both programming models are executed on the same reactive foundation that adapts non-blocking HTTP runtimes to the Reactive Streams API.

#### Client side

WebFlux includes a functional, reactive WebClient that offers a fully non-blocking and reactive alternative to the RestTemplate. It exposes network input and output as a reactive ClientHttpRequest and ClientHttpRespones where the body of the request and response is a Flux<DataBuffer> rather than an InputStream and OutputStream. In addition it supports the same reactive JSON, XML, and SSE serialization mechanism as on the server side so you can work with typed objects.

```java
@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
public class ApplicationIntegrationTest {

	WebTestClient webTestClient;

	List<BlogPost> expectedBlogPosts;
	List<Project> expectedProjects;

	@Autowired
	BlogPostRepository blogPostRepository;

	@Autowired
	ProjectRepository projectRepository;

	@Before
	public void setup() {
		webTestClient = WebTestClient.bindToController(new BlogPostController(blogPostRepository), new ProjectController(projectRepository)).build();

		expectedBlogPosts = blogPostRepository.findAll().collectList().block();
		expectedProjects = projectRepository.findAll().collectList().block();

	}

	@Test
	public void listAllBlogPostsIntegrationTest() {
		this.webTestClient.get().uri("/blogposts")
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
			.expectBodyList(BlogPost.class).isEqualTo(expectedBlogPosts);
	}

	@Test
	public void listAllProjectsIntegrationTest() {
		this.webTestClient.get().uri("/projects")
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
			.expectBodyList(Project.class).isEqualTo(expectedProjects);
	}

	@Test
	public void streamAllBlogPostsIntegrationTest() throws Exception {
		FluxExchangeResult<BlogPost> result = this.webTestClient.get()
			.uri("/blogposts")
			.accept(TEXT_EVENT_STREAM)
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(TEXT_EVENT_STREAM)
			.returnResult(BlogPost.class);

		StepVerifier.create(result.getResponseBody())
			.expectNext(expectedBlogPosts.get(0), expectedBlogPosts.get(1))
			.expectNextCount(1)
			.consumeNextWith(blogPost -> assertThat(blogPost.getAuthorId(), endsWith("4")))
			.thenCancel()
			.verify();
	}

	...
}

```
Please note that webClient is requesting [Server-Sent Events](https://community.oracle.com/docs/DOC-982924) (text/event-stream).
We could stream individual JSON objects (application/stream+json) but that would not be a valid JSON document as a whole and a browser client has no way to consume a stream other than using Server-Sent Events or WebSocket.

### Spring Reactive data

Spring Data Kay M1 is the first release ever that comes with support for reactive data access. Its initial set of supported stores — MongoDB, Apache Cassandra and Redis

The repositories programming model is the most high-level abstraction Spring Data users usually deal with. They’re usually comprised of a set of CRUD methods defined in a Spring Data provided interface and domain-specific query methods.

In contrast to the traditional repository interfaces, a reactive repository uses reactive types as return types and can do so for parameter types, too.

```java
public interface BlogPostRepository extends ReactiveSortingRepository<BlogPost, String>{

	Flux<BlogPost> findByTitle(Mono<String> title);

}
```
## CI with Travis

The application is build by [Travis](https://travis-ci.org/idugalic/reactive-company). [Pipeline](https://github.com/idugalic/reactive-company/blob/master/.travis.yml) is triggered on every push to master branch.

- Docker image is pushed to [Docker Hub](https://hub.docker.com/r/idugalic/reactive-company/)

## Running instructions

### Run the application by maven:

This application is using embedded mongo database.
You do not have to install and run mongo database before you run the application locally.

You can use NON-embedded version of mongo by setting scope of 'de.flapdoodle.embed.mongo' to 'test'. 
In this case you have to install mongo server locally:

```bash
$ brew install mongodb
$ brew services start mongodb
```

Run it:

```bash
$ cd reactive-company
$ ./mvnw spring-boot:run
```

### Run the application on Cloud Foundry

Run application on local workstation with PCF Dev

- Download and install PCF: https://pivotal.io/platform/pcf-tutorials/getting-started-with-pivotal-cloud-foundry-dev/introduction
- Start the PCF Dev: $ cf dev start -m 8192
- Push the app to PCF Dev: $ ./mvnw cf:push
- Enjoy: http://reactive-company.local.pcfdev.io/

You can adopt any CI pipeline you have to deploy your application on any cloud foundry instance, for example:

```bash
mvn cf:push [-Dcf.appname] [-Dcf.path] [-Dcf.url] [-Dcf.instances] [-Dcf.memory] [-Dcf.no-start] -Dcf.target=https://api.run.pivotal.io
```

### Run the application by Docker

I am running Docker Community Edition, version: 17.05.0-ce-mac11 (Channel: edge).

A [swarm](https://docs.docker.com/engine/swarm/) is a cluster of Docker engines, or nodes, where you deploy services. The Docker Engine CLI and API include commands to manage swarm nodes (e.g., add or remove nodes), and deploy and orchestrate services across the swarm. By running script bellow you will initialize a simple swarm with one node, and you will install services:

- reactive-company
- mongodb (mongo:3.0.4)

```bash
$ cd reactive-company
$ ./docker-swarm.sh
```

#### Manage docker swarm with Portainer

Portainer is a simple management solution for Docker, and is really simple to deploy:

```bash
$ docker service create \
    --name portainer \
    --publish 9000:9000 \
    --constraint 'node.role == manager' \
    --mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock \
    portainer/portainer \
    -H unix:///var/run/docker.sock
```
Visit http://localhost:9000

#### Manage docker swarm with CLI

##### List docker services

```bash
$ docker service ls
```

##### Scale docker services

```bash
$ docker service scale stack_reactive-company=2
```
Now you have two tasks/containers running for this service.

##### Browse docker service logs

```bash
$ docker service logs stack_reactive-company -f
```
You will be able to determine what task/container handled the request.

#### Swarm mode load balancer

When using HTTP/1.1, by default, the TCP connections are left open for reuse. Docker swarm load balancer will not work as expected in this case. You will get routed to the same task of the service every time.

You can use 'curl' command line tool (NOT BROWSER) to avoid this problem.

The Swarm load balancer is a basic Layer 4 (TCP) load balancer. Many applications require additional features, like these, to name just a few:

- SSL/TLS termination
- Content‑based routing (based, for example, on the URL or a header)
- Access control and authorization
- Rewrites and redirects


### Browse the application:

#### Index page

Open your browser and navigate to http://localhost:8080

The response is resolved by [HomeController.java](https://github.com/idugalic/reactive-company/blob/master/src/main/java/com/idugalic/web/HomeController.java) and home.html.

 - Blog posts are fully resolved by the Publisher - Thymeleaf will NOT be executed as a part of the data flow

 - Projects are fully resolved by the Publisher - Thymeleaf will NOT be executed as a part of the data flow

#### Server-Sent Events page

Open your browser and navigate to http://localhost:8080/stream

This view is resolved by [StreamController.java](https://github.com/idugalic/reactive-company/blob/master/src/main/java/com/idugalic/web/StreamController.java) and sse.html template. 


 - *Blog posts* are NOT fully resolved by the Publisher 
 - Thymeleaf will be executed as a part of the data flow 
 - These events will be rendered in HTML by Thymeleaf
 
 ```java
@GetMapping(value = "/stream/blog")
public String blog(final Model model) {
	final Flux<BlogPost> blogPostStream = this.blogPostRepository.findAll().log();
	model.addAttribute("blogPosts", new ReactiveDataDriverContextVariable(blogPostStream, 1000));
	return "sse :: #blogTableBody";
	}
 ```

 - *Projects* are NOT fully resolved by the Publisher 
 - Thymeleaf will be executed as a part of the data flow 
 - These events will be rendered in HTML by Thymeleaf
 
 ```java
@GetMapping(value = "/stream/project")
public String project(final Model model) {
	final Flux<Project> projectStream = this.projectRepository.findAll().log();
	model.addAttribute("projects", new ReactiveDataDriverContextVariable(projectStream, 1000));
	return "sse :: #projectTableBody";
	}
 ```
 
 - *Blog posts (tail)* are NOT fully resolved by the Publisher 
 - Thymeleaf will be executed as a part of the data flow 
 - These events will be rendered in JSON by Spring WebFlux (using Jackson) 
 - We are using a [Tailable Cursor](https://docs.mongodb.com/manual/core/tailable-cursors/) that remains open after the client exhausts the results in the initial cursor. Tailable cursors are conceptually equivalent to the tail Unix command with the -f option (i.e. with “follow” mode). After clients insert new additional documents into a capped collection, the tailable cursor will continue to retrieve documents. You may use a Tailable Cursor with [capped collections](https://docs.mongodb.com/manual/core/capped-collections/) only.
 - If you add a new blog post to the database, it will be displayed on the page in the HTML table.
 
 ```java
@GetMapping("/tail/blogposts")
Flux<BlogPost> tail() {
	LOG.info("Received request: BlogPost - Tail");
	try {
		// Using tailable cursor
		return this.blogPostRepository.findBy().log();
	} finally {
		LOG.info("Request pocessed: BlogPost - Tail");
	}
}
 ```


#### Blog posts (REST API):
```bash
$ curl http://localhost:8080/blogposts
```
or
```bash
$ curl -v -H "Accept: text/event-stream" http://localhost:8080/blogposts
```

#### Projects (REST API):
```bash
$ curl http://localhost:8080/projects
```
or
```bash
$ curl -v -H "Accept: text/event-stream" http://localhost:8080/projects
```

#### Blog posts - tial (REST API)
```bash
$ curl -v -H "Accept: text/event-stream" http://localhost:8080/tail/blogposts
```

##  Load testing with Gatling

Run application first (by maven or docker)

```bash
$ ./mvnw gatling:execute
```

By default src/main/test/scala/com/idugalic/RecordedSimulation.scala will be run.
The reports will be available in the console and in *html files within the 'target/gatling/results' folder

## Log output

A possible log output we could see is:
![Log - Reactive](assets/logs-reactive.png?raw=true)

As we can see the output of the controller method is evaluated after its execution in a different thread too!

```java
@GetMapping("/blogposts")
Flux<BlogPost> list() {
	LOG.info("Received request: BlogPost - List");
	try {
		return this.blogPostRepository.findAll().log();
	} finally {
		LOG.info("Request pocessed: BlogPost - List");
	}
}
```

We can no longer think in terms of a linear execution model where one request is handled by one thread. The reactive streams will be handled by a lot of threads in their lifecycle. This complicates things when we migrate from the old MVC framework. We no longer can rely on thread affinity for things like the security context or transaction handling.

## Slides
<iframe width='770' height='515' src='https://gitpitch.com/idugalic/reactive-company/master?grs=github&t=white' frameborder='0' allowfullscreen></iframe>

## References and further reading

- http://www.reactivemanifesto.org/
- https://www.oreilly.com/ideas/reactive-programming-vs-reactive-systems
- http://www.lightbend.com/blog/the-basics-of-reactive-system-design-for-traditional-java-enterprises
- http://docs.spring.io/spring-framework/docs/5.0.0.BUILD-SNAPSHOT/spring-framework-reference/html/web-reactive.html
- https://spring.io/blog/2016/06/07/notes-on-reactive-programming-part-i-the-reactive-landscape
- https://spring.io/blog/2016/06/13/notes-on-reactive-programming-part-ii-writing-some-code
- http://www.ducons.com/blog/tests-and-thoughts-on-asynchronous-io-vs-multithreading
- https://www.ivankrizsan.se/2016/05/06/introduction-to-load-testing-with-gatling-part-4/
- https://dzone.com/articles/functional-amp-reactive-spring-along-with-netflix
- [asynchronous and non-blocking IO](http://blog.omega-prime.co.uk/?p=155)
- [Functional and Reactive Spring with Reactor and Netflix OSS](https://dzone.com/articles/functional-amp-reactive-spring-along-with-netflix)
- https://www.youtube.com/watch?v=rdgJ8fOxJhc
- https://speakerdeck.com/sdeleuze/functional-web-applications-with-spring-and-kotlin
