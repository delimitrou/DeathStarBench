package com.idugalic.web.blog;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.idugalic.domain.blog.BlogPost;
import com.idugalic.domain.blog.BlogPostRepository;

import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

/**
 * @author idugalic
 * 
 *         Blog post controller
 * 
 *         On the server-side WebFlux supports 2 distinct programming models:
 *
 *         - Annotation-based with @Controller and the other annotations
 *         supported also with Spring MVC.
 *         - Functional, Java 8 lambda style routing and handling
 * 
 *         This is example of 'Annotation-based with @Controller' programming
 *         model.
 *
 */
@RestController
public class BlogPostController {

	private static final Logger LOG = LoggerFactory.getLogger(BlogPostController.class);

	private final BlogPostRepository blogPostRepository;

	public BlogPostController(BlogPostRepository blogPostRepository) {
		this.blogPostRepository = blogPostRepository;
	}

	@PostMapping("/blogposts")
	Mono<BlogPost> create(BlogPost blogPost) {
		LOG.info("Received request: BlogPost - Create");
		try {
			return this.blogPostRepository.save(blogPost).log();
		} finally {
			LOG.info("Request pocessed: BlogPost - Create");
		}
	}

	@GetMapping("/blogposts")
	Flux<BlogPost> list() {
		LOG.info("Received request: BlogPost - List");
		try {
			return this.blogPostRepository.findAll().log();
		} finally {
			LOG.info("Request pocessed: BlogPost - List");
		}
	}
	
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

	@GetMapping("/blogposts/{id}")
	Mono<BlogPost> findById(@PathVariable String id) {
		LOG.info("Received request: BlogPost - FindById");
		try {
			return this.blogPostRepository.findById(id).log();
		} finally {
			LOG.info("Request pocessed: BlogPost - FindById");
		}
	}

	@GetMapping("/blogposts/search/bytitle")
	Flux<BlogPost> findByTitle(@RequestParam String title) {
		LOG.info("Received request: BlogPost - FindByTitle");
		try {
			return this.blogPostRepository.findByTitle(Mono.just(title)).log();
		} finally {
			LOG.info("Request pocessed: BlogPost - FindByTitle");
		}
	}
}
