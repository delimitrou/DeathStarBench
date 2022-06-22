package com.idugalic.web;

import org.springframework.http.MediaType;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.thymeleaf.spring5.context.webflux.ReactiveDataDriverContextVariable;

import com.idugalic.domain.blog.BlogPost;
import com.idugalic.domain.blog.BlogPostRepository;
import com.idugalic.domain.project.Project;
import com.idugalic.domain.project.ProjectRepository;

import reactor.core.publisher.Flux;

/**
 * 
 * @author idugalic
 * 
 * An streaming / Server-Sent events controller.
 *
 */
@Controller
@RequestMapping("/stream")
public class StreamController {

	private final BlogPostRepository blogPostRepository;
	private final ProjectRepository projectRepository;

	public StreamController(final BlogPostRepository blogPostRepository, final ProjectRepository projectRepository) {
		super();
		this.blogPostRepository = blogPostRepository;
		this.projectRepository = projectRepository;
	}
	
	@GetMapping
	public String stream(final Model model) {
		return "sse";
	}

	@GetMapping(value = "/blog")
	public String blog(final Model model) {

		// Get the stream of BlogPost objects.
		final Flux<BlogPost> blogPostStream = this.blogPostRepository.findAll().log();
		
		// No need to fully resolve the Publisher! We will just let it drive (the "blogPosts" variable can be a Publisher<X>, in which case it will drive the execution of the engine and Thymeleaf will be executed as a part of the data flow)
        // Create a data-driver context variable that sets Thymeleaf in data-driven mode,
        // rendering HTML (iterations) as items are produced in a reactive-friendly manner.
        // This object also works as wrapper that avoids Spring WebFlux trying to resolve
        // it completely before rendering the HTML.
		model.addAttribute("blogPosts", new ReactiveDataDriverContextVariable(blogPostStream, 1000));
		
		// Will use the same "sse" template, but only a fragment: #blogTableBody
		return "sse :: #blogTableBody";
	}
	
	@GetMapping(value = "/project")
	public String project(final Model model) {

		// Get the stream of Project objects.
		final Flux<Project> projectStream = this.projectRepository.findAll().log();
		
		// No need to fully resolve the Publisher! We will just let it drive (the "projects" variable can be a Publisher<X>, in which case it will drive the execution of the engine and Thymeleaf will be executed as a part of the data flow)
        // Create a data-driver context variable that sets Thymeleaf in data-driven mode,
        // rendering HTML (iterations) as items are produced in a reactive-friendly manner.
        // This object also works as wrapper that avoids Spring WebFlux trying to resolve
        // it completely before rendering the HTML.
		model.addAttribute("projects", new ReactiveDataDriverContextVariable(projectStream, 1000));

		// Will use the same "sse" template, but only a fragment: #projectTableBody
		return "sse :: #projectTableBody";
	}


}
