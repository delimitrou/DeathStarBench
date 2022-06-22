package com.idugalic.domain.blog;

import org.springframework.data.mongodb.repository.Tailable;
import org.springframework.data.repository.reactive.ReactiveSortingRepository;

import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

public interface BlogPostRepository extends ReactiveSortingRepository<BlogPost, String> {

	Flux<BlogPost> findByTitle(Mono<String> title);
	
	/*
	 * By default, MongoDB will automatically close a cursor when the client has exhausted all results in the cursor. 
	 * However, for capped collections you may use a Tailable Cursor that remains open after the client exhausts the results in the initial cursor. 
	 * Tailable cursors are conceptually equivalent to the tail Unix command with the -f option (i.e. with “follow” mode). After clients insert new additional documents into a capped collection, the tailable cursor will continue to retrieve documents.
	 */
	@Tailable
	Flux<BlogPost> findBy();

}
