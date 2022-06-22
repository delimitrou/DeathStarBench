package com.idugalic.domain.project;

import org.springframework.data.repository.reactive.ReactiveSortingRepository;

import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

public interface ProjectRepository extends ReactiveSortingRepository<Project, String> {

	Flux<Project> findByName(Mono<String> name);

}
