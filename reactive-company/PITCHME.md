### Reactive Programming
### And
### Reactive Systems

<span style="color:gray">by example</span>

[https://github.com/idugalic/reactive-company](https://github.com/idugalic/reactive-company)

---

### Reactive Programming

is about non-blocking applications that are:

  - asynchronous
  - message-driven
  - require a small number of threads to scale vertically
  - use back-pressure

+++

### Reactive Programming

<span style="color:gray">Asynchronous</span>

  - Processing of a request occurs at an arbitrary point in time, sometime after it has been transmitted from client to service. 
  - Asynchronous IO refers to an interface where you supply a callback to an IO operation, which is invoked when the operation completes.
  
+++

### Reactive Programming

<span style="color:gray">Non-blocking</span>

  - In concurrent programming an algorithm is considered non-blocking if threads competing for a resource do not have their execution indefinitely postponed by mutual exclusion protecting that resource.
  - Non-blocking IO refers to an interface where IO operations will return immediately with a special error code if called when they are in a state that would otherwise cause them to block.

+++

### Reactive Programming

<span style="color:gray">Message-driven</span>

  - Event-driven system focuses on addressable event sources while a message-driven system concentrates on addressable recipients.
  - Resilience is more difficult to achieve in an event-driven system.

+++

### Reactive Programming

<span style="color:gray">Back-pressure</span>

  - Back-pressure is an important feedback mechanism that allows systems to gracefully respond to load rather than collapse under it.
  - A mechanism to ensure producers don’t overwhelm consumers.
  - One component in the chain will communicate the fact that it is under stress to upstream components and so get them to reduce the load.

+++

### Reactive Programming
#### From imperative to declarative async composition of logic

<span style="color:gray">Service</span>

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

+++

### Reactive Programming
#### From imperative to declarative async composition of logic

<span style="color:gray">Log</span>

A possible log output we could see is:
![Log - Reactive](assets/logs-reactive.png?raw=true)

As we can see the output of the controller method is evaluated after its execution in a different thread too!

+++

### Reactive Programming

<span style="color:gray">Summary</span>

<ol>
	<li class="fragment" data-fragment-index="1">We can no longer think in terms of a linear execution model where one request is handled by one thread</li>
	<li class="fragment" data-fragment-index="2">The reactive streams will be handled by a lot of threads in their lifecycle</li>
	<li class="fragment" data-fragment-index="3">We no longer can rely on thread affinity for things like the security context or transaction handling</li>
</ol>

---

### Reactive Systems

are:

  - responsive
  - resilient
  - elastic
  - message driven
  
![Reactive Traits](assets/reactive-traits.png?raw=true)

+++

### Reactive Systems

<span style="color:gray">Responsive</span>

  - The system responds in a timely manner if at all possible.

+++

### Reactive Systems

<span style="color:gray">Resilient</span>

  - The system stays responsive in the face of failure.

+++

### Reactive Systems
 
<span style="color:gray">Elastic</span>

  - The system stays responsive under varying workload.

+++

### Reactive Systems

<span style="color:gray">Message Driven</span>

  - Reactive Systems rely on asynchronous message-passing to establish a boundary between components that ensures loose coupling, isolation and location transparency.

+++

### Reactive Systems

<span style="color:gray">Summary</span>

<ol>
	<li class="fragment" data-fragment-index="1">Reactive programming offers productivity for developers—through performance and resource efficiency—at the component level for internal logic and dataflow transformation</li>
	<li class="fragment" data-fragment-index="2">Reactive systems offer productivity for architects and DevOps practitioners—through resilience and elasticity—at the system level</li>
</ol>

---

### Why now?

The promise of Reactive is that you can do more with less, specifically you can process higher loads with fewer threads. 

<ol>
	<li class="fragment" data-fragment-index="1">For the right problem, the effects are dramatic</li>
	<li class="fragment" data-fragment-index="2">For the wrong problem, the effects might go into reverse (you actually make things worse)</li>
</ol>

---

### Load And Performance testing

Is your web application responsive? <span style="color:gray">There is only one way to know this: test your web application!</span>

```bash
$ ./mvnw gatling:execute
```

---

### Thank you

 - Ivan Dugalic
 - [http://idugalic.pro/](http://idugalic.pro/)
 - [https://twitter.com/idugalic](https://twitter.com/idugalic)
 - [https://github.com/idugalic](https://github.com/idugalic)


