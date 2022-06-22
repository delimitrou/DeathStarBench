package com.idugalic;

import static org.hamcrest.CoreMatchers.endsWith;
import static org.junit.Assert.assertThat;
import static org.springframework.http.MediaType.TEXT_EVENT_STREAM;

import java.util.List;
import java.util.stream.Collectors;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.http.MediaType;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.reactive.server.EntityExchangeResult;
import org.springframework.test.web.reactive.server.FluxExchangeResult;
import org.springframework.test.web.reactive.server.WebTestClient;

import com.idugalic.domain.blog.BlogPost;
import com.idugalic.domain.blog.BlogPostRepository;
import com.idugalic.domain.project.Project;
import com.idugalic.domain.project.ProjectRepository;
import com.idugalic.web.blog.BlogPostController;
import com.idugalic.web.project.ProjectController;

import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;


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
	
	@Test
	public void streamAllProjectsIntegrationTest() throws Exception {
		FluxExchangeResult<Project> result = this.webTestClient.get()
			.uri("/projects")
			.accept(TEXT_EVENT_STREAM)
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(TEXT_EVENT_STREAM)
			.returnResult(Project.class);

		StepVerifier.create(result.getResponseBody())
			.expectNext(expectedProjects.get(0), expectedProjects.get(1))
			.expectNextCount(1)
			.consumeNextWith(project -> assertThat(project.getName(), endsWith("4")))
			.thenCancel()
			.verify();
	}
	
	@Test
	public void listBlogPostsByTitleIntegrationTest() {
		this.webTestClient.get().uri("/blogposts/search/bytitle?title=title1")
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
			.expectBodyList(BlogPost.class).isEqualTo(expectedBlogPosts.stream().filter(bp->bp.getTitle().equals("title1")).collect(Collectors.toList()));
	}
	
	@Test
	public void listProjectsByNameIntegrationTest() {
		this.webTestClient.get().uri("/projects/search/byname?name=name1")
			.exchange()
			.expectStatus().isOk()
			.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
			.expectBodyList(Project.class).isEqualTo(expectedProjects.stream().filter(bp->bp.getName().equals("name1")).collect(Collectors.toList()));
	}
	
	@Test
	public void getBlogPostIntegrationTest() throws Exception {
		this.webTestClient.get().uri("/blogposts/"+expectedBlogPosts.get(0).getId())
				.exchange()
				.expectStatus().isOk()
				.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
				.expectBody(BlogPost.class).isEqualTo(expectedBlogPosts.get(0));
	}
	
	@Test
	public void getProjectIntegrationTest() throws Exception {
		this.webTestClient.get().uri("/projects/"+expectedProjects.get(0).getId())
				.exchange()
				.expectStatus().isOk()
				.expectHeader().contentType(MediaType.APPLICATION_JSON_UTF8)
				.expectBody(Project.class).isEqualTo(expectedProjects.get(0));
	}
	
	@Test
	public void postBlogPostIntegrationTest() throws Exception {
		this.webTestClient.post().uri("/blogposts")
		        .body(Mono.just(new BlogPost("authorId5", "title5", "content5", "tagString5")), BlogPost.class)
				.exchange()
				.expectStatus().isOk();
				//.expectBody().isEmpty();
	}
	
	@Test
	public void postProjectIntegrationTest() throws Exception {
		this.webTestClient.post().uri("/projects")
				.body(Mono.just(new Project("name5", "repoUrl5", "siteUrl5", "category5", "description5")), Project.class)
				.exchange()
				.expectStatus().isOk();
				//.expectBody().isEmpty();
	}

}
