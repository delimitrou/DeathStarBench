package com.idugalic;

import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.data.mongodb.config.EnableMongoAuditing;
import org.springframework.data.mongodb.core.CollectionOptions;
import org.springframework.data.mongodb.core.ReactiveMongoTemplate;
import org.springframework.data.mongodb.repository.config.EnableReactiveMongoRepositories;

import com.idugalic.domain.blog.BlogPost;
import com.idugalic.domain.blog.BlogPostRepository;
import com.idugalic.domain.project.Project;
import com.idugalic.domain.project.ProjectRepository;

@SpringBootApplication
@EnableMongoAuditing
@EnableReactiveMongoRepositories
//@EnableWebFlux
public class ReactiveComapnyApplication {

	public static void main(String[] args) {
		SpringApplication.run(ReactiveComapnyApplication.class, args);
	}
	
	@Bean
	CommandLineRunner initData(ReactiveMongoTemplate reactiveMongoTemplate, BlogPostRepository blogPostRepository, ProjectRepository projectRepository) {
		return (p) -> {
			reactiveMongoTemplate.dropCollection(BlogPost.class).then(reactiveMongoTemplate.createCollection(
					BlogPost.class, CollectionOptions.empty().capped(104857600).size(104857600))).block();
			
			//blogPostRepository.deleteAll().block();
			blogPostRepository.save(new BlogPost("authorId1", "title1", "content1", "tagString1")).block();
			blogPostRepository.save(new BlogPost("authorId2", "title2", "content2", "tagString2")).block();
			blogPostRepository.save(new BlogPost("authorId3", "title3", "content3", "tagString3")).block();
			blogPostRepository.save(new BlogPost("authorId4", "title4", "content4", "tagString4")).block();

			projectRepository.deleteAll().block();
			projectRepository.save(new Project("name1", "repoUrl1", "siteUrl1", "category1", "description1")).block();
			projectRepository.save(new Project("name2", "repoUrl2", "siteUrl2", "category2", "description2")).block();
			projectRepository.save(new Project("name3", "repoUrl3", "siteUrl3", "category3", "description3")).block();
			projectRepository.save(new Project("name4", "repoUrl4", "siteUrl4", "category4", "description4")).block();
		};
	}
}
